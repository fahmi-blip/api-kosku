package helper

import (
	"github.com/gofiber/fiber/v2"
)

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(WebResponse{
		Success: false,
		Message: message,
		Data:    nil,
	})
}

func Ok(c *fiber.Ctx, status int, message string, data any) error {
	return Success(c, status, message, data)
}

func Created(c *fiber.Ctx, message string, data any, location string) error {
	if location != "" {
		c.Set("Location", location)
	}
	return Success(c, fiber.StatusCreated, message, data)
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
