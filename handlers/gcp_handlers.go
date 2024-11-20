package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"multitenant/cloud"
	"multitenant/db"
	"multitenant/models"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

// handles requests to create a GCP Compute Engine instance
func CreateComputeEngineHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID      string `json:"session_id"`
		Name           string `json:"name"`
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

	// Fetch Project ID dynamically
	projectID, err := cloud.FetchProjectID()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch GCP project ID: %v", err), http.StatusInternalServerError)
		return
	}

	// Configuration details
	config := bson.M{
		"name":            req.Name,
		"zone":            req.Zone,
		"machine_type":    req.MachineType,
		"image_project":   req.ImageProject,
		"image_family":    req.ImageFamily,
		"network":         req.Network,
		"subnetwork":      req.Subnetwork,
		"service_account": req.ServiceAccount,
		"region":          req.Region,
	}

	// Proceed with service creation
	gcpRequest := models.GCPInstanceRequest{
		Name:           req.Name,
		ProjectID:      projectID,
		Zone:           req.Zone,
		MachineType:    req.MachineType,
		ImageProject:   req.ImageProject,
		ImageFamily:    req.ImageFamily,
		Network:        req.Network,
		Subnetwork:     req.Subnetwork,
		ServiceAccount: req.ServiceAccount,
		Region:         req.Region,
	}

	_, err = cloud.CreateComputeEngineInstance(gcpRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Compute Engine instance: %v", err), http.StatusInternalServerError)
		return
	}

		// Respond with the creation result (service creation successful)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Compute Engine instance created successfully",
		"config":       config,
		"instance_name": req.Name,
		"region":       req.Region,
	})

}

// CreateCloudStorageHandler handles requests to create a GCP Cloud Storage bucket
func CreateCloudStorageHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID  string `json:"session_id"`
		BucketName string `json:"bucket_name"`
		Region     string `json:"region"`
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
	status, ok := session["status"].(string)
	if !ok || status != "ok" {
		http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
		return
	}

	// Configuration details
	config := bson.M{
		"bucket_name": req.BucketName,
		"region":      req.Region,
	}

	// Proceed with bucket creation
	_, err = cloud.CreateCloudStorage(req.BucketName, req.Region)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Cloud Storage bucket: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the creation result (service creation successful)
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
	"message":     "Cloud Storage bucket created successfully",
	"config":      config,
	"bucket_name": req.BucketName,
	"region":      req.Region,
})

}

// CreateGKEClusterHandler handles requests to create a GKE cluster
func CreateGKEClusterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID   string `json:"session_id"`
		ClusterName string `json:"cluster_name"`
		Zone        string `json:"zone"`
		Region      string `json:"region"`
		MachineType string `json:"machine_type"`
		Network     string `json:"network"`
		Subnetwork  string `json:"subnetwork"`
		NodeCount   int    `json:"node_count"`
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
	status, ok := session["status"].(string)
	if !ok || status != "ok" {
		http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
		return
	}

	// Configuration details
	config := bson.M{
		"cluster_name": req.ClusterName,
		"zone":         req.Zone,
		"region":       req.Region,
		"machine_type": req.MachineType,
		"network":      req.Network,
		"subnetwork":   req.Subnetwork,
		"node_count":   req.NodeCount,
	}

	// Call the cloud function to create the GKE cluster
	operation, err := cloud.CreateGKECluster(req.ClusterName, req.Zone, req.Region, req.MachineType, req.Network, req.Subnetwork, req.NodeCount)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create GKE cluster: %v", err), http.StatusInternalServerError)
		return
	}

// Respond with the creation result (service creation successful)
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
	"message":      "GKE cluster created successfully",
	"operation": operation,
	"config":       config,
	"cluster_name": req.ClusterName,
	"region":       req.Region,
})

}

// CreateBigQueryDatasetHandler handles requests to create a BigQuery dataset
func CreateBigQueryDatasetHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
		DatasetID string `json:"dataset_id"`
		Region    string `json:"region"`
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
	status, ok := session["status"].(string)
	if !ok || status != "ok" {
		http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
		return
	}

	// Configuration details
	config := bson.M{
		"dataset_id": req.DatasetID,
		"region":     req.Region,
	}

	// Call the cloud function to create the BigQuery dataset
	dataset, err := cloud.CreateBigQueryDataset(req.DatasetID, req.Region)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create BigQuery dataset: %v", err), http.StatusInternalServerError)
		return
	}

// Respond with the creation result (service creation successful)
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
	"message":    "BigQuery dataset created successfully",
	"dataset": dataset,
	"config":     config,
	"dataset_id": req.DatasetID,
	"region":     req.Region,
})

}

// // DeployCloudFunctionHandler handles requests to deploy a Google Cloud Function
// func DeployCloudFunctionHandler(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		FunctionName        string            `json:"function_name"`
// 		Region              string            `json:"region"`
// 		Runtime             string            `json:"runtime"`
// 		EntryPoint          string            `json:"entry_point"`
// 		BucketName          string            `json:"bucket_name"`
// 		ObjectName          string            `json:"object_name"`
// 		TriggerHTTP         bool              `json:"trigger_http"`
// 		EnvironmentVariables map[string]string `json:"environment_variables"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	op, err := cloud.DeployCloudFunction(req.FunctionName, req.Region, req.Runtime, req.EntryPoint, req.BucketName, req.ObjectName, req.EnvironmentVariables, req.TriggerHTTP)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to deploy Cloud Function: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	// Respond with operation metadata
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"message":       "Cloud Function deployed successfully",
// 		"function_name": req.FunctionName,
// 		"region":        req.Region,
// 		"operation_id":  op.GetName(),
// 	})
// }

// CreateCloudSQLHandler handles requests to create a Cloud SQL instance
func CreateCloudSQLHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID       string `json:"session_id"`
		InstanceName    string `json:"instance_name"`
		Region          string `json:"region"`
		Tier            string `json:"tier"`
		DatabaseVersion string `json:"database_version"`
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
	status, ok := session["status"].(string)
	if !ok || status != "ok" {
		http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
		return
	}

	// Configuration details
	config := bson.M{
		"instance_name":    req.InstanceName,
		"region":           req.Region,
		"tier":             req.Tier,
		"database_version": req.DatabaseVersion,
	}

	// Fetch Project ID dynamically
	projectID, err := cloud.FetchProjectID()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch GCP project ID: %v", err), http.StatusInternalServerError)
		return
	}

	// Proceed with Cloud SQL instance creation
	result, err := cloud.CreateCloudSQLInstance(req.InstanceName, projectID, req.Region, req.Tier, req.DatabaseVersion)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Cloud SQL instance: %v", err), http.StatusInternalServerError)
		return
	}

// Respond with the creation result (service creation successful)
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
	"message":          "Cloud SQL instance created successfully",
	"result":           result,
	"config":           config,
	"instance_name":    req.InstanceName,
	"region":           req.Region,
	"database_version": req.DatabaseVersion,
})

}
