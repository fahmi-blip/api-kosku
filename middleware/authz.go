package middleware

import (
	"kosku-api/helper"

	"github.com/gofiber/fiber/v2"
)

func RequirePermission(perms *helper.PermissionSet, permission string) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}
		if !perms.Can(user.Role, permission) {
			return helper.Forbidden("role " + user.Role + " tidak memiliki hak akses " + permission)
		}
		return c.Next()
	}
}

func RequireRole(roles ...string) func(*fiber.Ctx) error {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Unauthorized("belum terautentikasi")
		}

		if _, granted := allowed[user.Role]; !granted {
			return helper.Forbidden("role anda tidak berhak mengakses endpoint ini")
		}

		return c.Next()
	}
}