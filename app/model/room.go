package model

import "time"

const (
	RoomAvailable   = "available"
	RoomOccupied    = "occupied"
	RoomMaintenance = "maintenance"
)

type Room struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Code          string    `gorm:"type:varchar(50);unique;not null" json:"code"`
	Name          string    `gorm:"type:varchar(100);not null" json:"name"`
	PricePerMonth float64   `gorm:"type:decimal(12,2);not null" json:"price_per_month"`
	Facilities    string    `gorm:"type:text" json:"facilities"`
	Status        string    `gorm:"type:varchar(20);not null;default:'available'" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	Rentals       []Rental  `gorm:"foreignKey:RoomID" json:"rentals,omitempty"`
}