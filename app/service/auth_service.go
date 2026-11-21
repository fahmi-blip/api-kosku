package service

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"kosku-api/app/model"
	"kosku-api/app/repository"
	"kosku-api/helper"

	"github.com/gofiber/fiber/v2"
)

type AuthService interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Refresh(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	Me(c *fiber.Ctx) error
}

type authService struct {
	userRepo   repository.UserRepository
	tokenRepo  repository.TokenRepository
	jwtManager *helper.JWTManager
	refreshTTL time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwtManager *helper.JWTManager,
	refreshTTL time.Duration,
) AuthService {
	return &authService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
		refreshTTL: refreshTTL,
	}
}

func (s *authService) Register(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)

	if req.Name == "" {
		return helper.Error(c, fiber.StatusBadRequest, "Nama wajib diisi")
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "Format email tidak valid")
	}

	if len(req.Password) < 8 {
		return helper.Error(c, fiber.StatusBadRequest, "Password minimal 8 karakter")
	}

	existingUser, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		helper.VerifyDummyPassword(req.Password)
		return helper.Error(c, fiber.StatusConflict, "Email sudah terdaftar")
	}

	hash, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal memproses password")
	}

	user := model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		Phone:        req.Phone,
		Role:         model.RolePenghuni,
	}

	if err := s.userRepo.Create(ctx, &user); err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal mendaftarkan pengguna")
	}

	resp := model.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
		Role:  user.Role,
	}

	return helper.Success(c, fiber.StatusCreated, "Registrasi berhasil", resp)
}

func (s *authService) Login(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		return helper.Error(c, fiber.StatusBadRequest, "Email dan password wajib diisi")
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil || user == nil {
		helper.VerifyDummyPassword(req.Password)
		return helper.Error(c, fiber.StatusUnauthorized, "Email atau password salah")
	}

	if !helper.CheckPassword(req.Password, user.PasswordHash) {
		return helper.Error(c, fiber.StatusUnauthorized, "Email atau password salah")
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return helper.Internal(err)
	}

	resp := model.AuthResponse{
		Token: *pair,
		User: model.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Phone: user.Phone,
			Role:  user.Role,
		},
	}

	return helper.Success(c, fiber.StatusOK, "Login berhasil", resp)
}

func (s *authService) Refresh(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		return helper.Error(c, fiber.StatusBadRequest, "refresh_token wajib diisi")
	}

	hash := helper.SHA256Hex(strings.TrimSpace(req.RefreshToken))

	stored, err := s.tokenRepo.FindByHash(ctx, hash)
	if err != nil || stored == nil {
		return helper.Error(c, fiber.StatusUnauthorized, "Refresh token tidak valid")
	}

	if stored.RevokedAt != nil {
		_ = s.tokenRepo.RevokeAllByUser(ctx, stored.UserID)
		return helper.Error(c, fiber.StatusUnauthorized, "Token sudah dicabut — semua sesi dihentikan")
	}

	if stored.ExpiresAt.Before(time.Now()) {
		return helper.Error(c, fiber.StatusUnauthorized, "Refresh token kedaluwarsa")
	}

	_ = s.tokenRepo.RevokeByHash(ctx, hash)

	user, err := s.userRepo.FindByID(ctx, stored.UserID)
	if err != nil || user == nil {
		return helper.Error(c, fiber.StatusUnauthorized, "Pengguna tidak ditemukan")
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal membuat token baru")
	}

	return helper.Success(c, fiber.StatusOK, "Token berhasil diperbarui", pair)
}

func (s *authService) Logout(c *fiber.Ctx) error {
	ctx := c.UserContext()

	authUser, ok := helper.CurrentUser(c)
	if !ok || authUser.UserID == 0 {
		return helper.Error(c, fiber.StatusUnauthorized, "Pengguna tidak terautentikasi")
	}

	if err := s.tokenRepo.RevokeAllByUser(ctx, authUser.UserID); err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal mengakhiri sesi pengguna")
	}

	return helper.Success(c, fiber.StatusOK, "Logout berhasil", nil)
}

func (s *authService) Me(c *fiber.Ctx) error {
	ctx := c.UserContext()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Error(c, fiber.StatusUnauthorized, "Pengguna tidak terautentikasi")
	}

	user, err := s.userRepo.FindByID(ctx, authUser.UserID)
	if err != nil || user == nil {
		return helper.Error(c, fiber.StatusNotFound, "Data pengguna tidak ditemukan")
	}

	resp := model.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
		Role:  user.Role,
	}

	return helper.Success(c, fiber.StatusOK, "Berhasil mengambil profil pengguna", resp)
}

func (s *authService) issueTokenPair(ctx context.Context, user *model.User) (*model.TokenPair, error) {
	access, err := s.jwtManager.GenerateAccess(*user)
	if err != nil {
		return nil, err
	}

	raw, err := helper.RandomToken(32)
	if err != nil {
		return nil, err
	}

	rt := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: helper.SHA256Hex(raw),
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}
	if err := s.tokenRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwtManager.AccessTTL().Seconds()),
	}, nil
}
