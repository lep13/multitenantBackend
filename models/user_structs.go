package models

import "time"

// ServiceRequest represents a cloud service requested by a user
type ServiceRequest struct {
	ServiceID      string                 `json:"service_id" bson:"service_id"`         // Unique identifier for the service
	UserID         string                 `json:"user_id" bson:"user_id"`               // ID of the user requesting the service
	GroupName      string                 `json:"group_name" bson:"group_name"`         // Group to which the user belongs
	CloudProvider  string                 `json:"cloud_provider" bson:"cloud_provider"` // AWS or GCP
	ServiceType    string                 `json:"service_type" bson:"service_type"`     // Type of service (e.g., EC2, S3, Compute Engine)
	Config         map[string]interface{} `json:"config" bson:"config"`                 // Configuration details (e.g., instance type, storage)
	Status         string                 `json:"status" bson:"status"`                 // Status of the service (e.g., running, terminated)
	StartDate      time.Time              `json:"start_date" bson:"start_date"`         // When the service was created
	EndDate        time.Time              `json:"end_date" bson:"end_date"`             // When the service is scheduled to end
	Cost           float64                `json:"cost" bson:"cost"`                     // Total cost for the service duration
	Duration       int                    `json:"duration" bson:"duration"`             // Duration in hours or days (based on configuration)
}

// Budget represents the budget details for a group
type Budget struct {
	GroupName      string  `json:"group_name" bson:"group_name"`             // Group using this budget
	TotalBudget    float64 `json:"total_budget" bson:"total_budget"`         // Total budget allocated to the group
	BudgetUsed     float64 `json:"budget_used" bson:"budget_used"`           // Amount of budget used
	BudgetRemaining float64 `json:"budget_remaining" bson:"budget_remaining"` // Remaining budget after usage
	Alerts         bool    `json:"alerts" bson:"alerts"`                     // Alerts if budget is nearing limits
}

// ServiceHistory represents historical data of services
type ServiceHistory struct {
	ServiceID    string    `json:"service_id" bson:"service_id"`       // Unique identifier for the service
	UserID       string    `json:"user_id" bson:"user_id"`             // ID of the user who requested the service
	GroupName    string    `json:"group_name" bson:"group_name"`       // Group to which the user belongs
	CloudProvider string   `json:"cloud_provider" bson:"cloud_provider"` // AWS or GCP
	ServiceType  string    `json:"service_type" bson:"service_type"`   // Type of service (e.g., EC2, S3)
	Config       map[string]interface{} `json:"config" bson:"config"`  // Configuration details
	Status       string    `json:"status" bson:"status"`               // Status of the service (e.g., running, terminated)
	StartDate    time.Time `json:"start_date" bson:"start_date"`       // When the service was created
	EndDate      time.Time `json:"end_date" bson:"end_date"`           // When the service ended
	Cost         float64   `json:"cost" bson:"cost"`                   // Total cost of the service
}

// BudgetUpdateResponse represents a response structure when updating the budget
type BudgetUpdateResponse struct {
	Status  string `json:"status"`             // Status of the update (success or error)
	Message string `json:"message"`            // Message detailing the update result
	Updated bool   `json:"updated,omitempty"`  // Indicates if the budget was successfully updated
}

// ServiceResponse represents a response structure for service-related actions
type ServiceResponse struct {
	Status  string      `json:"status"`               // Status of the action (success or error)
	Message string      `json:"message"`              // Message detailing the result
	Data    interface{} `json:"data,omitempty"`       // Additional data (e.g., service details)
}

// BudgetResponse represents a response structure for budget-related actions
type BudgetResponse struct {
	Status        string  `json:"status"`               // Status of the action (success or error)
	Message       string  `json:"message"`              // Message detailing the result
	TotalBudget   float64 `json:"total_budget"`         // Total budget allocated to the group
	BudgetUsed    float64 `json:"budget_used"`          // Amount of budget used
	BudgetRemaining float64 `json:"budget_remaining"`   // Remaining budget
	Alerts        bool    `json:"alerts,omitempty"`     // Alerts status for nearing budget limits
}