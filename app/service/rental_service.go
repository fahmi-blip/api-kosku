package service

import (
	"strconv"
	"time"

	"kosku-api/app/model"
	"kosku-api/app/repository"
	"kosku-api/helper"

	"github.com/gofiber/fiber/v2"
)

type RentalService interface {
	Apply(c *fiber.Ctx) error
	GetAll(c *fiber.Ctx) error
	GetLogs(c *fiber.Ctx) error
	Approve(c *fiber.Ctx) error
	Reject(c *fiber.Ctx) error
	Cancel(c *fiber.Ctx) error
}

type rentalService struct {
	rentalRepo repository.RentalRepository
	roomRepo   repository.RoomRepository
}

func NewRentalService(rentalRepo repository.RentalRepository, roomRepo repository.RoomRepository) RentalService {
	return &rentalService{
		rentalRepo: rentalRepo,
		roomRepo:   roomRepo,
	}
}

func getUserID(c *fiber.Ctx) uint {
	if user, ok := helper.CurrentUser(c); ok {
		return user.UserID
	}
	val := c.Locals("user_id")
	switch v := val.(type) {
	case uint:
		return v
	case float64:
		return uint(v)
	case int:
		return uint(v)
	}
	return 0
}

func getRole(c *fiber.Ctx) string {
	if user, ok := helper.CurrentUser(c); ok {
		return user.Role
	}
	val := c.Locals("role")
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func (s *rentalService) Apply(c *fiber.Ctx) error {
	role := getRole(c)
	if role != model.RolePenghuni {
		return helper.Error(c, fiber.StatusForbidden, "Hanya penghuni yang dapat mengajukan sewa")
	}

	userID := getUserID(c)
	if userID == 0 {
		return helper.Error(c, fiber.StatusUnauthorized, "Pengguna tidak terautentikasi")
	}

	var req model.RentalRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	if req.RoomID == 0 {
		return helper.Error(c, fiber.StatusBadRequest, "ID kamar wajib diisi")
	}

	if req.DurationMonths < 1 {
		return helper.Error(c, fiber.StatusBadRequest, "Durasi sewa minimal 1 bulan")
	}

	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Konfigurasi timezone tidak tersedia")
	}
	startDate, err := time.ParseInLocation("2006-01-02", req.StartDate, location)
	if err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "Format tanggal mulai harus YYYY-MM-DD")
	}

	today := time.Now().In(location).Truncate(24 * time.Hour)
	if startDate.Before(today) {
		return helper.Error(c, fiber.StatusBadRequest, "Tanggal mulai sewa tidak boleh di masa lalu")
	}

	room, err := s.roomRepo.FindByID(req.RoomID)
	if err != nil {
		return helper.Error(c, fiber.StatusNotFound, "Kamar tidak ditemukan")
	}

	if room.Status != model.RoomAvailable {
		return helper.Error(c, fiber.StatusBadRequest, "Kamar tidak tersedia untuk disewa")
	}

	hasPending, err := s.rentalRepo.ExistsPendingByUserAndRoom(userID, req.RoomID)
	if err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal memeriksa status pengajuan sewa")
	}
	if hasPending {
		return helper.Error(c, fiber.StatusConflict, "Anda sudah memiliki pengajuan sewa yang berstatus pending untuk kamar ini")
	}

	rental := model.Rental{
		UserID:         userID,
		RoomID:         req.RoomID,
		StartDate:      startDate,
		DurationMonths: req.DurationMonths,
		Status:         model.StatusPending,
	}

	if err := s.rentalRepo.CreateWithLog(&rental, &model.RentalStatusLog{
		ActorID:    userID,
		FromStatus: "",
		ToStatus:   model.StatusPending,
		Note:       "Pengajuan sewa dibuat",
	}); err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal mengajukan sewa kamar")
	}

	resp := model.RentalResponse{
		ID:             rental.ID,
		UserID:         rental.UserID,
		RoomID:         rental.RoomID,
		StartDate:      rental.StartDate.Format("2006-01-02"),
		DurationMonths: rental.DurationMonths,
		Status:         rental.Status,
		CreatedAt:      rental.CreatedAt,
	}

	return helper.Success(c, fiber.StatusCreated, "Pengajuan sewa berhasil dibuat", resp)
}

func (s *rentalService) GetAll(c *fiber.Ctx) error {
	role := getRole(c)
	userID := getUserID(c)

	var rentals []model.Rental
	var err error

	if role == model.RolePemilik {
		rentals, err = s.rentalRepo.FindAll()
	} else if role == model.RolePenghuni {
		rentals, err = s.rentalRepo.FindByUserID(userID)
	} else {
		return helper.Error(c, fiber.StatusForbidden, "Akses ditolak")
	}

	if err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data sewa")
	}

	var resp []model.RentalResponse
	for _, item := range rentals {
		rentalItem := model.RentalResponse{
			ID:             item.ID,
			UserID:         item.UserID,
			RoomID:         item.RoomID,
			StartDate:      item.StartDate.Format("2006-01-02"),
			DurationMonths: item.DurationMonths,
			Status:         item.Status,
			CreatedAt:      item.CreatedAt,
		}

		if item.User.ID != 0 {
			rentalItem.User = &model.UserResponse{
				ID:    item.User.ID,
				Name:  item.User.Name,
				Email: item.User.Email,
				Phone: item.User.Phone,
				Role:  item.User.Role,
			}
		}

		if item.Room.ID != 0 {
			rentalItem.Room = &model.RoomResponse{
				ID:            item.Room.ID,
				Code:          item.Room.Code,
				Name:          item.Room.Name,
				PricePerMonth: item.Room.PricePerMonth,
				Facilities:    item.Room.Facilities,
				Status:        item.Room.Status,
				CreatedAt:     item.Room.CreatedAt,
			}
		}

		resp = append(resp, rentalItem)
	}

	return helper.Success(c, fiber.StatusOK, "Berhasil mengambil data sewa", resp)
}

