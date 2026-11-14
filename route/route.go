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