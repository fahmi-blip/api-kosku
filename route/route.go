package route

import (
	"kosku-api/app/service"
	"kosku-api/helper"
	"kosku-api/middleware"

	"github.com/gofiber/fiber/v2"
)

type Dependencies struct {
	JWT           *helper.JWTManager
	Permissions   *helper.PermissionSet
	AuthService   service.AuthService
	RoomService   service.RoomService
	RentalService service.RentalService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return helper.Ok(c, fiber.StatusOK, "server berjalan", nil)
	})

	// --- Auth ---
	auth := api.Group("/auth")
	auth.Post("/register", middleware.RequireJSON, deps.AuthService.Register)
	auth.Post("/login", middleware.RequireJSON, middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", middleware.RequireJSON, deps.AuthService.Refresh)
	auth.Post("/logout", middleware.RequireAuth(deps.JWT), deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// --- Rooms ---
	rooms := api.Group("/rooms")
	rooms.Get("/", deps.RoomService.GetAll)
	rooms.Get("/available", deps.RoomService.GetAvailable)
	rooms.Get("/:id", middleware.RequireAuth(deps.JWT), middleware.RequirePermission(deps.Permissions, "room:read"), deps.RoomService.GetByID)
	rooms.Post("/", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), middleware.RequirePermission(deps.Permissions, "room:create"), deps.RoomService.Create)
	rooms.Put("/:id", middleware.RequireJSON, middleware.RequireAuth(deps.JWT), middleware.RequirePermission(deps.Permissions, "room:update"), deps.RoomService.Update)
	rooms.Delete("/:id", middleware.RequireAuth(deps.JWT), middleware.RequirePermission(deps.Permissions, "room:delete"), deps.RoomService.Delete)
}
