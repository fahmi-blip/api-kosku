package database

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunSQLMigrations(ctx context.Context, pool *pgxpool.Pool, dirPath string) error {
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

	var upFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}

	sort.Strings(upFiles)

	for _, fileName := range upFiles {
		filePath := filepath.Join(dirPath, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		log.Printf("Mengeksekusi migrasi file: %s", fileName)
		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return err
		}
	}

	log.Println("Seluruh berkas migrasi SQL berhasil dieksekusi")
	return nil
}

func DropTables(ctx context.Context, pool *pgxpool.Pool) error {
	log.Println("Menghapus tabel database...")
	query := "DROP TABLE IF EXISTS rental_status_logs, refresh_tokens, rentals, rooms, users CASCADE"
	_, err := pool.Exec(ctx, query)
	return err
}
