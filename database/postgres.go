package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"kosku-api/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func NewPostgresDB() *gorm.DB {
	dbHost := config.GetEnv("DB_HOST", "localhost")
	dbPort := config.GetEnv("DB_PORT", "5432")
	dbUser := config.GetEnv("DB_USER", "postgres")
	dbPassword := config.GetEnv("DB_PASSWORD", "postgres")
	dbName := config.GetEnv("DB_NAME", "kosku_db")
	sslMode := config.GetEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		dbHost,
		dbUser,
		dbPassword,
		dbName,
		dbPort,
		sslMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Gagal mendapatkan sql.DB: %v", err)
	}

	sqlDB.SetMaxOpenConns(config.GetEnvInt("DB_MAX_CONNS", 10))
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		sqlDB.Close()
		log.Fatalf("Gagal ping database PostgreSQL: %v", err)
	}

	if err := RunSQLMigrations(db, "migrations"); err != nil {
		log.Fatalf("Gagal menjalankan migrasi SQL: %v", err)
	}

	if err := Migrate(db); err != nil {
		log.Fatalf("Gagal migrasi database: %v", err)
	}

	seedOwner(db)

	return db
}
