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



func (s *roomService) GetAll(c *fiber.Ctx) error {
	rooms, err := s.roomRepo.FindAll()
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
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "ID kamar tidak valid")
	}

	room, err := s.roomRepo.FindByID(uint(id))
	if err != nil {
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