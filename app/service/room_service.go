package service

import (
	"strconv"

	"kosku-api/app/model"
	"kosku-api/app/repository"
	"kosku-api/helper"

	"github.com/gofiber/fiber/v2"
)

type RoomService interface {
	GetAll(c *fiber.Ctx) error
	GetAvailable(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type roomService struct {
	roomRepo repository.RoomRepository
}

func NewRoomService(roomRepo repository.RoomRepository) RoomService {
	return &roomService{roomRepo: roomRepo}
}

func (s *roomService) GetAll(c *fiber.Ctx) error {
	ctx := c.UserContext()
	rooms, err := s.roomRepo.FindAll(ctx)
	if err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data kamar")
	}

	var resp []model.RoomResponse
	for _, r := range rooms {
		resp = append(resp, model.RoomResponse{
			ID:            r.ID,
			Code:          r.Code,
			Name:          r.Name,
			PricePerMonth: r.PricePerMonth,
			Facilities:    r.Facilities,
			Status:        r.Status,
			CreatedAt:     r.CreatedAt,
		})
	}

	return helper.Success(c, fiber.StatusOK, "Berhasil mengambil data kamar", resp)
}

func (s *roomService) GetByID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "ID kamar tidak valid")
	}

	room, err := s.roomRepo.FindByID(ctx, uint(id))
	if err != nil || room == nil {
		return helper.Error(c, fiber.StatusNotFound, "Kamar tidak ditemukan")
	}

	resp := model.RoomResponse{
		ID:            room.ID,
		Code:          room.Code,
		Name:          room.Name,
		PricePerMonth: room.PricePerMonth,
		Facilities:    room.Facilities,
		Status:        room.Status,
		CreatedAt:     room.CreatedAt,
	}

	return helper.Success(c, fiber.StatusOK, "Berhasil mengambil detail kamar", resp)
}

func (s *roomService) GetAvailable(c *fiber.Ctx) error {
	ctx := c.UserContext()
	rooms, err := s.roomRepo.FindAvailable(ctx)
	if err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data kamar yang tersedia")
	}

	var resp []model.RoomResponse
	for _, r := range rooms {
		resp = append(resp, model.RoomResponse{
			ID:            r.ID,
			Code:          r.Code,
			Name:          r.Name,
			PricePerMonth: r.PricePerMonth,
			Facilities:    r.Facilities,
			Status:        r.Status,
			CreatedAt:     r.CreatedAt,
		})
	}

	return helper.Success(c, fiber.StatusOK, "Berhasil mengambil data kamar tersedia", resp)
}

func (s *roomService) Create(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var req model.RoomRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	if req.Code == "" {
		return helper.Error(c, fiber.StatusBadRequest, "Kode kamar wajib diisi")
	}

	if req.Name == "" {
		return helper.Error(c, fiber.StatusBadRequest, "Nama kamar wajib diisi")
	}

	if req.PricePerMonth <= 0 {
		return helper.Error(c, fiber.StatusBadRequest, "Harga sewa per bulan harus lebih dari 0")
	}

	existing, _ := s.roomRepo.FindByCode(ctx, req.Code)
	if existing != nil {
		return helper.Error(c, fiber.StatusConflict, "Kode kamar sudah digunakan")
	}

	status := req.Status
	if status == "" {
		status = model.RoomAvailable
	}
	if status != model.RoomAvailable && status != model.RoomOccupied && status != model.RoomMaintenance {
		return helper.Error(c, fiber.StatusBadRequest, "Status kamar tidak valid")
	}

	room := model.Room{
		Code:          req.Code,
		Name:          req.Name,
		PricePerMonth: req.PricePerMonth,
		Facilities:    req.Facilities,
		Status:        status,
	}

	if err := s.roomRepo.Create(ctx, &room); err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal membuat kamar baru")
	}

	resp := model.RoomResponse{
		ID:            room.ID,
		Code:          room.Code,
		Name:          room.Name,
		PricePerMonth: room.PricePerMonth,
		Facilities:    room.Facilities,
		Status:        room.Status,
		CreatedAt:     room.CreatedAt,
	}

	return helper.Success(c, fiber.StatusCreated, "Kamar berhasil dibuat", resp)
}

func (s *roomService) Update(c *fiber.Ctx) error {
	ctx := c.UserContext()
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "ID kamar tidak valid")
	}

	room, err := s.roomRepo.FindByID(ctx, uint(id))
	if err != nil || room == nil {
		return helper.Error(c, fiber.StatusNotFound, "Kamar tidak ditemukan")
	}

	var req model.RoomRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	if req.Code == "" {
		return helper.Error(c, fiber.StatusBadRequest, "Kode kamar tidak boleh kosong")
	}

	if req.Name == "" {
		return helper.Error(c, fiber.StatusBadRequest, "Nama kamar tidak boleh kosong")
	}

	if req.PricePerMonth <= 0 {
		return helper.Error(c, fiber.StatusBadRequest, "Harga sewa per bulan harus lebih dari 0")
	}

	if req.Code != room.Code {
		existing, _ := s.roomRepo.FindByCode(ctx, req.Code)
		if existing != nil && existing.ID != room.ID {
			return helper.Error(c, fiber.StatusConflict, "Kode kamar sudah digunakan")
		}
	}

	status := req.Status
	if status == "" {
		status = room.Status
	}
	if status != model.RoomAvailable && status != model.RoomOccupied && status != model.RoomMaintenance {
		return helper.Error(c, fiber.StatusBadRequest, "Status kamar tidak valid")
	}

	room.Code = req.Code
	room.Name = req.Name
	room.PricePerMonth = req.PricePerMonth
	room.Facilities = req.Facilities
	room.Status = status

	if err := s.roomRepo.Update(ctx, room); err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal memperbarui kamar")
	}

	resp := model.RoomResponse{
		ID:            room.ID,
		Code:          room.Code,
		Name:          room.Name,
		PricePerMonth: room.PricePerMonth,
		Facilities:    room.Facilities,
		Status:        room.Status,
		CreatedAt:     room.CreatedAt,
	}

	return helper.Success(c, fiber.StatusOK, "Kamar berhasil diperbarui", resp)
}

func (s *roomService) Delete(c *fiber.Ctx) error {
	ctx := c.UserContext()
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "ID kamar tidak valid")
	}

	room, err := s.roomRepo.FindByID(ctx, uint(id))
	if err != nil || room == nil {
		return helper.Error(c, fiber.StatusNotFound, "Kamar tidak ditemukan")
	}

	if room.Status == model.RoomOccupied {
		return helper.Error(c, fiber.StatusBadRequest, "Kamar berstatus occupied tidak boleh dihapus")
	}

	if err := s.roomRepo.Delete(ctx, uint(id)); err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal menghapus kamar")
	}

	return helper.Success(c, fiber.StatusOK, "Kamar berhasil dihapus", nil)
}
