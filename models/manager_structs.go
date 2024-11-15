package models
 
// Struct for groups
type Group struct {
    GroupID   string   `json:"group_id" bson:"group_id"`       
    Manager   string   `json:"manager" bson:"manager"`          // Manager username
    GroupName string   `json:"group_name" bson:"group_name"`    // Name of the group
    Members   []string `json:"members" bson:"members"`          // List of group members
    Budget    float64  `json:"budget,omitempty" bson:"budget"`  // Optional budget field
}
 
// Response structure
type UserResponse struct {
    Status  string      `json:"status,omitempty"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}