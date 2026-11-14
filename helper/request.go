package helper

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func ReqCtx(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func ParamID(c *fiber.Ctx) (uint, bool) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}
