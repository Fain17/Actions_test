package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBInstance() *mongo.Client {
	//Loading the env file and checking if it is available or not
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	//Getting the MONGODB_URI using the Getenv function
	mongoDb := os.Getenv("MONGODB_URI")

	//Creates a clientOptions var which stores the config for connecting to the mongodb instance, it uses the ApplyURI function to get the mongoDB URI
	clientOptions := options.Client().ApplyURI(mongoDb)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//Using the connect function, the ctx with a timeout of 100 seconds and clientOptions is passed as the config variable
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	//Pinging the database sever and passing a placeholder context with the clients read preference
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Server not pingable ,", err)
	}

	fmt.Println("Connected to the DB !!!")

	return client
}

// Creating a DB instance client
var Client *mongo.Client = DBInstance()

// Opening a collection using the collection variable and using the client as the connection variable
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading the env file")
	}

	dbName := os.Getenv("DB_NAME")

	var collection *mongo.Collection = client.Database(dbName).Collection(collectionName)
	return collection
}
