package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"multitenant/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var newClient *mongo.Client

// "groups" collection
func GroupsCollection() *mongo.Collection {
	return newClient.Database("mydatabase").Collection("groups")
}

// "user_sessions" collection
func GetUserSessionCollection() *mongo.Collection {
	return newClient.Database("mydatabase").Collection("user_sessions")
}

// "services" collection
func GetServicesCollection() *mongo.Collection {
	return newClient.Database("mydatabase").Collection("services")
}

// Initialize MongoDB connection
func init() {
	var err error
	newClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI(config.MongoURI))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to MongoDB: %v", err))
	}
}

// GenerateSessionID generates a unique session ID
func GenerateSessionID() string {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic("Failed to generate random bytes for session ID")
	}
	return hex.EncodeToString(randomBytes) + "-" + time.Now().Format("20060102150405")
}

// StartSession starts a new session for a user
func StartSession(username, provider string) (string, error) {
    // Fetch user's group details
    var group struct {
        Groupname     string  `bson:"group_name"`
        TotalBudget   float64 `bson:"budget"`
        CurrentBudget float64 `bson:"current_budget,omitempty"` // Optional field
    }

    // Updated query to check if the user exists in the 'members' array
    err := GetGroupsCollection().FindOne(context.Background(), bson.M{"members": bson.M{"$elemMatch": bson.M{"$eq": username}}}).Decode(&group)
    if err != nil {
        return "", fmt.Errorf("group not found for the user: %v", err)
    }

    // If `current_budget` is missing, initialize it with `total_budget`
    if group.CurrentBudget == 0 {
        group.CurrentBudget = group.TotalBudget
        // Update the `groups` collection to set `current_budget`
        _, err = GetGroupsCollection().UpdateOne(context.Background(),
            bson.M{"members": bson.M{"$elemMatch": bson.M{"$eq": username}}},
            bson.M{"$set": bson.M{"current_budget": group.TotalBudget}},
        )
        if err != nil {
            return "", fmt.Errorf("failed to initialize current budget: %v", err)
        }
    }

    // Generate session ID
    sessionID := GenerateSessionID()

    // Insert session into database
    session := bson.M{
        "username":       username,
        "groupname":      group.Groupname,
        "provider":       provider,
        "session_id":     sessionID,
        "status":         "in-progress",
        "group_budget":   group.TotalBudget,
        "current_budget": group.CurrentBudget,
    }

    _, err = GetUserSessionCollection().InsertOne(context.Background(), session)
    if err != nil {
        return "", fmt.Errorf("failed to start session: %v", err)
    }

    return sessionID, nil
}

// FetchGroupBudget fetches the user's group budget
func FetchGroupBudget(username string) (float64, float64, error) {
	var group struct {
		TotalBudget   float64 `bson:"total_budget"`
		CurrentBudget float64 `bson:"current_budget"`
	}
	err := GetGroupsCollection().FindOne(context.Background(), bson.M{"username": username}).Decode(&group)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch group budget: %v", err)
	}
	return group.TotalBudget, group.CurrentBudget, nil
}

// UpdateSession updates the session document with service and date details
func UpdateSession(sessionID, service, startDate, endDate string) error {
	// Define the filter to find the session by session ID
	filter := bson.M{"session_id": sessionID}

	// Define the update object
	update := bson.M{
		"$set": bson.M{
			"service":    service,
			"start_date": startDate,
			"end_date":   endDate,
		},
	}

	// Perform the update
	result, err := GetUserSessionCollection().UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update session: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no session found with session_id: %s", sessionID)
	}

	return nil
}