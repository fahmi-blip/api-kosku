package repository

import (
	"kosku-api/app/model"

	"gorm.io/gorm"
)

type RoomRepository interface {
	FindAll() ([]model.Room, error)
	FindAvailable() ([]model.Room, error)
	FindByID(id uint) (*model.Room, error)
	FindByCode(code string) (*model.Room, error)
	Create(room *model.Room) error
	Update(room *model.Room) error
	Delete(id uint) error
	UpdateStatus(id uint, status string) error
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) FindAll() ([]model.Room, error) {
	var rooms []model.Room
	err := r.db.Order("id desc").Find(&rooms).Error
	return rooms, err
}

func (r *roomRepository) FindAvailable() ([]model.Room, error) {
	var rooms []model.Room
	err := r.db.Where("status = ?", model.RoomAvailable).Order("id desc").Find(&rooms).Error
	return rooms, err
}

func (r *roomRepository) FindByID(id uint) (*model.Room, error) {
	var room model.Room
	err := r.db.First(&room, id).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) FindByCode(code string) (*model.Room, error) {
	var room model.Room
	err := r.db.Where("code = ?", code).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) Create(room *model.Room) error {
	return r.db.Create(room).Error
}

func (r *roomRepository) Update(room *model.Room) error {
	return r.db.Save(room).Error
}

func (r *roomRepository) Delete(id uint) error {
	return r.db.Delete(&model.Room{}, id).Error
}

func (r *roomRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Room{}).Where("id = ?", id).Update("status", status).Error
}
