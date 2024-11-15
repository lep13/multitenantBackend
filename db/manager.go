package db
 
import (
    "context"
    "fmt"
    "log"
    "multitenant/config"
    "multitenant/models"
    "strings"
    "crypto/rand"
    "encoding/hex"
 
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)
 
var Client *mongo.Client
 
func GetManagerCollection() *mongo.Collection {
    return Client.Database("mydatabase").Collection("managers")
}
 
func GetGroupsCollection() *mongo.Collection {
    return Client.Database("mydatabase").Collection("groups")
}
 
func GetUsersCollection() *mongo.Collection {
    return Client.Database("mydatabase").Collection("users")
}
 
// Initialize MongoDB connection
func InitMongoDB() {
    var err error
    Client, err = mongo.Connect(context.Background(), options.Client().ApplyURI(config.MongoURI))
    if err != nil {
        log.Fatal(err)
    }
}
 
// CreateUser creates a new user with validations
func CreateUser(username, password, email string) models.UserResponse {
    // Username validation
    if strings.TrimSpace(username) == "" {
        return models.UserResponse{
            Message: "Username cannot be empty",
            Status:  "error",
        }
    }
    if !isValidUsernameLength(username) {
        return models.UserResponse{
            Message: "Username must be at least 6 characters long",
            Status:  "error",
        }
    }
    if !containsOnlyAllowedUsernameCharacters(username) {
        return models.UserResponse{
            Message: "Username can only contain alphabets, numbers, '-', and '_'",
            Status:  "error",
        }
    }
    if strings.Contains(username, " ") {
        return models.UserResponse{
            Message: "Username cannot contain spaces",
            Status:  "error",
        }
    }
 
    // Password validation
    if strings.TrimSpace(password) == "" {
        return models.UserResponse{
            Message: "Password cannot be empty",
            Status:  "error",
        }
    }
    if !isValidPasswordLength(password) {
        return models.UserResponse{
            Message: "Password must be at least 6 characters long",
            Status:  "error",
        }
    }
    if !containsUppercase(password) {
        return models.UserResponse{
            Message: "Password must contain at least one uppercase letter",
            Status:  "error",
        }
    }
    if !containsLowercase(password) {
        return models.UserResponse{
            Message: "Password must contain at least one lowercase letter",
            Status:  "error",
        }
    }
    if !containsNumber(password) {
        return models.UserResponse{
            Message: "Password must contain at least one number",
            Status:  "error",
        }
    }
    if !containsSpecialCharacter(password) {
        return models.UserResponse{
            Message: "Password must contain at least one special character (!@#$%^&*)",
            Status:  "error",
        }
    }
 
    // Email validation
    if strings.TrimSpace(email) == "" {
        return models.UserResponse{
            Message: "Email cannot be empty",
            Status:  "error",
        }
    }
    if !isValidEmail(email) {
        return models.UserResponse{
            Message: "Invalid email format. Email must contain '@' and '.com'",
            Status:  "error",
        }
    }
 
    // Check if the user already exists
    var existingUser bson.M
    err := GetUsersCollection().FindOne(context.Background(), bson.M{"username": username}).Decode(&existingUser)
    if err == nil {
        return models.UserResponse{
            Message: "Failed to create user: Username already exists",
            Status:  "error",
        }
    } else if err != mongo.ErrNoDocuments {
        return models.UserResponse{
            Message: fmt.Sprintf("Error checking user existence: %v", err),
            Status:  "error",
        }
    }
 
    // Insert user into the collection
    user := bson.M{
        "username": username,
        "password": password,
        "email":    email,
        "tag":      "user",
    }
    _, err = GetUsersCollection().InsertOne(context.Background(), user)
    if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("Error creating user: %v", err),
            Status:  "error",
        }
    }
 
    return models.UserResponse{
        Message: "User created successfully",
        Status:  "success",
    }
}
 
// GenerateGroupID generates a unique group ID
func GenerateGroupID() string {
    randomBytes := make([]byte, 8) // 8 bytes for a 16-character unique ID
    _, err := rand.Read(randomBytes)
    if err != nil {
        panic("Failed to generate group ID")
    }
    return hex.EncodeToString(randomBytes)
}
 
