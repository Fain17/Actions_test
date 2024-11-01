package helper

import (
	"context"
	"log"
	"os"
	"project_10/database"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Creating a SignedDetails struct
type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	User_type  string
	Uid        string
	jwt.StandardClaims
}

// Opening the database collection and storing it in userCollection
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

// Getting the secret key using the Getenv function, this is for creating the jwt
var SECRET_KEY string = os.Getenv("SECRET_KEY")

// Creating a GenerateAllTokens along with the jwt.StandardClaims and the signedToken, signedRefreshToken, and error is given as the output
func GenerateAllTokens(email string, firstName string, lastName string, userType string, userId string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		User_type:  userType,
		Uid:        userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	// Another refresh token with a longer duration is also created using the StandardClaims
	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	//This function takes in the signing method as well as the generated claims, and using the SECRET_KEY the token is being generated
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	//Using the same function NewWithClaims the refresh token claims are passed along with the SECRET_KEY to generate the refresh token
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

// This function is created to update the tokens regularly
func UpdateAllTokens(signedToken string, signedRefreshToken string, userID string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	//Creation of a mongoDB object of type primitive.D
	var updateObj primitive.D

	//updateObj using bson.E for single records
	updateObj = append(updateObj, bson.E{"token", signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", Updated_at})

	upsert := true

	filter := bson.M{"user_id": userID}

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	//Update the set fields
	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)

	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}
}

// Function created for validating a token
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	//Passing the signed token as well as signedDetails struct, and a function that returns the byte form of the secret key, it also stored the details in the SignedDetails struct
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	//It is a type assertion to check if Claims can store the token into SignedDetails, if yes ok comes as true else false, and claims contains the claims
	claims, ok := token.Claims.(*SignedDetails)

	if !ok {
		msg = "The token is invalid"
		return
	}

	//To check the expiry date of the claims
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "The token has expired"
		return
	}

	return claims, msg
}
