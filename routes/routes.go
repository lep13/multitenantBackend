package routes

import (
	"net/http"
	"multitenant/handlers" // Ensure this import path is correct
)

// InitializeRoutes sets up all the routes for the microservice
func InitializeRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	
	// Authentication routes
	mux.HandleFunc("/login", handlers.LoginHandler) // Correct handler reference here
	return mux
}
