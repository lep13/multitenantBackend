package handlers
 
import (
    "encoding/json"
    "fmt"
    "multitenant/db"
    "multitenant/models"
    "net/http"
    "os"
    "time"
 
    "github.com/golang-jwt/jwt/v4"
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
        // Generate JWT token
        expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
        claims := &Claims{
            Username: loginRequest.Username,
            Tag:      tag,
            RegisteredClaims: jwt.RegisteredClaims{
                ExpiresAt: jwt.NewNumericDate(expirationTime),
            },
        }
 
        // Create the JWT token
        jwtKey := []byte(os.Getenv("JWT_SECRET")) // Fetch JWT secret from environment
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        tokenString, err := token.SignedString(jwtKey)
        if err != nil {
            http.Error(w, "Failed to generate token", http.StatusInternalServerError)
            return
        }
 
        // Include the token in the response
        response = models.LoginResponse{
            Success:     true,
            Message:     "Login successful",
            Token:       tokenString, // Add the token to the response
            RedirectURL: getRedirectURL(tag), // Get the appropriate URL based on tag
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
 
// getRedirectURL returns the appropriate redirect URL based on the user's tag
func getRedirectURL(tag string) string {
    switch tag {
    case "admin":
        return "/admin"
    case "manager":
        return "/manager"
    case "user":
        return "/user"
    default:
        return "/"
    }
}