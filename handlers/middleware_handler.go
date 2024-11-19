package handlers
 
import (
    "context"
    "net/http"
    "os"
    "strings"
 
    "github.com/golang-jwt/jwt/v4"
)
 
// Claims struct for decoding JWT tokens
type Claims struct {
    Username string `json:"username"`
    Tag      string `json:"tag"`
    jwt.RegisteredClaims
}
 
// Middleware to handle CORS
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200") //  frontend's origin
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
 
        // Handle preflight requests
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
 
        // Pass the request to the next middleware/handler
        next.ServeHTTP(w, r)
    })
}
 
// Middleware to verify JWT token
func Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenStr := r.Header.Get("Authorization")
        if tokenStr == "" || !strings.HasPrefix(tokenStr, "Bearer ") {
            http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
            return
        }
        tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
 
        claims := &Claims{}
        jwtKey := []byte(os.Getenv("JWT_SECRET")) // Fetch JWT secret from environment
        token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
            return jwtKey, nil
        })
        if err != nil || !token.Valid {
            http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
            return
        }
 
        // Add username and tag to context
        r = r.WithContext(context.WithValue(r.Context(), "username", claims.Username))
        r = r.WithContext(context.WithValue(r.Context(), "tag", claims.Tag))
 
        next.ServeHTTP(w, r)
    })
}
 
// Middleware to check role-based access
func Authorize(allowedTags ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tag := r.Context().Value("tag").(string)
            for _, allowed := range allowedTags {
                if tag == allowed {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            http.Error(w, "Forbidden: Access denied", http.StatusForbidden)
        })
    }
}