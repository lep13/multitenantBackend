package handlers

import (
	"encoding/json"
	"fmt"
	"multitenant/db"
	"net/http"
)

// StartSessionHandler triggers a new session
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