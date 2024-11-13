package models

// GCPInstanceRequest represents the input for creating a Compute Engine instance
type GCPInstanceRequest struct {
	Name           string `json:"name"`             // Instance name
	ProjectID      string `json:"project_id"`       // Project ID
	Zone           string `json:"zone"`            // Zone for the instance
	MachineType    string `json:"machine_type"`     // Machine type
	ImageProject   string `json:"image_project"`    // Project where the image is hosted
	ImageFamily    string `json:"image_family"`     // Image family
	Network        string `json:"network"`          // Network name
	Subnetwork     string `json:"subnetwork"`       // Subnetwork name
	ServiceAccount string `json:"service_account"`  // Service account email
	Region         string `json:"region"`           // Region for subnetwork
}
