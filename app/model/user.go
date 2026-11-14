package model

import "time"

const (
	RolePemilik  = "pemilik"
	RolePenghuni = "penghuni"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`
	Email        string    `gorm:"type:varchar(100);unique;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Phone        string    `gorm:"type:varchar(20)" json:"phone"`
	Role         string    `gorm:"type:varchar(20);not null;default:'penghuni'" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	Rentals      []Rental  `gorm:"foreignKey:UserID" json:"rentals,omitempty"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
