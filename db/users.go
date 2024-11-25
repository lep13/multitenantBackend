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

// saves service data in the `services` collection
func PushToServicesCollection(session bson.M, config bson.M) error {
    // Add configuration details and timestamp
    session["config"] = config
    session["timestamp"] = time.Now()

    // Remove `_id` to avoid duplicate key errors
    delete(session, "_id")

    // Use `Upsert` to ensure no duplicate documents
    filter := bson.M{"session_id": session["session_id"]}
    update := bson.M{"$set": session}

    _, err := GetServicesCollection().UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(true))
    if err != nil {
        return fmt.Errorf("failed to upsert into services collection: %w", err)
    }
    return nil
}

// updates the service_status if service is deleted
func UpdateawsServiceStatus(username, serviceType, identifier, status string) error {
    var filter bson.M

    // Build the filter dynamically based on the service type
    switch serviceType {
    case "Amazon S3 (Simple Storage Service)":
        filter = bson.M{
            "username":           username,
            "service":            serviceType,
            "config.bucket_name": identifier,
        }
    case "Amazon EC2 (Elastic Compute Cloud)":
        filter = bson.M{
            "username":              username,
            "service":               serviceType,
            "config.instance_name":  identifier,
        }
    case "AWS Lambda":
        filter = bson.M{
            "username":              username,
            "service":               serviceType,
            "config.function_name":  identifier,
        }
    case "Amazon RDS (Relational Database Service)":
        filter = bson.M{
            "username":              username,
            "service":               serviceType,
            "config.instance_id":    identifier,
        }
    case "AWS CloudFront":
        filter = bson.M{
            "username":              username,
            "service":               serviceType,
            "config.distribution_id": identifier,
        }
    case "Amazon VPC (Virtual Private Cloud)":
        filter = bson.M{
            "username":        username,
            "service":         serviceType,
            "config.name":     identifier,
        }
    default:
        return fmt.Errorf("unsupported service type: '%s'", serviceType)
    }

    // Log the filter for debugging
    fmt.Printf("Filter used for update: %+v\n", filter)

    // Update query
    update := bson.M{
        "$set": bson.M{
            "service_status": status,
            "end_timestamp":  time.Now(),
        },
    }

    // Execute the update operation
    result, err := GetServicesCollection().UpdateOne(context.Background(), filter, update)
    if err != nil {
        return fmt.Errorf("failed to update service status: %w", err)
    }

    // log result of update
    // fmt.Printf("Matched Count: %d, Modified Count: %d\n", result.MatchedCount, result.ModifiedCount)

    // If no matching documents were found, debug deeper
    if result.MatchedCount == 0 {
        // Debug: Fetch the document to see why it isn't matching
        var document bson.M
        err = GetServicesCollection().FindOne(context.Background(), bson.M{
            "username": username,
        }).Decode(&document)
        if err == nil {
            fmt.Printf("Fetched document for debugging: %+v\n", document)
        } else {
            fmt.Printf("Failed to fetch document for debugging: %v\n", err)
        }

        return fmt.Errorf("no matching service found for user '%s' with service type '%s' and identifier '%s'",
            username, serviceType, identifier)
    }

    return nil
}

func UpdategcpServiceStatus(username, serviceType, identifier, status string) error {
    var filter bson.M

    // Build the filter dynamically based on the service type
    switch serviceType {
    case "Compute Engine":
        filter = bson.M{
            "username":     username,
            "service":      serviceType,
            "config.name":  identifier, // Match by the "name" field in the config
        }
    case "Cloud Storage":
        filter = bson.M{
            "username":     username,
            "service":      serviceType,
            "config.bucket_name": identifier,
        }
    case "Google Kubernetes Engine (GKE)":
        filter = bson.M{
            "username":     username,
            "service":      serviceType,
            "config.cluster_name":  identifier, 
        }
    case "BigQuery":
        filter = bson.M{
            "username":     username,
            "service":      serviceType,
            "config.dataset_id": identifier,
        }
    case "Cloud SQL":
        filter = bson.M{
            "username":     username,
            "service":      serviceType,
            "config.instance_name":  identifier, 
        }
    default:
        return fmt.Errorf("unsupported service type: '%s'", serviceType)
    }

    // Log the filter for debugging
    fmt.Printf("Filter used for update: %+v\n", filter)

    // Update query
    update := bson.M{
        "$set": bson.M{
            "service_status": status,
            "end_timestamp":  time.Now(),
        },
    }

    // Execute the update operation
    result, err := GetServicesCollection().UpdateOne(context.Background(), filter, update)
    if err != nil {
        return fmt.Errorf("failed to update service status: %w", err)
    }

    // Log the result of the update
    fmt.Printf("Matched Count: %d, Modified Count: %d\n", result.MatchedCount, result.ModifiedCount)

    // If no matching documents were found, debug deeper
    if result.MatchedCount == 0 {
        // Debug: Fetch the document to see why it isn't matching
        var document bson.M
        err = GetServicesCollection().FindOne(context.Background(), bson.M{
            "username": username,
        }).Decode(&document)
        if err == nil {
            fmt.Printf("Fetched document for debugging: %+v\n", document)
        } else {
            fmt.Printf("Failed to fetch document for debugging: %v\n", err)
        }

        return fmt.Errorf("no matching service found for user '%s' with service type '%s' and identifier '%s'",
            username, serviceType, identifier)
    }

    return nil
}

// based on the username and instance name.
func GetInstanceIDByInstanceName(username, serviceType, instanceName string) (string, error) {
    var serviceData bson.M

    // Determine the appropriate field for instance name based on the service type
    var instanceNameField string
    if serviceType == "Amazon EC2 (Elastic Compute Cloud)" {
        instanceNameField = "config.instance_name"
    } else if serviceType == "Amazon RDS (Relational Database Service)" {
        instanceNameField = "config.db_name"
    } else {
        return "", fmt.Errorf("unsupported service type for instance ID retrieval: %s", serviceType)
    }

    // Query the MongoDB collection
    err := GetServicesCollection().FindOne(context.Background(), bson.M{
        "username": username,
        "service":  serviceType,
        instanceNameField: instanceName,
    }).Decode(&serviceData)

    if err != nil {
        if err == mongo.ErrNoDocuments {
            return "", fmt.Errorf("no matching document found for username '%s' and instance name '%s'", username, instanceName)
        }
        return "", fmt.Errorf("failed to fetch service details from database: %w", err)
    }

    // Log the fetched data for debugging
    fmt.Printf("Fetched service document: %+v\n", serviceData)

    // Extract the config field
    config, ok := serviceData["config"].(bson.M)
    if !ok {
        return "", fmt.Errorf("invalid config format in service document")
    }

    // Extract the instance_id
    instanceID, ok := config["instance_id"].(string)
    if !ok || instanceID == "" {
        return "", fmt.Errorf("instance ID not found for instance name '%s'", instanceName)
    }

    return instanceID, nil
}
