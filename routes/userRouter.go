package routes

import (
	"project_10/controllers"

	"project_10/middleware"

	"github.com/gofiber/fiber/v2"
)

// Creating routes with authentication for getting users
func UserRoutes(incomingRoutes *fiber.App) {

	//Using the authenticate middleware to make sure the user has a jwt token
	incomingRoutes.Use(middleware.Authenticate())

	//only allowing to call the get function if the user has a valid jwt token
	incomingRoutes.Get("/users", controllers.GetUsers())
	incomingRoutes.Get("/users/:user_id", controllers.GetUser())
}
