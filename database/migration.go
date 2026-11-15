package database

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"kosku-api/app/model"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	log.Println("Menjalankan migrasi skema database...")
	err := db.AutoMigrate(
		&model.User{},
		&model.Room{},
		&model.Rental{},
		&model.RentalStatusLog{},
		&model.RefreshToken{},
	)
	if err != nil {
		return err
	}
	log.Println("Migrasi skema database selesai")
	return nil
}

func RunSQLMigrations(db *gorm.DB, dirPath string) error {
	if dirPath == "" {
		if _, err := os.Stat("database/migrations"); err == nil {
			dirPath = "database/migrations"
		} else {
			dirPath = "migrations"
		}
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`).Error; err != nil {
		return err
	}

	var upFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}

	sort.Strings(upFiles)

	for _, fileName := range upFiles {
		var applied bool
		if err := db.Raw("SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = ?)", fileName).Scan(&applied).Error; err != nil {
			return err
		}
		if applied {
			continue
		}

		filePath := filepath.Join(dirPath, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		log.Printf("Mengeksekusi migrasi file: %s", fileName)
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(content)).Error; err != nil {
				return err
			}
			return tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", fileName).Error
		}); err != nil {
			return err
		}
	}

	log.Println("Seluruh berkas migrasi SQL berhasil dieksekusi")
	return nil
}

func DropTables(db *gorm.DB) error {
	log.Println("Menghapus tabel database...")
	return db.Migrator().DropTable(
		&model.RentalStatusLog{},
		&model.RefreshToken{},
		&model.Rental{},
		&model.Room{},
		&model.User{},
	)
}
