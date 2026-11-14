package model

import "time"

const (
	StatusPending   = "pending"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusCancelled = "cancelled"
)

type Rental struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null" json:"user_id"`
	RoomID         uint      `gorm:"not null" json:"room_id"`
	StartDate      time.Time `gorm:"type:date;not null" json:"start_date"`
	DurationMonths int       `gorm:"not null" json:"duration_months"`
	Status         string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	CreatedAt      time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Room Room `gorm:"foreignKey:RoomID" json:"room,omitempty"`
}

type RentalRequest struct {
	RoomID         uint   `json:"room_id"`
	StartDate      string `json:"start_date"`
	DurationMonths int    `json:"duration_months"`
}

type RentalResponse struct {
	ID             uint          `json:"id"`
	UserID         uint          `json:"user_id"`
	RoomID         uint          `json:"room_id"`
	StartDate      string        `json:"start_date"`
	DurationMonths int           `json:"duration_months"`
	Status         string        `json:"status"`
	CreatedAt      time.Time     `json:"created_at"`
	User           *UserResponse `json:"user,omitempty"`
	Room           *RoomResponse `json:"room,omitempty"`
}