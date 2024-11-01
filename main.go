package main

import (
	"log"
	"net/http"
	"os"

	routes "project_10/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {

	//Getting port for hosting the API
	port := os.Getenv("PORT")

	//Check if port is nil
	if port == "" {
		port = "8000"
	}

	//Create a new fiber app
	app := fiber.New()

	//Initiate logger to use the fiber app
	app.Use(logger.New())

	//Call the routes using the routes handler
	routes.AuthRoutes(app)
	routes.UserRoutes(app)

	app.Get("/api-1", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(&fiber.Map{
			"success": "Access granted for api-1",
		})
	})

	app.Get("/api-2", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(&fiber.Map{
			"Success": "Access granted for api-2",
		})
	})

	//Test

	//Run the app on the specified port
	log.Fatal(app.Listen(":" + port))
}
