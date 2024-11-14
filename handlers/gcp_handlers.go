package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"multitenant/cloud"
	"multitenant/db"
	"multitenant/models"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

// CreateComputeEngineHandler handles requests to create a GCP Compute Engine instance
func CreateComputeEngineHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID      string `json:"session_id"`
		Name           string `json:"name"`
		ProjectID      string `json:"project_id"`
		Zone           string `json:"zone"`
		MachineType    string `json:"machine_type"`
		ImageProject   string `json:"image_project"`
		ImageFamily    string `json:"image_family"`
		Network        string `json:"network"`
		Subnetwork     string `json:"subnetwork"`
		ServiceAccount string `json:"service_account"`
		Region         string `json:"region"`
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

	// Check if the session is approved
	status := session["status"].(string)
	if status != "ok" {
		http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
		return
	}

	// Proceed with service creation
	gcpRequest := models.GCPInstanceRequest{
		Name:           req.Name,
		ProjectID:      req.ProjectID,
		Zone:           req.Zone,
		MachineType:    req.MachineType,
		ImageProject:   req.ImageProject,
		ImageFamily:    req.ImageFamily,
		Network:        req.Network,
		Subnetwork:     req.Subnetwork,
		ServiceAccount: req.ServiceAccount,
		Region:         req.Region,
	}

	result, err := cloud.CreateComputeEngineInstance(gcpRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Compute Engine instance: %v", err), http.StatusInternalServerError)
		return
	}

	// Finalize the session
	completeReq := struct {
		SessionID string `json:"session_id"`
		Status    string `json:"status"`
	}{
		SessionID: req.SessionID,
		Status:    "completed",
	}

	completeData, _ := json.Marshal(completeReq)
	http.Post("http://localhost:8080/complete-session", "application/json", bytes.NewReader(completeData))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
