package db

import (
	"context"
	"multitenant/config"
	"multitenant/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// ConnectMongoDB initializes the MongoDB client
func ConnectMongoDB() error {
	clientOptions := options.Client().ApplyURI(config.MongoURI)
	var err error
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

	// Ping MongoDB to verify connection
	return client.Ping(context.TODO(), nil)
}

// DisconnectMongoDB closes the MongoDB connection
func DisconnectMongoDB() {
	if client != nil {
		client.Disconnect(context.TODO())
	}
}

// AuthenticateUser checks if the user exists with the correct credentials in the users collection
func AuthenticateUser(username, password string) (bool, error) {
	collection := client.Database(config.DatabaseName).Collection("users")

	// Check if the user exists with the given username and password
	var user models.User
	err := collection.FindOne(context.TODO(), bson.M{"username": username, "password": password}).Decode(&user)
	if err != nil {
		// If the error is not nil, it might mean the user is not found
		if err == mongo.ErrNoDocuments {
			return false, nil // User not found
		}
		return false, err // Other error
	}

	// If user is found and matched
	return true, nil
}
