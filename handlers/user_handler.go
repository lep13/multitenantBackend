package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"multitenant/db"
	"net/http"
	"time"

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
			"Amazon DynamoDB",
			"AWS CloudFront",
			"Amazon VPC (Virtual Private Cloud)",
		}
	case "gcp":
		services = []string{
			"Compute Engine",
			"Cloud Storage",
			"Google Kubernetes Engine (GKE)",
			"BigQuery",
			"Cloud Functions",
			"Cloud SQL",
			"Cloud Pub/Sub",
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
	var req struct {
		SessionID string `json:"session_id"`
		Status    string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Fetch session details
	var session bson.M
	err := db.GetUserSessionCollection().FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&session)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Extract relevant fields
	estimatedCost := session["estimated_cost"].(float64)
	groupBudget := session["group_budget"].(float64)
	username := session["username"].(string)

	// Check if the session status is "denied"
	if session["status"] == "denied" || estimatedCost > groupBudget {
		// Delete the session from user_sessions collection
		_, err := db.GetUserSessionCollection().DeleteOne(context.Background(), bson.M{"session_id": req.SessionID})
		if err != nil {
			http.Error(w, "Failed to delete denied session", http.StatusInternalServerError)
			return
		}

		// Prepare and send the error response
		response := map[string]interface{}{
			"error":           "Insufficient budget",
			"message":         fmt.Sprintf("User '%s', you don't have enough budget to create this service. Estimated cost: %.2f, Group budget: %.2f. Please request your manager for additional budget.", username, estimatedCost, groupBudget),
			"estimated_cost":  estimatedCost,
			"group_budget":    groupBudget,
			"session_deleted": true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Finalize the session if status is not "denied"
	session["status"] = req.Status
	session["timestamp"] = time.Now()
	session["service_status"] = "running"

	// Insert session into the services collection
	_, err = db.GetServicesCollection().InsertOne(context.Background(), session)
	if err != nil {
		http.Error(w, "Failed to move session to services collection", http.StatusInternalServerError)
		return
	}

	// Delete session from user_sessions collection
	_, err = db.GetUserSessionCollection().DeleteOne(context.Background(), bson.M{"session_id": req.SessionID})
	if err != nil {
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Session finalized successfully"))
}
