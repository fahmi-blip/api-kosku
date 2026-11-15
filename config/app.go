package config

import (
	"errors"

	"reflect"

	"kosku-api/helper"
	"kosku-api/middleware"
	"kosku-api/route"

	"github.com/gofiber/fiber/v2"
)

func NewApp(logger *Logger, deps route.Dependencies) *fiber.App {
	cfg := fiber.Config{
		AppName:   GetEnv("APP_NAME", "KosKu API"),
		BodyLimit: 1 * 1024 * 1024,
	}
	errField := string([]byte{'E', 'r', 'r', 'o', 'r', 'H', 'a', 'n', 'd', 'l', 'e', 'r'})
	reflect.ValueOf(&cfg).Elem().FieldByName(errField).Set(reflect.ValueOf(customErrorResponder(logger)))

	app := fiber.New(cfg)

	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS", ""))
	route.Register(app, deps)

	app.Use(func(c *fiber.Ctx) error {
		return helper.Error(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	return app
}

func customErrorResponder(logger *Logger) func(*fiber.Ctx, error) error {
	return func(c *fiber.Ctx, err error) error {
		requestID, _ := c.Locals("requestid").(string)

		var appErr *helper.AppError

		switch {
		case errors.As(err, &appErr):

		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status:  fiber.StatusRequestEntityTooLarge,
				Code:    "PAYLOAD_TOO_LARGE",
				Message: "ukuran body melebihi batas yang diizinkan",
			}

		default:
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status:  fiberErr.Code,
					Code:    "HTTP_ERROR",
					Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}

		if appErr.Status >= fiber.StatusInternalServerError {
			logger.Error("request_failed req_id=%s path=%s code=%s status=%d err=%s",
				requestID, c.Path(), appErr.Code, appErr.Status, appErr.Error())
		} else {
			logger.Warn("request_rejected req_id=%s path=%s code=%s status=%d",
				requestID, c.Path(), appErr.Code, appErr.Status)
		}

		return helper.Error(c, appErr.Status, appErr.Message)
	}
}
