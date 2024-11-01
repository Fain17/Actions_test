package middleware

import (
	"net/http"
	helper "project_10/helpers"

	"github.com/gofiber/fiber/v2"
)

func Authenticate() fiber.Handler {
	return func(c *fiber.Ctx) error {

		//Using the c.Get() function to get the token from the URL header
		clientToken := c.Get("token")

		//Checking if token is present or not
		if clientToken == "" {
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": "Token was not found",
			})
			return nil
		}

		//Validation of the token using the tokenHelper function
		claims, err := helper.ValidateToken(clientToken)

		if err != "" {
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": err,
			})
			return nil
		}

		//Setting the value of the claims interface using the Locals value
		c.Locals("email", claims.Email)
		c.Locals("first_name", claims.First_name)
		c.Locals("last_name", claims.Last_name)
		c.Locals("user_id", claims.Uid)
		c.Locals("user_type", claims.User_type)

		//After the value has been set, the Next function can called using c.Next()
		return c.Next()
	}
}
