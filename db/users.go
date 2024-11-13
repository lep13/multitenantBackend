package db

import (
	"context"
	"fmt"
	"log"
	"multitenant/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

// CheckBudgetForRequest checks if a requested service fits within the group's remaining budget.
func CheckBudgetForRequest(groupName string, requestedCost float64) (bool, string) {
	var budgetInfo models.Budget
	err := GetBudgetCollection().FindOne(context.Background(), bson.M{"group_name": groupName}).Decode(&budgetInfo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, "Group budget not found"
		}
		log.Printf("Error checking budget for request: %v", err)
		return false, "Error accessing budget data"
	}

	// Calculate the remaining budget
	remainingBudget := budgetInfo.BudgetRemaining

	if requestedCost <= remainingBudget {
		return true, "Request is within budget"
	} else {
		return false, fmt.Sprintf("Request exceeds remaining budget. Available: $%.2f, Requested: $%.2f", remainingBudget, requestedCost)
	}
}

// UpdateBudgetUsage updates the budget usage of a group after a service request is created.
func UpdateBudgetUsage(groupName string, cost float64) (bool, string) {
	// Increment the used budget by the cost of the requested service and adjust remaining budget
	update := bson.M{
		"$inc": bson.M{
			"budget_used":      cost,
			"budget_remaining": -cost,
		},
	}
	result, err := GetBudgetCollection().UpdateOne(context.Background(), bson.M{"group_name": groupName}, update)

	if err != nil {
		log.Printf("Error updating budget usage: %v", err)
		return false, "Failed to update budget usage"
	}

	if result.MatchedCount == 0 {
		return false, "Group budget not found"
	}

	return true, "Budget updated successfully"
}

// CreateServiceRequest inserts a new service request into the services collection.
func CreateServiceRequest(service models.ServiceRequest) (bool, string) {
	_, err := GetServicesCollection().InsertOne(context.Background(), service)
	if err != nil {
		log.Printf("Error creating service request: %v", err)
		return false, "Failed to create service request"
	}
	return true, "Service request created successfully"
}

// UpdateServiceStatus updates the status of a service when its end date is reached or it is manually terminated.
func UpdateServiceStatus(serviceID string, newStatus string) (bool, string) {
	filter := bson.M{"service_id": serviceID}
	update := bson.M{
		"$set": bson.M{
			"status":   newStatus,
			"end_date": time.Now(),
		},
	}
	result, err := GetServicesCollection().UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error updating service status: %v", err)
		return false, "Failed to update service status"
	}
	if result.MatchedCount == 0 {
		return false, "Service not found"
	}
	return true, "Service status updated successfully"
}

// ListActiveServices retrieves active services for a user or group.
func ListActiveServices(userID string, groupName string) ([]models.ServiceRequest, string) {
	filter := bson.M{"status": "running"}
	if userID != "" {
		filter["user_id"] = userID
	} else if groupName != "" {
		filter["group_name"] = groupName
	} else {
		return nil, "No user ID or group name provided"
	}

	cursor, err := GetServicesCollection().Find(context.Background(), filter)
	if err != nil {
		log.Printf("Error listing active services: %v", err)
		return nil, "Failed to retrieve active services"
	}
	defer cursor.Close(context.Background())

	var services []models.ServiceRequest
	if err = cursor.All(context.Background(), &services); err != nil {
		log.Printf("Error decoding active services: %v", err)
		return nil, "Failed to decode active services"
	}
	return services, "Active services retrieved successfully"
}

// ListTerminatedServices retrieves terminated services for a user or group.
func ListTerminatedServices(userID string, groupName string) ([]models.ServiceRequest, string) {
	filter := bson.M{"status": "terminated"}
	if userID != "" {
		filter["user_id"] = userID
	} else if groupName != "" {
		filter["group_name"] = groupName
	} else {
		return nil, "No user ID or group name provided"
	}

	cursor, err := GetServicesCollection().Find(context.Background(), filter)
	if err != nil {
		log.Printf("Error listing terminated services: %v", err)
		return nil, "Failed to retrieve terminated services"
	}
	defer cursor.Close(context.Background())

	var services []models.ServiceRequest
	if err = cursor.All(context.Background(), &services); err != nil {
		log.Printf("Error decoding terminated services: %v", err)
		return nil, "Failed to decode terminated services"
	}
	return services, "Terminated services retrieved successfully"
}
