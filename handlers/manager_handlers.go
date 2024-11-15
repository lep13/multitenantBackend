package handlers

import (
	"encoding/json"
	"multitenant/db"
	"net/http"
	"strings"
)

/// CreateUserHandler handles the creation of a new user
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	response := db.CreateUser(input.Username, input.Password, input.Email)

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}

// CreateGroupHandler handles the creation of a new group
func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username  string `json:"username"`
		GroupName string `json:"group_name"`
	}

	// Decode the request body
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if strings.TrimSpace(input.Username) == "" || strings.TrimSpace(input.GroupName) == "" {
		http.Error(w, "Username and group name cannot be empty or whitespace", http.StatusBadRequest)
		return
	}

	// Call CreateGroup logic
	response := db.CreateGroup(input.Username, input.GroupName)

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}

// AddUserHandler adds an existing user to a group
func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Manager  string `json:"manager"`
		GroupID  string `json:"group_id"`
		Username string `json:"username"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	response := db.AddUserToGroup(input.Manager, input.GroupID, input.Username)
	// Send the response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}

// RemoveUserHandler removes a user from a group
func RemoveUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Manager  string `json:"manager"`
		GroupID  string `json:"group_id"`
		Username string `json:"username"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	} 

	response := db.RemoveUserFromGroup(input.Manager, input.GroupID, input.Username)
	// Send the response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}

// DeleteUserHandler handles the deletion of a user
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
	}
	// Decode the JSON request body to get the username
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil || input.Username == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// Call the DeleteUser function to delete the user
	response := db.DeleteUser(input.Username)
	// Send the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func ListGroupsHandler(w http.ResponseWriter, r *http.Request) {
    username := r.URL.Query().Get("username")

    if username == "" {
        http.Error(w, "Invalid input: username is required", http.StatusBadRequest)
        return
    }

    response := db.ListGroupsByManager(username)

    // Set headers for the response
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

    // Send the response
    if response.Status == "error" {
        w.WriteHeader(http.StatusBadRequest)
    } else {
        w.WriteHeader(http.StatusOK)
    }
    json.NewEncoder(w).Encode(response)
}


// AddBudgetHandler assigns a budget to a group
func AddBudgetHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Manager  string  `json:"manager"`
		GroupID  string  `json:"group_id"`
		Budget   float64 `json:"budget"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	response := db.AddBudget(input.Manager, input.GroupID, input.Budget)
	// Send the response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}

func UpdateBudgetHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Manager   string  `json:"manager"`
		GroupName string  `json:"group_name"`
		Budget    float64 `json:"budget"`
	}
	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// Validate input
	if strings.TrimSpace(input.Manager) == "" || strings.TrimSpace(input.GroupName) == "" {
		http.Error(w, "Manager and Group name cannot be empty or whitespace", http.StatusBadRequest)
		return
	}
	if input.Budget <= 0 {
		http.Error(w, "Budget must be greater than zero", http.StatusBadRequest)
		return
	}

	// Call UpdateBudget function
	response := db.UpdateBudget(input.Manager, input.GroupName, input.Budget)
	// Send the response
	w.Header().Set("Content-Type", "application/json")
	if response.Status == "error" {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// CheckUserGroupHandler handles checking if a user is already part of a group
func CheckUserGroupHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the username from the query parameters
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Invalid input: 'username' parameter is required", http.StatusBadRequest)
		return
	}

	// Call the CheckUserGroup function
	response := db.CheckUserGroup(username)

	// Set headers for CORS
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Write the response
	if response.Status == "error" {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}