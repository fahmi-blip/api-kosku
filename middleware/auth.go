package middleware

import (
	"errors"
	"strings"
	"time"

	"kosku-api/helper"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RequireAuth(jwtManager *helper.JWTManager) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		token, err := bearerToken(c)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Unauthorized("header Authorization tidak ada atau salah bentuk")
		}

		authUser, err := jwtManager.Parse(token)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.Unauthorized("access token kedaluwarsa")
			}
			return helper.Unauthorized("access token tidak valid")
		}

		c.Locals(helper.LocalsAuthUser, authUser)
		c.Locals("user_id", authUser.UserID)
		c.Locals("role", authUser.Role)

		return c.Next()
	}
}

func bearerToken(c *fiber.Ctx) (string, error) {
	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		return "", errors.New("header kosong")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("format bukan Bearer")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("token kosong")
	}
	return token, nil
}


