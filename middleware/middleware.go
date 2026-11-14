package middleware

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"kosku-api/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func Register(app *fiber.App, logger Logger, allowedOrigins string) {
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(corsPolicy(allowedOrigins))
	if logger != nil {
		app.Use(RequestLogger(logger))
	}
}

func corsPolicy(allowedOrigins string) func(*fiber.Ctx) error {
	if strings.TrimSpace(allowedOrigins) == "" {
		allowedOrigins = "*"
	}
	return cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	})
}