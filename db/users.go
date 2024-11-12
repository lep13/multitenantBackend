package db

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// GetServicesCollection returns the "services" collection
func GetServicesCollection() *mongo.Collection {
	return client.Database("mydatabase").Collection("services")
}

// GetBudgetCollection returns the "budget" collection
func GetBudgetCollection() *mongo.Collection {
	return client.Database("mydatabase").Collection("budget")
}
