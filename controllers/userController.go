package controllers

import (
	"context"
	"log"
	"net/http"
	"project_10/database"
	helper "project_10/helpers"
	"project_10/models"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Creating a database collection named "user"
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

// Initializing a new validator
var validate = validator.New()

// Function to hash a password
func HashPassword(providedPassword string) string {
	password, err := bcrypt.GenerateFromPassword([]byte(providedPassword), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(password)
}

// Function to verify if the userPassword is the same as foundUserPassword
func VerifyPassword(userPassword string, foundUserPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(foundUserPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "Email or password is incorrect"
		check = false
	}

	return check, msg
}

func Signup() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		defer cancel()

		if err := c.BodyParser(&user); err != nil {
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{
				"error": err.Error(),
			})
			return nil
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{
				"error": validationErr.Error(),
			})
			return nil
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})

		if count > 0 {
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": "this email already exists",
			})
			return nil
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		if err != nil {
			log.Panic(err)
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": "error occured while checking for the records",
			})
			return nil
		}

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone_number": user.Phone})

		if err != nil {
			log.Panic(err)
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": "error occured while checking for the phone number",
			})
			return nil
		}

		if count > 0 {
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": "this phone number already exists",
			})
			return nil
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		insertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := "User item was not created"
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": msg,
			})
			defer cancel()
			return nil
		}

		c.Status(http.StatusOK).JSON(insertionNumber)

		return nil
	}

}

func Login() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		var user models.User
		var foundUser models.User

		defer cancel()

		if err := c.BodyParser(&user); err != nil {
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{
				"error": err.Error(),
			})
			return nil
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()

		if err != nil {
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": "Email or Password is incorrect",
			})
			return nil
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)

		if !passwordIsValid {
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": msg,
			})
			return nil
		}

		if foundUser.Email == nil {
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": "user not found",
			})
			return nil
		}

		token, refreshToken, err := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": err.Error(),
			})
			return nil
		}

		c.Status(http.StatusOK).JSON(foundUser)

		return nil
	}
}

func GetUsers() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{
				"error": err.Error(),
			})
			return nil
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))

		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}

		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}},
			{"total_count", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push", "$$ROOT"}}},
		}}}

		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items",
					bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		defer cancel()

		if err != nil {
			c.Status(http.StatusInternalServerError).JSON(&fiber.Map{
				"error": "error occured while fetching users",
			})
			return nil
		}

		var allUsers []bson.M

		if err = result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}

		c.Status(http.StatusOK).JSON(allUsers[0])

		return nil
	}
}

func GetUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{
				"error": err.Error(),
			})
			return nil
		}

		userId := c.Params("user_id")

		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{
				"error": err.Error(),
			})
			return nil
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)

		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)

		defer cancel()

		if err != nil {
			c.Status(http.StatusBadRequest).JSON(&fiber.Map{
				"error": err.Error(),
			})

			return nil
		}

		c.Status(http.StatusOK).JSON(user)

		return nil
	}
}
