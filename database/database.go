package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"kosku-api/app/model"
	"kosku-api/config"
	"kosku-api/helper"
)

func SeedOwner(ctx context.Context, pool *pgxpool.Pool) {
	ownerEmail := config.GetEnv("SEED_OWNER_EMAIL", "pemilik@kosku.com")
	ownerPassword := config.GetEnv("SEED_OWNER_PASSWORD", "pemilik12345")

	var count int64
	err := pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM users WHERE email = $1 OR role = $2",
		ownerEmail,
		model.RolePemilik,
	).Scan(&count)
	if err != nil {
		log.Printf("Gagal memeriksa akun pemilik: %v", err)
		return
	}
	if count > 0 {
		return
	}

	hash, err := helper.HashPassword(ownerPassword)
	if err != nil {
		log.Printf("Gagal hash password seed pemilik: %v", err)
		return
	}

	query := `
		INSERT INTO users (name, email, password_hash, phone, role)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = pool.Exec(ctx, query, "Pemilik Kos", ownerEmail, hash, "081234567890", model.RolePemilik)
	if err != nil {
		log.Printf("Gagal membuat akun seed pemilik: %v", err)
		return
	}

	log.Printf("Akun seed pemilik berhasil dibuat: %s", ownerEmail)
}
