package helper

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// Function to check if the user type is ADMIN, and if it is not, then the access for that user is not granted
func CheckUserType(c *fiber.Ctx, role string) (err error) {
	userType := c.Locals("user_type")
	fmt.Println(userType)
	err = nil
	if userType != role {
		err = errors.New("unauthorized to access this resource")
		return err
	}
	return err
}

// Function to check the user type as well as to match the user ID stored in the locals to the provided user id
func MatchUserTypeToUid(c *fiber.Ctx, userId string) (err error) {
	userType := c.Locals("user_type").(string)
	uid := c.Locals("user_id")

	if userType == "USER" && uid != userId {
		err = errors.New("unauthorized to access this resource")
		return err
	}

	err = CheckUserType(c, userType)
	return err
}
