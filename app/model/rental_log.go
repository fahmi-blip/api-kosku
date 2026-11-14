package model

import "time"

type RentalStatusLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RentalID   uint      `gorm:"not null;index" json:"rental_id"`
	ActorID    uint      `gorm:"not null;index" json:"actor_id"`
	FromStatus string    `gorm:"type:varchar(20);not null" json:"from_status"`
	ToStatus   string    `gorm:"type:varchar(20);not null" json:"to_status"`
	Note       string    `gorm:"type:text" json:"note"`
	CreatedAt  time.Time `json:"created_at"`

	Actor  User   `gorm:"foreignKey:ActorID" json:"actor,omitempty"`
	Rental Rental `gorm:"foreignKey:RentalID" json:"-"`
}