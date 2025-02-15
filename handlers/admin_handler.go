package handlers

import (
	"encoding/json"
	"multitenant/db"
	"multitenant/models"
	"net/http"
)

// CreateManagerHandler handles the request to create a manager
func CreateManagerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request models.CreateManagerRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate username
	if request.Username == "admin" {
		http.Error(w, "Username 'admin' is not allowed", http.StatusBadRequest)
		return
	}

	// Call AddManager to add the manager
	response := db.AddManager(request.Username, request.Password, request.Email, request.GroupLimit)
	w.Header().Set("Content-Type", "application/json")
	// Set response header to JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}


func RemoveManagerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Decode the request body into RemoveManagerRequest model
	var request models.RemoveManagerRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Ensure that the username is provided
	if request.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Call RemoveManager to remove the manager from both collections
	response := db.RemoveManager(request.Username)
	// Set response header to JSON
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}