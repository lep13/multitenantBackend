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
	// Allow only POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Set necessary headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Parse the request body into LoginRequest struct
	var loginRequest models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Authenticate the user and get the user's tag
	isAuthenticated, tag, err := db.AuthenticateUser(loginRequest.Username, loginRequest.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error during authentication: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	var response models.LoginResponse

	if isAuthenticated {
		// Set the RedirectURL based on the user's role (tag)
		switch tag {
		case "admin":
			response = models.LoginResponse{
				Success:     true,
				Message:     "Login successful",
				RedirectURL: "/admin",
			}
		case "manager":
			response = models.LoginResponse{
				Success:     true,
				Message:     "Login successful",
				RedirectURL: "/manager",
			}
		case "user":
			response = models.LoginResponse{
				Success:     true,
				Message:     "Login successful",
				RedirectURL: "/user",
			}
		default:
			response = models.LoginResponse{
				Success: false,
				Message: "Invalid user tag",
			}
			w.WriteHeader(http.StatusUnauthorized)
		}
	} else {
		// If credentials are invalid, send an error response
		response = models.LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}
		w.WriteHeader(http.StatusUnauthorized)
	}

	// Encode and send the JSON response
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}
