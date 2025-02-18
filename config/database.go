package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Client {
	log.Println("Connecting to MongoDB....")
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("MongoDB is not reachable %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB is not reachable %v", err)
	}

	log.Println("Successfuly connected to MongoDB!!!! ")
	return client
}

var Client *mongo.Client = ConnectDB()

func OpenCollection(collectionName string) *mongo.Collection {
	if Client == nil {
		log.Fatal("MongoDB Client is not initalized. Please ConnectDB first.")
	}

	return Client.Database("userdb").Collection(collectionName)
}
