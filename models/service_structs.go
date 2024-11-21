package models

// Service represents a cloud service document in MongoDB
type Service struct {
	UserEmail     string  `bson:"user_email" json:"user_email"`         // Email of the user owning the service
	CloudProvider string  `bson:"cloud_provider" json:"cloud_provider"` // Cloud provider (e.g., AWS, GCP, Azure)
	ServiceName   string  `bson:"service_name" json:"service_name"`     // Name of the service
	Name          string  `bson:"name" json:"name"`                     // Name of the service
	Provider      string  `bson:"provider" json:"provider"`             // Cloud provider name
	Cost          float64 `bson:"cost" json:"cost"`                     // Cost of the service
	Usage         float64 `bson:"usage" json:"usage"`                   // Current usage of the service
}