// CreateGroup logic with group_id
func CreateGroup(username, groupName string) models.UserResponse {
    var manager models.Manager
 
    // Check if the manager exists
    err := GetManagerCollection().FindOne(context.Background(), bson.M{"username": username}).Decode(&manager)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return models.UserResponse{
                Message: "Manager not found",
                Status:  "error",
            }
        }
        return models.UserResponse{
            Message: fmt.Sprintf("Error finding manager: %v", err),
            Status:  "error",
        }
    }
 
    // Check if the group name already exists for the manager
    var existingGroup bson.M
    err = GetGroupsCollection().FindOne(context.Background(), bson.M{"manager": username, "group_name": groupName}).Decode(&existingGroup)
    if err == nil {
        return models.UserResponse{
            Message: "Group with this name already exists for the manager",
            Status:  "error",
        }
    } else if err != mongo.ErrNoDocuments {
        return models.UserResponse{
            Message: fmt.Sprintf("Error checking existing groups: %v", err),
            Status:  "error",
        }
    }
 
    // Check if the group limit has been reached
    groupCount, err := GetGroupsCollection().CountDocuments(context.Background(), bson.M{"manager": username})
    if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("Error counting groups: %v", err),
            Status:  "error",
        }
    }
 
    if groupCount >= int64(manager.GroupLimit) {
        return models.UserResponse{
            Message: "Cannot create more groups, group limit reached",
            Status:  "error",
        }
    }
 
    // Generate a unique group ID
    groupID := GenerateGroupID()
 
    // Create the group object
    group := bson.M{
        "group_id":   groupID,
        "manager":    username,
        "group_name": groupName,
        "members":    []string{},
    }
 
    // Insert the group into the database
    _, err = GetGroupsCollection().InsertOne(context.Background(), group)
    if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("Error creating group: %v", err),
            Status:  "error",
        }
    }
 
    return models.UserResponse{
        Message: "Group created successfully",
        Status:  "success",
        Data:    bson.M{"group_id": groupID}, // Include group_id in the response
    }
}
 
// AddUserToGroup adds a user to a group by group ID
func AddUserToGroup(manager, groupID, username string) models.UserResponse {
    // Check if the user exists
    var user bson.M
    err := GetUsersCollection().FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
    if err == mongo.ErrNoDocuments {
        return models.UserResponse{
            Message: "User does not exist",
            Status:  "error",
        }
    } else if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("Error checking user existence: %v", err),
            Status:  "error",
        }
    }
 
    // Check if the user is already in another group
    filter := bson.M{"members": username}
    var existingGroup models.Group
    err = GetGroupsCollection().FindOne(context.Background(), filter).Decode(&existingGroup)
    if err == nil {
        return models.UserResponse{
            Message: "User is already a member of another group and cannot be added to multiple groups",
            Status:  "error",
        }
    } else if err != mongo.ErrNoDocuments {
        return models.UserResponse{
            Message: fmt.Sprintf("Error checking user's group membership: %v", err),
            Status:  "error",
        }
    }
 
    // Add user to the specified group
    update := bson.M{"$push": bson.M{"members": username}}
    _, err = GetGroupsCollection().UpdateOne(context.Background(), bson.M{"manager": manager, "group_id": groupID}, update)
    if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("Error adding user to group: %v", err),
            Status:  "error",
        }
    }
 
    return models.UserResponse{
        Message: "User added to group successfully",
        Status:  "success",
    }
}
 
 
// RemoveUserFromGroup removes a user from a group by group ID
func RemoveUserFromGroup(manager, groupID, username string) models.UserResponse {
    update := bson.M{"$pull": bson.M{"members": username}}
    result, err := GetGroupsCollection().UpdateOne(context.Background(), bson.M{"manager": manager, "group_id": groupID}, update)
    if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("Error removing user from group: %v", err),
            Status:  "error",
        }
    }
 
    if result.MatchedCount == 0 {
        return models.UserResponse{
            Message: "Group not found or user is not in the group",
            Status:  "error",
        }
    }
 
    return models.UserResponse{
        Message: "User removed from group successfully",
        Status:  "success",
    }
}
 
 
// DeleteUser deletes a user from the "users" collection
func DeleteUser(username string) models.UserResponse {
    // Check if the user exists
    count, err := GetUsersCollection().CountDocuments(context.Background(), bson.M{"username": username})
    if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("error checking user existence: %v", err),
            Status:  "error",
        }
    }
 
    // If user does not exist, return a specific message
    if count == 0 {
        return models.UserResponse{
            Message: "User does not exist",
            Status:  "error",
        }
    }
 
    // Attempt to delete the user
    _, err = GetUsersCollection().DeleteOne(context.Background(), bson.M{"username": username})
    if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("error deleting user: %v", err),
            Status:  "error",
        }
    }
 
    // If deletion is successful
    return models.UserResponse{
        Message: "User deleted successfully",
        Status:  "success",
    }
}
 
