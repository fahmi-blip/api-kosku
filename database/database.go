package database

import (
	"log"

	"kosku-api/app/model"
	"kosku-api/config"
	"kosku-api/helper"

	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	return NewPostgresDB()
}

func seedOwner(db *gorm.DB) {
	ownerEmail := config.GetEnv("SEED_OWNER_EMAIL", "pemilik@kosku.com")
	ownerPassword := config.GetEnv("SEED_OWNER_PASSWORD", "pemilik12345")

	var count int64
	db.Model(&model.User{}).Where("email = ? OR role = ?", ownerEmail, model.RolePemilik).Count(&count)
	if count > 0 {
		return
	}

	hash, err := helper.HashPassword(ownerPassword)
	if err != nil {
		log.Printf("Gagal hash password seed pemilik: %v", err)
		return
	}

	owner := model.User{
		Name:         "Pemilik Kos",
		Email:        ownerEmail,
		PasswordHash: hash,
		Phone:        "081234567890",
		Role:         model.RolePemilik,
	}

	if err := db.Create(&owner).Error; err != nil {
		log.Printf("Gagal membuat akun seed pemilik: %v", err)
		return
	}

	log.Printf("Akun seed pemilik berhasil dibuat: %s", ownerEmail)
}
