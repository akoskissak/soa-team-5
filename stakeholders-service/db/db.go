package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

func Connect() {
	mongoUri := os.Getenv("MONGODB_URI")
	if mongoUri == "" {
        log.Fatal("MONGODB_URI env not set")
    }
	
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://" + mongoUri + "?retryWrites=true&w=majority").
		SetServerAPIOptions(serverAPIOptions)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("Connected!")
	MongoClient = client
}