func ListGroupsByManager(manager string) models.UserResponse {
    filter := bson.M{"manager": manager}
 
    // Check if any groups exist for the manager
    count, err := GetGroupsCollection().CountDocuments(context.Background(), filter)
    if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("Error checking manager existence: %v", err),
            Status:  "error",
        }
    }
 
    if count == 0 {
        return models.UserResponse{
            Message: fmt.Sprintf("Manager '%s' not found or has no groups", manager),
            Status:  "error",
        }
    }
 
    // Query groups for the manager
    cursor, err := GetGroupsCollection().Find(context.Background(), filter)
    if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("Error retrieving groups: %v", err),
            Status:  "error",
        }
    }
    defer cursor.Close(context.Background())
 
    // Decode groups into the updated Group structure
    var groups []models.Group
    if err = cursor.All(context.Background(), &groups); err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("Error decoding groups: %v", err),
            Status:  "error",
        }
    }
 
    return models.UserResponse{
        Message: "Groups retrieved successfully",
        Status:  "success",
        Data:    groups,
    }
}
 
 
// AddBudget assigns a budget to a group by group ID
func AddBudget(manager, groupID string, budget float64) models.UserResponse {
    if budget <= 0 {
        return models.UserResponse{
            Message: "budget must be greater than zero",
            Status:  "error",
        }
    }
 
    // Check if the group exists for the given manager
    filter := bson.M{"manager": manager, "group_id": groupID}
    var group models.Group
    err := GetGroupsCollection().FindOne(context.Background(), filter).Decode(&group)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return models.UserResponse{
                Message: fmt.Sprintf("group with ID '%s' not found for manager '%s'", groupID, manager),
                Status:  "error",
            }
        }
        return models.UserResponse{
            Message: fmt.Sprintf("error finding group: %v", err),
            Status:  "error",
        }
    }
 
    // Check if the budget is already assigned to the group
    if group.Budget > 0 {
        return models.UserResponse{
            Message: fmt.Sprintf("budget is already allocated for the group '%s'", group.GroupName),
            Status:  "error",
        }
    }
 
    // Assign the new budget to the group
    update := bson.M{"$set": bson.M{"budget": budget}}
    _, err = GetGroupsCollection().UpdateOne(context.Background(), filter, update)
    if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("error updating budget: %v", err),
            Status:  "error",
        }
    }
 
    return models.UserResponse{
        Message: fmt.Sprintf("Budget successfully allocated to group '%s'", group.GroupName),
        Status:  "success",
    }
}
 
func UpdateBudget(manager, groupName string, budget float64) models.UserResponse {
    if budget <= 0 {
        return models.UserResponse{
            Message: "Budget must be greater than zero",
            Status:  "error",
        }
    }
 
    // Check if the group exists for the given manager
    filter := bson.M{"manager": manager, "group_name": groupName}
    var group models.Group
    err := GetGroupsCollection().FindOne(context.Background(), filter).Decode(&group)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return models.UserResponse{
                Message: fmt.Sprintf("Group '%s' not found for manager '%s'", groupName, manager),
                Status:  "error",
            }
        }
        return models.UserResponse{
            Message: fmt.Sprintf("Error finding group: %v", err),
            Status:  "error",
        }
    }
 
    // Check if the group already has a budget assigned
    if group.Budget == 0 {
        return models.UserResponse{
            Message: fmt.Sprintf("No budget allocated for group '%s'. Cannot update budget.", groupName),
            Status:  "error",
        }
    }
 
    // Update the existing budget in the group
    update := bson.M{"$set": bson.M{"budget": budget}}
    result, err := GetGroupsCollection().UpdateOne(context.Background(), filter, update)
    if err != nil {
        return models.UserResponse{
            Message: fmt.Sprintf("Error updating budget: %v", err),
            Status:  "error",
        }
    }
 
    if result.MatchedCount == 0 {
        return models.UserResponse{
            Message: fmt.Sprintf("Group '%s' not found for manager '%s'", groupName, manager),
            Status:  "error",
        }
    }
 
    return models.UserResponse{
        Message: fmt.Sprintf("Budget successfully updated for group '%s'", groupName),
        Status:  "success",
    }
}
 
// CheckUserGroup checks if a user is already a member of any group.
func CheckUserGroup(username string) models.UserResponse {
    // Define filter to check if the user is already in a group
    filter := bson.M{"members": username}
    var existingGroup bson.M
 
    // Query the "groups" collection to find if the user is a member of any group
    err := GetGroupsCollection().FindOne(context.Background(), filter).Decode(&existingGroup)
    if err == nil {
        // User is found in an existing group
        return models.UserResponse{
            Message: "User is already a member of another group and cannot be added to multiple groups",
            Status:  "error",
        }
    } else if err != mongo.ErrNoDocuments {
        // Handle any other error during the query
        return models.UserResponse{
            Message: fmt.Sprintf("Error checking user's group membership: %v", err),
            Status:  "error",
        }
    }
 
    // If no documents were found, the user is not in any group
    return models.UserResponse{
        Message: "User is not a member of any group",
        Status:  "success",
    }
}