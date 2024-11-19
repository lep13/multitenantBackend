package db

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"multitenant/config"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var newClient *mongo.Client

// Collections
func GetUserSessionCollection() *mongo.Collection {
	return newClient.Database("mydatabase").Collection("user_sessions")
}

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
	var group struct {
		Groupname   string  `bson:"group_name"`
		GroupID     string  `bson:"group_id"`
		TotalBudget float64 `bson:"budget"`
		// CurrentBudget float64 `bson:"current_budget,omitempty"`
	}

	// Fetch the group the user belongs to
	err := GetGroupsCollection().FindOne(context.Background(), bson.M{"members": bson.M{"$elemMatch": bson.M{"$eq": username}}}).Decode(&group)
	if err != nil {
		return "", fmt.Errorf("group not found for the user: %v", err)
	}

	// // Initialize current budget if not present
	// if group.CurrentBudget == 0 {
	// 	group.CurrentBudget = group.TotalBudget
	// 	_, err = GetGroupsCollection().UpdateOne(context.Background(),
	// 		bson.M{"group_id": group.GroupID},
	// 		bson.M{"$set": bson.M{"current_budget": group.TotalBudget}})
	// 	if err != nil {
	// 		return "", fmt.Errorf("failed to initialize current budget: %v", err)
	// 	}
	// }

	// Create and store session
	sessionID := GenerateSessionID()
	session := bson.M{
		"username":     username,
		"groupname":    group.Groupname,
		"group_id":     group.GroupID,
		"provider":     provider,
		"session_id":   sessionID,
		"status":       "in-progress",
		"group_budget": group.TotalBudget,
		// "current_budget": group.CurrentBudget,
	}

	_, err = GetUserSessionCollection().InsertOne(context.Background(), session)
	if err != nil {
		return "", fmt.Errorf("failed to start session: %v", err)
	}

	return sessionID, nil
}

// UpdateSession updates the session with the selected service
func UpdateSession(sessionID, service string) error {
	filter := bson.M{"session_id": sessionID}
	update := bson.M{
		"$set": bson.M{
			"service": service,
		},
	}

	result, err := GetUserSessionCollection().UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update session: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no session found with session_id: %s", sessionID)
	}

	return nil
}

// UpdateSessionWithCost updates the session with the estimated cost (quarterly) and status
func UpdateSessionWithCost(sessionID string, estimatedCost float64, status string) (string, error) {
	// Define the filter to find the session by session ID
	filter := bson.M{"session_id": sessionID}

	// Define the update object
	update := bson.M{
		"$set": bson.M{
			"estimated_cost": estimatedCost,
			"status":         status,
		},
	}

	// Perform the update
	result, err := GetUserSessionCollection().UpdateOne(context.Background(), filter, update)
	if err != nil {
		return "", fmt.Errorf("failed to update session: %v", err)
	}

	if result.MatchedCount == 0 {
		return "", fmt.Errorf("no session found with session_id: %s", sessionID)
	}

	return status, nil
}

// sends a POST request to finalize a session using the user's JWT token
func FinalizeSessionWithJWT(sessionID, token string) error {
	// Prepare the request payload
	completeReq := struct {
		SessionID string `json:"session_id"`
	}{
		SessionID: sessionID,
	}
	completeData, _ := json.Marshal(completeReq)

	// Create a new POST request
	client := &http.Client{}
	request, _ := http.NewRequest("POST", "http://localhost:8080/user/complete-session", bytes.NewReader(completeData))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", token)

	// Execute the request
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to finalize session: %v", err)
	}
	defer response.Body.Close()

	// Check the response status
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("failed to finalize session: %s", body)
	}

	return nil
}

// DeleteSession deletes an incomplete session
func DeleteSession(sessionID string) error {
	_, err := GetUserSessionCollection().DeleteOne(context.Background(), bson.M{"session_id": sessionID})
	return err
}

// MarkSessionCompleted updates the session with a "completed" status and service status
func MarkSessionCompleted(sessionID string, serviceStatus string) error {
	filter := bson.M{"session_id": sessionID}
	update := bson.M{
		"$set": bson.M{
			"status":         "completed",
			"service_status": serviceStatus,
			"timestamp":      time.Now(),
		},
	}
	_, err := GetUserSessionCollection().UpdateOne(context.Background(), filter, update)
	return err
}

// PushToServicesCollection moves completed session data to the services collection
func PushToServicesCollection(session bson.M) error {
	_, err := GetServicesCollection().InsertOne(context.Background(), session)
	return err
}
