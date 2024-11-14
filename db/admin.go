package db

import (
    "context"
    "fmt"
    "multitenant/config"
    "multitenant/models"
    "strings"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// AddManager adds a new manager and user to the respective collections
func AddManager(username, password, email string, groupLimit int) models.ManagerResponse {
    // Input validation
    if strings.TrimSpace(username) == "" {
        return models.ManagerResponse{
            Success: false,
            Message: "Username cannot be empty",
        }
    }
    if !isValidUsernameLength(username) {
        return models.ManagerResponse{
            Success: false,
            Message: "Username must be at least 6 characters long",
        }
    }
    if !containsOnlyAllowedUsernameCharacters(username) {
        return models.ManagerResponse{
            Success: false,
            Message: "Username can only contain alphabets, numbers, '-', and '_'",
        }
    }
    if strings.Contains(username, " ") {
        return models.ManagerResponse{
            Success: false,
            Message: "Username cannot contain spaces",
        }
    }
    if strings.TrimSpace(password) == "" {
        return models.ManagerResponse{
            Success: false,
            Message: "Password cannot be empty",
        }
    }
    if !isValidPasswordLength(password) {
        return models.ManagerResponse{
            Success: false,
            Message: "Password must be at least 6 characters long",
        }
    }
    if !containsUppercase(password) {
        return models.ManagerResponse{
            Success: false,
            Message: "Password must contain at least one uppercase letter",
        }
    }
    if !containsLowercase(password) {
        return models.ManagerResponse{
            Success: false,
            Message: "Password must contain at least one lowercase letter",
        }
    }
    if !containsNumber(password) {
        return models.ManagerResponse{
            Success: false,
            Message: "Password must contain at least one number",
        }
    }
    if !containsSpecialCharacter(password) {
        return models.ManagerResponse{
            Success: false,
            Message: "Password must contain at least one special character (!@#$%^&*)",
        }
    }
    if strings.TrimSpace(email) == "" {
        return models.ManagerResponse{
            Success: false,
            Message: "Email cannot be empty",
        }
    }
    if !isValidEmail(email) {
        return models.ManagerResponse{
            Success: false,
            Message: "Invalid email format. Email must contain '@' and '.com'",
        }
    }
    if groupLimit <= 0 {
        return models.ManagerResponse{
            Success: false,
            Message: "Group limit must be greater than zero",
        }
    }

    clientOptions := options.Client().ApplyURI(config.MongoURI)
    client, err := mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        return models.ManagerResponse{
            Success: false,
            Message: fmt.Sprintf("Could not connect to MongoDB: %v", err),
        }
    }
    defer client.Disconnect(context.TODO())

    // Define collections
    managerCollection := client.Database("mydatabase").Collection("managers")
    userCollection := client.Database("mydatabase").Collection("users")

    // Check if a manager with the same username already exists
    var existingManager models.Manager
    err = managerCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&existingManager)
    if err == nil {
        return models.ManagerResponse{
            Success: false,
            Message: fmt.Sprintf("Username '%s' already exists", username),
        }
    } else if err != mongo.ErrNoDocuments {
        return models.ManagerResponse{
            Success: false,
            Message: fmt.Sprintf("Error checking for existing manager: %v", err),
        }
    }

    // Insert into managers collection
    manager := models.Manager{
        Username:   username,
        Email:      email,
        GroupLimit: groupLimit,
    }
    _, err = managerCollection.InsertOne(context.Background(), manager)
    if err != nil {
        return models.ManagerResponse{
            Success: false,
            Message: fmt.Sprintf("Could not insert manager: %v", err),
        }
    }

    // Insert into users collection
    user := models.User{
        Username: username,
        Password: password,
        Email:    email,
        Tag:      "manager",
    }
    _, err = userCollection.InsertOne(context.Background(), user)
    if err != nil {
        return models.ManagerResponse{
            Success: false,
            Message: fmt.Sprintf("Manager created, but could not add user: %v", err),
        }
    }

    return models.ManagerResponse{
        Success: true,
        Message: "Manager created successfully",
    }
}

// Helper functions for validation

func isValidUsernameLength(username string) bool {
    return len(username) >= 6
}

func containsOnlyAllowedUsernameCharacters(username string) bool {
    for _, char := range username {
        if !(char == '-' || char == '_' || (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
            return false
        }
    }
    return true
}

func isValidPasswordLength(password string) bool {
    return len(password) >= 6
}

func containsUppercase(password string) bool {
    for _, char := range password {
        if char >= 'A' && char <= 'Z' {
            return true
        }
    }
    return false
}

func containsLowercase(password string) bool {
    for _, char := range password {
        if char >= 'a' && char <= 'z' {
            return true
        }
    }
    return false
}

func containsNumber(password string) bool {
    for _, char := range password {
        if char >= '0' && char <= '9' {
            return true
        }
    }
    return false
}

func containsSpecialCharacter(password string) bool {
    specialChars := "!@#$%^&*"
    for _, char := range password {
        if strings.ContainsRune(specialChars, char) {
            return true
        }
    }
    return false
}

func isValidEmail(email string) bool {
    return strings.Contains(email, "@") && strings.HasSuffix(email, ".com")
}

// RemoveManager removes a manager and corresponding user from the collections
func RemoveManager(username string) models.ManagerResponse {
    // Input validation with whitespace trimming
    if strings.TrimSpace(username) == "" {
        return models.ManagerResponse{
            Success: false,
            Message: "username cannot be empty",
        }
    }

    clientOptions := options.Client().ApplyURI(config.MongoURI)
    client, err := mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        return models.ManagerResponse{
            Success: false,
            Message: fmt.Sprintf("could not connect to MongoDB: %v", err),
        }
    }
    defer client.Disconnect(context.TODO())

    // Define collections
    managerCollection := client.Database("mydatabase").Collection("managers")
    userCollection := client.Database("mydatabase").Collection("users")

    // Check if the manager exists in the managers collection
    var existingManager models.Manager
    err = managerCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&existingManager)
    if err == mongo.ErrNoDocuments {
        return models.ManagerResponse{
            Success: false,
            Message: fmt.Sprintf("manager with username '%s' does not exist", username),
        }
    } else if err != nil {
        return models.ManagerResponse{
            Success: false,
            Message: fmt.Sprintf("error checking for existing manager: %v", err),
        }
    }

    // Remove the manager from the managers collection
    _, err = managerCollection.DeleteOne(context.Background(), bson.M{"username": username})
    if err != nil {
        return models.ManagerResponse{
            Success: false,
            Message: fmt.Sprintf("could not remove manager: %v", err),
        }
    }

    // Remove the user from the users collection (the user tag is "manager")
    _, err = userCollection.DeleteOne(context.Background(), bson.M{"username": username, "tag": "manager"})
    if err != nil {
        return models.ManagerResponse{
            Success: false,
            Message: fmt.Sprintf("could not remove user: %v", err),
        }
    }
    return models.ManagerResponse{
        Success: true,
        Message: fmt.Sprintf("manager with username '%s' removed successfully", username),
    }
}
