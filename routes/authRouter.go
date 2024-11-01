package routes

import (
	"project_10/controllers"

	"github.com/gofiber/fiber/v2"
)

// Creating auth routes for signup and login
func AuthRoutes(incomingRoutes *fiber.App) {
	incomingRoutes.Post("users/signup", controllers.Signup())
	incomingRoutes.Post("users/login", controllers.Login())
}
