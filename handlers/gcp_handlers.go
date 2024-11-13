package handlers

import (
	"encoding/json"
	"fmt"
	"multitenant/cloud"
	"multitenant/models"
	"net/http"
)

// CreateComputeEngineHandler handles GCP Compute Engine instance creation requests
func CreateComputeEngineHandler(w http.ResponseWriter, r *http.Request) {
	var req models.GCPInstanceRequest

	// Decode the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Call the cloud function to create the instance
	op, err := cloud.CreateComputeEngineInstance(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not create Compute Engine instance: %v", err), http.StatusInternalServerError)
		return
	}

	// Extract meaningful data from the operation
	operationName := op.Name()        // Call Name() method to get the string
	// operationStatus := op.Status()   // Call Status() method to get the string

	// Prepare a meaningful response
	response := map[string]string{
		"status":          "success",
		"message":         fmt.Sprintf("VM '%s' was created successfully.", req.Name),
		"operation_name":  operationName,
		// "operation_status": operationStatus,
	}

	// Set the response header and encode the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