func (s *rentalService) GetLogs(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, fiber.StatusBadRequest, "ID sewa tidak valid")
	}

	rental, err := s.rentalRepo.FindByID(uint(id))
	if err != nil {
		return helper.Error(c, fiber.StatusNotFound, "Pengajuan sewa tidak ditemukan")
	}
	userID := getUserID(c)
	if getRole(c) == model.RolePenghuni && rental.UserID != userID {
		return helper.Error(c, fiber.StatusForbidden, "Anda tidak dapat melihat log sewa pengguna lain")
	}

	logs, err := s.rentalRepo.FindLogsByRentalID(uint(id))
	if err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal mengambil log sewa")
	}
	resp := make([]model.RentalLogResponse, 0, len(logs))
	for _, entry := range logs {
		item := model.RentalLogResponse{
			ID: entry.ID, RentalID: entry.RentalID, ActorID: entry.ActorID,
			FromStatus: entry.FromStatus, ToStatus: entry.ToStatus,
			Note: entry.Note, CreatedAt: entry.CreatedAt,
		}
		if entry.Actor.ID != 0 {
			item.ActorName = entry.Actor.Name
		}
		resp = append(resp, item)
	}
	return helper.Success(c, fiber.StatusOK, "Berhasil mengambil log sewa", resp)
}

func (s *rentalService) Approve(c *fiber.Ctx) error {
	role := getRole(c)
	if role != model.RolePemilik {
		return helper.Error(c, fiber.StatusForbidden, "Hanya pemilik yang dapat menyetujui pengajuan sewa")
	}

	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "ID sewa tidak valid")
	}

	rental, err := s.rentalRepo.FindByID(uint(id))
	if err != nil {
		return helper.Error(c, fiber.StatusNotFound, "Pengajuan sewa tidak ditemukan")
	}

	if rental.Status != model.StatusPending {
		return helper.Error(c, fiber.StatusBadRequest, "Pengajuan sewa harus berstatus pending")
	}

	room, err := s.roomRepo.FindByID(rental.RoomID)
	if err != nil {
		return helper.Error(c, fiber.StatusNotFound, "Kamar tidak ditemukan")
	}

	if room.Status != model.RoomAvailable {
		return helper.Error(c, fiber.StatusBadRequest, "Kamar sudah tidak berstatus available")
	}

	hasApproved, err := s.rentalRepo.ExistsApprovedByRoom(rental.RoomID)
	if err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal memeriksa status sewa kamar")
	}
	if hasApproved {
		return helper.Error(c, fiber.StatusConflict, "Kamar sudah memiliki sewa yang disetujui")
	}

	if err := s.rentalRepo.ApproveTx(rental.ID, rental.RoomID, getUserID(c)); err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal menyetujui sewa kamar")
	}

	return helper.Success(c, fiber.StatusOK, "Pengajuan sewa berhasil disetujui", nil)
}

func (s *rentalService) Reject(c *fiber.Ctx) error {
	role := getRole(c)
	if role != model.RolePemilik {
		return helper.Error(c, fiber.StatusForbidden, "Hanya pemilik yang dapat menolak pengajuan sewa")
	}

	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "ID sewa tidak valid")
	}

	rental, err := s.rentalRepo.FindByID(uint(id))
	if err != nil {
		return helper.Error(c, fiber.StatusNotFound, "Pengajuan sewa tidak ditemukan")
	}

	if rental.Status != model.StatusPending {
		return helper.Error(c, fiber.StatusBadRequest, "Pengajuan sewa harus berstatus pending")
	}

	if err := s.rentalRepo.RejectTx(rental.ID, getUserID(c)); err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal menolak pengajuan sewa")
	}

	return helper.Success(c, fiber.StatusOK, "Pengajuan sewa berhasil ditolak", nil)
}

func (s *rentalService) Cancel(c *fiber.Ctx) error {
	if getRole(c) != model.RolePenghuni {
		return helper.Error(c, fiber.StatusForbidden, "Hanya penghuni yang dapat membatalkan sewa")
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || id == 0 {
		return helper.Error(c, fiber.StatusBadRequest, "ID sewa tidak valid")
	}
	rental, err := s.rentalRepo.FindByID(uint(id))
	if err != nil {
		return helper.Error(c, fiber.StatusNotFound, "Pengajuan sewa tidak ditemukan")
	}
	if rental.UserID != getUserID(c) {
		return helper.Error(c, fiber.StatusForbidden, "Anda hanya dapat membatalkan sewa sendiri")
	}
	if rental.Status != model.StatusPending && rental.Status != model.StatusApproved {
		return helper.Error(c, fiber.StatusBadRequest, "Sewa hanya dapat dibatalkan saat pending atau approved")
	}
	if err := s.rentalRepo.CancelTx(rental.ID, rental.RoomID, getUserID(c)); err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal membatalkan sewa")
	}
	return helper.Success(c, fiber.StatusOK, "Sewa berhasil dibatalkan", nil)
}
