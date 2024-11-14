package models

// User represents a user document in MongoDB
type User struct {
    Username string `bson:"username"`
    Password string `bson:"password"`
    Email    string `bson:"email"`
    Tag      string `bson:"tag"`
}

// LoginRequest represents the structure of the login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the structure of the login response
type LoginResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	RedirectURL string `json:"redirectURL,omitempty"`
}
