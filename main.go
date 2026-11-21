package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kosku-api/app/repository"
	"kosku-api/app/service"
	"kosku-api/config"
	"kosku-api/database"
	"kosku-api/helper"
	"kosku-api/route"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek (minimal %d karakter)", minSecretLength)
		os.Exit(1)
	}

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("Gagal terhubung ke database: %v", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.RunSQLMigrations(context.Background(), pool, "migrations"); err != nil {
		logger.Error("Gagal menjalankan migrasi SQL: %v", err)
		os.Exit(1)
	}

	database.SeedOwner(context.Background(), pool)

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "kosku-api"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)
	refreshTTL := time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7)) * 24 * time.Hour

	perms := helper.DefaultPermissionSet()

	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	roomRepo := repository.NewRoomRepository(pool)
	rentalRepo := repository.NewRentalRepository(pool)

	authService := service.NewAuthService(userRepo, tokenRepo, jwtManager, refreshTTL)
	roomService := service.NewRoomService(roomRepo)
	rentalService := service.NewRentalService(rentalRepo, roomRepo)

	deps := route.Dependencies{
		JWT:           jwtManager,
		Permissions:   perms,
		AuthService:   authService,
		RoomService:   roomService,
		RentalService: rentalService,
	}

	app := config.NewApp(logger, deps)

	port := config.GetEnv("APP_PORT", "8080")
	addr := fmt.Sprintf(":%s", port)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("Server KosKu berjalan pada port %s", port)
		if err := app.Listen(addr); err != nil {
			logger.Error("Server berhenti: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Info("Menerima sinyal berhenti, mematikan server...")
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		logger.Error("Shutdown server tidak rapi: %v", err)
	}
	logger.Info("Server berhasil dimatikan dengan rapi")
}
