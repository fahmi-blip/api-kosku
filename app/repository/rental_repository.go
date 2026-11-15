package repository

import (
	"kosku-api/app/model"

	"gorm.io/gorm"
)

type RentalRepository interface {
	Create(rental *model.Rental) error
	CreateWithLog(rental *model.Rental, log *model.RentalStatusLog) error
	FindByID(id uint) (*model.Rental, error)
	FindAll() ([]model.Rental, error)
	FindByUserID(userID uint) ([]model.Rental, error)
	FindLogsByRentalID(rentalID uint) ([]model.RentalStatusLog, error)
	ExistsApprovedByRoom(roomID uint) (bool, error)
	ExistsPendingByUserAndRoom(userID, roomID uint) (bool, error)
	ApproveTx(rentalID, roomID, actorID uint) error
	RejectTx(rentalID, actorID uint) error
	CancelTx(rentalID, roomID, actorID uint) error
}

type rentalRepository struct {
	db *gorm.DB
}

func NewRentalRepository(db *gorm.DB) RentalRepository {
	return &rentalRepository{db: db}
}

func (r *rentalRepository) Create(rental *model.Rental) error {
	return r.db.Create(rental).Error
}

func (r *rentalRepository) CreateWithLog(rental *model.Rental, logEntry *model.RentalStatusLog) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rental).Error; err != nil {
			return err
		}
		logEntry.RentalID = rental.ID
		return tx.Create(logEntry).Error
	})
}

func (r *rentalRepository) FindByID(id uint) (*model.Rental, error) {
	var rental model.Rental
	err := r.db.Preload("User").Preload("Room").First(&rental, id).Error
	if err != nil {
		return nil, err
	}
	return &rental, nil
}

func (r *rentalRepository) FindAll() ([]model.Rental, error) {
	var rentals []model.Rental
	err := r.db.Preload("User").Preload("Room").Order("id desc").Find(&rentals).Error
	return rentals, err
}

func (r *rentalRepository) FindByUserID(userID uint) ([]model.Rental, error) {
	var rentals []model.Rental
	err := r.db.Preload("User").Preload("Room").Where("user_id = ?", userID).Order("id desc").Find(&rentals).Error
	return rentals, err
}

func (r *rentalRepository) FindLogsByRentalID(rentalID uint) ([]model.RentalStatusLog, error) {
	var logs []model.RentalStatusLog
	err := r.db.Preload("Actor").Where("rental_id = ?", rentalID).Order("id asc").Find(&logs).Error
	return logs, err
}

func (r *rentalRepository) ExistsApprovedByRoom(roomID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Rental{}).
		Where("room_id = ? AND status = ?", roomID, model.StatusApproved).
		Count(&count).Error
	return count > 0, err
}

func (r *rentalRepository) ExistsPendingByUserAndRoom(userID, roomID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Rental{}).
		Where("user_id = ? AND room_id = ? AND status = ?", userID, roomID, model.StatusPending).
		Count(&count).Error
	return count > 0, err
}

func (r *rentalRepository) ApproveTx(rentalID, roomID, actorID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.Rental{}).
			Where("id = ? AND room_id = ? AND status = ?", rentalID, roomID, model.StatusPending).
			Update("status", model.StatusApproved)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		roomResult := tx.Model(&model.Room{}).
			Where("id = ? AND status = ?", roomID, model.RoomAvailable).
			Update("status", model.RoomOccupied)
		if roomResult.Error != nil {
			return roomResult.Error
		}
		if roomResult.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.Create(&model.RentalStatusLog{
			RentalID:   rentalID,
			ActorID:    actorID,
			FromStatus: model.StatusPending,
			ToStatus:   model.StatusApproved,
			Note:       "Pengajuan disetujui",
		}).Error
	})
}

func (r *rentalRepository) RejectTx(rentalID, actorID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.Rental{}).
			Where("id = ? AND status = ?", rentalID, model.StatusPending).
			Update("status", model.StatusRejected)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.Create(&model.RentalStatusLog{
			RentalID:   rentalID,
			ActorID:    actorID,
			FromStatus: model.StatusPending,
			ToStatus:   model.StatusRejected,
			Note:       "Pengajuan ditolak",
		}).Error
	})
}

func (r *rentalRepository) CancelTx(rentalID, roomID, actorID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var rental model.Rental
		if err := tx.First(&rental, rentalID).Error; err != nil {
			return err
		}
		if rental.Status != model.StatusPending && rental.Status != model.StatusApproved {
			return gorm.ErrInvalidData
		}
		if result := tx.Model(&model.Rental{}).
			Where("id = ? AND room_id = ? AND status = ?", rentalID, roomID, rental.Status).
			Update("status", model.StatusCancelled); result.Error != nil {
			return result.Error
		} else if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		if rental.Status == model.StatusApproved {
			roomResult := tx.Model(&model.Room{}).
				Where("id = ? AND status = ?", roomID, model.RoomOccupied).
				Update("status", model.RoomAvailable)
			if roomResult.Error != nil {
				return roomResult.Error
			}
			if roomResult.RowsAffected != 1 {
				return gorm.ErrRecordNotFound
			}
		}
		return tx.Create(&model.RentalStatusLog{
			RentalID:   rentalID,
			ActorID:    actorID,
			FromStatus: rental.Status,
			ToStatus:   model.StatusCancelled,
			Note:       "Sewa dibatalkan oleh penghuni",
		}).Error
	})
}
