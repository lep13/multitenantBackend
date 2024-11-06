package handlers

import (
	"encoding/json"
	"fmt"
	"multitenant/db"
	"multitenant/models"
	"net/http"
)

// LoginHandler handles the login requests
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Decode the request body into LoginRequest
	var loginRequest models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Authenticate the user using the provided credentials
	isAuthenticated, err := db.AuthenticateUser(loginRequest.Username, loginRequest.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error during authentication: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	var response models.LoginResponse
	if isAuthenticated {
		response = models.LoginResponse{
			Success: true,
			Message: "Login successful",
		}
	} else {
		response = models.LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}
	}
  
	// Set response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Send the response
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}
