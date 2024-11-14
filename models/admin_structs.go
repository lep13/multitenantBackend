package models

// Manager represents the structure of a manager document in the "managers" collection
type Manager struct {
    Username   string `bson:"username"`
    Email      string `bson:"email"`
    GroupLimit int    `bson:"group_limit"`
}

// CreateManagerRequest represents the input required to create a new manager
type CreateManagerRequest struct {
    Username   string `json:"username"`
    Password   string `json:"password"`
    Email      string `json:"email"`
    GroupLimit int    `json:"group_limit"`
}

// CreateManagerResponse represents the structure of the response after creating a manager
type ManagerResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

type RemoveManagerRequest struct {
	Username string `json:"username"`
}