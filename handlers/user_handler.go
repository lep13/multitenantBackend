package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"multitenant/db"
	"net/http"
	"time"
	"multitenant/models"

	"go.mongodb.org/mongo-driver/bson"
)

// GetCloudServicesHandler handles the request to fetch available services for AWS or GCP
func GetCloudServicesHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the provider from the query parameters
	provider := r.URL.Query().Get("provider")

	var services []string

	// Determine services based on the provider
	switch provider {
	case "aws":
		services = []string{
			"Amazon EC2 (Elastic Compute Cloud)",
			"Amazon S3 (Simple Storage Service)",
			"AWS Lambda",
			"Amazon RDS (Relational Database Service)",
			// "Amazon DynamoDB",
			"AWS CloudFront",
			"Amazon VPC (Virtual Private Cloud)",
		}
	case "gcp":
		services = []string{
			"Compute Engine",
			"Cloud Storage",
			"Google Kubernetes Engine (GKE)",
			"BigQuery",
			// "Cloud Functions",
			"Cloud SQL",
			// "Cloud Pub/Sub",
		}
	default:
		http.Error(w, "Invalid provider. Supported values are 'aws' and 'gcp'.", http.StatusBadRequest)
		return
	}

	// Return the list of services as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{
		"services": services,
	})
}

// StartSessionHandler starts a new session for the user
func StartSessionHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	provider := r.URL.Query().Get("provider")

	if username == "" || provider == "" {
		http.Error(w, "Missing username or provider", http.StatusBadRequest)
		return
	}

	sessionID, err := db.StartSession(username, provider)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start session: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"session_id": sessionID,
	})
}

// UpdateSessionHandler updates the session with service
func UpdateSessionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
		Service   string `json:"service"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" || req.Service == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	err := db.UpdateSession(req.SessionID, req.Service)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update session: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Session updated successfully"))
}

// finalizes the session and moves it to the services collection or deletes it if denied
func CompleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting CompleteSessionHandler")

	var req struct {
		SessionID string `json:"session_id"`
		Status    string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v\n", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Printf("Decoded request: %+v\n", req)

	var session bson.M
	err := db.GetUserSessionCollection().FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&session)
	if err != nil {
		log.Printf("Session not found: %v\n", err)
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	log.Printf("Fetched session: %+v\n", session)

	// Check config validity
	config, ok := session["config"].(bson.M)
	if !ok {
		log.Println("Invalid or missing config in session")
		http.Error(w, "Invalid or missing config in session", http.StatusInternalServerError)
		return
	}

	log.Printf("Fetched config: %+v\n", config)

	// Update session details
	session["status"] = req.Status
	session["timestamp"] = time.Now()
	session["service_status"] = "running"

	// Check if session already exists in services collection
	existing := db.GetServicesCollection().FindOne(context.Background(), bson.M{"session_id": req.SessionID})
	if existing.Err() == nil {
		log.Println("Session already exists in services collection")
		http.Error(w, "Session already exists in services collection", http.StatusConflict)
		return
	}

	// Add session to `services` collection
	err = db.PushToServicesCollection(session, config)
	if err != nil {
		log.Printf("Failed to move session to services collection: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to move session to services collection: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("Session added to services collection")

	// Fetch manager information from the groups collection
	groupID, _ := session["group_id"].(string)
	manager, err := db.GetManagerByGroupID(groupID)
	if err != nil {
		log.Printf("Failed to fetch manager by group ID: %v\n", err)
	} else {
		// Create a notification for the manager
		username, _ := session["username"].(string)
		serviceName, _ := session["service"].(string)
		cloudProvider, _ := session["provider"].(string)
		timestamp, _ := session["timestamp"].(time.Time) // Use the session's timestamp

		message := fmt.Sprintf("%s has created the service %s on %s.", username, serviceName, cloudProvider)

		notification := models.Notification{
			Manager:   manager,
			Message:   message,
			Timestamp: timestamp, // Use the existing timestamp
		}

		_, err := db.GetNotificationsCollection().InsertOne(context.Background(), notification)
		if err != nil {
			log.Printf("Failed to save notification: %v\n", err)
		} else {
			log.Println("Notification saved successfully")
		}
	}

	// Delete session from `user_sessions` collection
	_, err = db.GetUserSessionCollection().DeleteOne(context.Background(), bson.M{"session_id": req.SessionID})
	if err != nil {
		log.Printf("Failed to delete session: %v\n", err)
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}
	log.Println("Session deleted from user_sessions collection")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Session finalized successfully"))
	log.Println("CompleteSessionHandler finished successfully")
}

func SendNotificationHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username        string  `json:"username"`
		Manager         string  `json:"manager"`
		RequestedService string `json:"requested_service"`
		EstimatedCost   float64 `json:"estimated_cost"`
		Budget          float64 `json:"budget"` 
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Construct the notification message
	message := fmt.Sprintf(
		"%s has requested an increase in budget to create the service %s with an estimated cost of %.2f. Current budget is %.2f.",
		req.Username, req.RequestedService, req.EstimatedCost, req.Budget,
	)

	notification := models.Notification{
		Manager:   req.Manager,
		Message:   message,
		Timestamp: time.Now(),
	}

	// Save notification to the notifications collection
	_, err := db.GetNotificationsCollection().InsertOne(context.Background(), notification)
	if err != nil {
		log.Printf("Failed to save notification: %v\n", err)
		http.Error(w, "Failed to save notification", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Notification sent successfully"))
}

// NotifyManagerOnServiceAction sends notifications for service creation or deletion
func NotifyManagerOnServiceAction(username string, service string, action string, groupID string, timestamp time.Time) error {
    // Fetch manager name using groupID
    var group bson.M
    err := db.GetGroupsCollection().FindOne(context.Background(), bson.M{"group_id": groupID}).Decode(&group)
    if err != nil {
        return fmt.Errorf("Failed to fetch manager for group: %v", err)
    }

    manager, ok := group["manager"].(string)
    if !ok || manager == "" {
        return fmt.Errorf("Manager name not found for group: %s", groupID)
    }

    // Construct notification message
    message := fmt.Sprintf("%s has %s service %s on %s.", username, action, service, timestamp.Format("Jan 02, 2006 15:04:05"))

    // Create notification object
    notification := models.Notification{
        Manager:   manager,
        Message:   message,
        Timestamp: timestamp,
    }

    // Save notification to the database
    _, err = db.GetNotificationsCollection().InsertOne(context.Background(), notification)
    if err != nil {
        return fmt.Errorf("Failed to save notification: %v", err)
    }

    return nil
}
