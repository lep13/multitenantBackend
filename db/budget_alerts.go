package db

import (
	"context"
	"fmt"
	"multitenant/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TrackServiceChanges compares the old and new service details, and notifies the manager if changes are detected.
func TrackServiceChanges(manager, serviceID string, oldService, newService models.Service) models.UserResponse {
	// Compare the old and new service details and collect the changes.
	var changes []string

	// Example comparison: Check for changes in service fields (service name, cloud provider, etc.)
	if oldService.Name != newService.Name {
		changes = append(changes, fmt.Sprintf("Service Name changed from '%s' to '%s'", oldService.Name, newService.Name))
	}
    
	if oldService.Provider != newService.Provider {
		changes = append(changes, fmt.Sprintf("Cloud Provider changed from '%s' to '%s'", oldService.Provider, newService.Provider))
	}

	if oldService.Cost != newService.Cost {
		changes = append(changes, fmt.Sprintf("Cost changed from $%.2f to $%.2f", oldService.Cost, newService.Cost))
	}

	// Add more fields as needed, for example, usage, region, etc.

	// If there are any changes, notify the manager
	if len(changes) > 0 {
		changeMessage := fmt.Sprintf("The following changes were made to service '%s':\n", serviceID)
		for _, change := range changes {
			changeMessage += fmt.Sprintf("- %s\n", change)
		}
		changeMessage += "\nPlease review these changes.\n\nThank you for using our platform."

		// Send email notification to the manager
		err := sendEmailAlert(manager, changeMessage)
		if err != nil {
			return models.UserResponse{
				Message: fmt.Sprintf("Failed to send email notification to manager: %v", err),
				Status:  "error",
			}
		}
	}

	// Return success response
	return models.UserResponse{
		Message: "Service changes tracked and manager notified successfully",
		Status:  "success",
	}
}

// LogServiceCreation logs a new service creation and sends an email notification to the user
func LogServiceCreation(userEmail, cloudProvider, serviceName string) models.UserResponse {
	// Log the service creation
	logMessage := fmt.Sprintf("User created a new service: Provider=%s, Service=%s", cloudProvider, serviceName)
	fmt.Println(logMessage)

	// Send email notification to the user
	alertMessage := fmt.Sprintf(
		"Hello,\n\nYou have successfully created a new service on %s: %s.\n\nThank you for using our platform.",
		cloudProvider, serviceName,
	)
	err := sendEmailAlert(userEmail, alertMessage)
	if err != nil {
		return models.UserResponse{
			Message: fmt.Sprintf("Failed to send email notification: %v", err),
			Status:  "error",
		}
	}

	return models.UserResponse{
		Message: "Service creation logged and email notification sent successfully",
		Status:  "success",
	}
}

// LogServiceDeletion logs a service deletion and sends an email notification to the user
func LogServiceDeletion(userEmail, cloudProvider, serviceID string) models.UserResponse {
	// Log the service deletion
	logMessage := fmt.Sprintf("User deleted a service: Provider=%s, ServiceID=%s", cloudProvider, serviceID)
	fmt.Println(logMessage)

	// Send email notification to the user
	alertMessage := fmt.Sprintf(
		"Hello,\n\nYou have successfully deleted the service with ID: %s on %s.\n\nThank you for using our platform.",
		serviceID, cloudProvider,
	)
	err := sendEmailAlert(userEmail, alertMessage)
	if err != nil {
		return models.UserResponse{
			Message: fmt.Sprintf("Failed to send email notification: %v", err),
			Status:  "error",
		}
	}

	return models.UserResponse{
		Message: "Service deletion logged and email notification sent successfully",
		Status:  "success",
	}
}

// CheckBudgetUsage checks if the usage exceeds 75% of the allocated budget and sends an alert
func CheckBudgetUsage(manager, groupName string, usedAmount float64) models.UserResponse {
	// Retrieve the group to get the allocated budget
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

	if group.Budget == 0 {
		return models.UserResponse{
			Message: "No budget allocated for this group",
			Status:  "error",
		}
	}

	// Check if the usage has reached 75% of the budget
	threshold := 0.75 * group.Budget
	if usedAmount >= threshold {
		alertMessage := fmt.Sprintf("Alert: Group '%s' has used 75%% of its allocated budget of $%.2f. Current usage: $%.2f", groupName, group.Budget, usedAmount)
		sendEmailAlert(manager, alertMessage)
	}

	return models.UserResponse{
		Message: "Budget usage checked successfully",
		Status:  "success",
	}
}

// StartBudgetMonitoring periodically checks budget usage for each group
func StartBudgetMonitoring(manager, groupName string) {
	ticker := time.NewTicker(24 * time.Hour) // Check daily
	defer ticker.Stop()

	for {
		<-ticker.C
		usedAmount := getCurrentUsageForGroup(manager, groupName) // Define a function to calculate actual usage
		CheckBudgetUsage(manager, groupName, usedAmount)
	}
}

// getCurrentUsageForGroup calculates the current usage for a group
func getCurrentUsageForGroup(manager, groupName string) float64 {
	// Stub: Replace with actual logic to get current usage for the group
	return 800.0 // Example usage value
}

// QuarterlyUsageReport sends a report for all active services for a user
func QuarterlyUsageReport(userEmail string, services []models.Service) {
	message := "Hello,\n\nHere is your quarterly usage report:\n\n"
	totalCost := 0.0
	for _, service := range services {
		message += fmt.Sprintf("- %s (%s): $%.2f\n", service.Name, service.Provider, service.Cost)
		totalCost += service.Cost
	}
	message += fmt.Sprintf("\nTotal cost this quarter: $%.2f\n\nThank you for using our platform.", totalCost)

	sendEmailAlert(userEmail, message)
}

// NotifyServiceFailure sends an email alert when a service fails
func NotifyServiceFailure(userEmail, serviceName, provider string) {
	message := fmt.Sprintf(
		"Hello,\n\nWe have detected a failure in your service '%s' on %s. Please check your service or contact support for assistance.\n\nThank you.",
		serviceName, provider,
	)
	sendEmailAlert(userEmail, message)
}

// sendEmailAlert sends an email alert
func sendEmailAlert(email, message string) error {
	// Implement the email sending logic here (e.g., using an SMTP client or email API)
	fmt.Printf("Sending email to %s: %s\n", email, message)
	return nil // Replace with actual email sending logic
}
