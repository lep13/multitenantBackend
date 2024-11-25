package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"multitenant/cloud"
	"multitenant/db"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

// Handler for creating EC2 instance
func CreateEC2InstanceHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        SessionID       string `json:"session_id"`
        InstanceType    string `json:"instance_type"`
        AmiID           string `json:"ami_id"`
        KeyName         string `json:"key_name"`
        SubnetID        string `json:"subnet_id"`
        SecurityGroupID string `json:"security_group_id"`
        InstanceName    string `json:"instance_name"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Fetch session details
    var session bson.M
    err := db.GetUserSessionCollection().FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&session)
    if err != nil {
        http.Error(w, "Session not found", http.StatusNotFound)
        return
    }

    // Check if the session is approved
    status := session["status"].(string)
    if status != "ok" {
        http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
        return
    }

    // Create the EC2 instance
    result, err := cloud.CreateEC2Instance(req.InstanceType, req.AmiID, req.KeyName, req.SubnetID, req.SecurityGroupID, req.InstanceName)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to create EC2 instance: %v", err), http.StatusInternalServerError)
        return
    }

    // Extract the instance ID
    if len(result.Instances) == 0 {
        http.Error(w, "Failed to fetch instance ID", http.StatusInternalServerError)
        return
    }
    instanceID := *result.Instances[0].InstanceId

    // Update the configuration with the instance ID
    config := bson.M{
        "instance_type":    req.InstanceType,
        "ami_id":           req.AmiID,
        "key_name":         req.KeyName,
        "subnet_id":        req.SubnetID,
        "security_group_id": req.SecurityGroupID,
        "instance_name":    req.InstanceName,
        "instance_id":      instanceID, // Add instance ID here
    }

    // Store configuration in the user_sessions collection
    filter := bson.M{"session_id": req.SessionID}
    update := bson.M{"$set": bson.M{"config": config}}

    _, err = db.GetUserSessionCollection().UpdateOne(context.Background(), filter, update)
    if err != nil {
        http.Error(w, "Failed to store configuration in user_sessions", http.StatusInternalServerError)
        return
    }

    // Respond with the creation result
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message":     "EC2 instance created successfully",
        "result":      result,
        "config":      config,
        "instance_id": instanceID,
    })
}

// CreateS3BucketHandler handles requests to create an S3 bucket
func CreateS3BucketHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID  string `json:"session_id"`
		BucketName string `json:"bucket_name"`
		Versioning bool   `json:"versioning"`
		Region     string `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Fetch session details
	var session bson.M
	err := db.GetUserSessionCollection().FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&session)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Check if the session is approved
	status := session["status"].(string)
	if status != "ok" {
		http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
		return
	}

	// Configuration details
	config := bson.M{
		"bucket_name":    req.BucketName,
		"versioning":     req.Versioning,
		"region":         req.Region,
	}

	// Proceed with service creation
	result, err := cloud.CreateS3Bucket(req.BucketName, req.Versioning, req.Region)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create S3 bucket: %v", err), http.StatusInternalServerError)
		return
	}

	// Store configuration in the user_sessions collection
	filter := bson.M{"session_id": req.SessionID}
	update := bson.M{"$set": bson.M{"config": config}}

	_, err = db.GetUserSessionCollection().UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, "Failed to store configuration in user_sessions", http.StatusInternalServerError)
		return
	}

	// Respond with the creation result (service creation successful)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "S3 bucket created successfully",
		"result":  result,
		"config":  config,
	})
}

// CreateLambdaFunctionHandler handles requests to create a Lambda function
func CreateLambdaFunctionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID    string `json:"session_id"`
		FunctionName string `json:"function_name"`
		Handler      string `json:"handler"`
		Runtime      string `json:"runtime"`
		ZipFilePath  string `json:"zip_file_path"`
		Region       string `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Fetch session details
	var session bson.M
	err := db.GetUserSessionCollection().FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&session)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Check if the session is approved
	status := session["status"].(string)
	if status != "ok" {
		http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
		return
	}

	// Configuration details
	config := bson.M{
		"function_name": req.FunctionName,
		"handler":       req.Handler,
		"runtime":       req.Runtime,
		"zip_file_path": req.ZipFilePath,
		"region":        req.Region,
	}

	// Proceed with service creation
	result, err := cloud.CreateLambdaFunction(req.FunctionName, req.Handler, req.Runtime, req.ZipFilePath, req.Region)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Lambda function: %v", err), http.StatusInternalServerError)
		return
	}

	// Store configuration in the user_sessions collection
	filter := bson.M{"session_id": req.SessionID}
	update := bson.M{"$set": bson.M{"config": config}}

	_, err = db.GetUserSessionCollection().UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, "Failed to store configuration in user_sessions", http.StatusInternalServerError)
		return
	}

	// Respond with the creation result (service creation successful)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Lambda function created successfully",
		"result":       result,
		"config":       config,
		"function_name": req.FunctionName,
		"region":       req.Region,
	})

}

// Handler for creating RDS instance
func CreateRDSInstanceHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID        string `json:"session_id"`
		DBName           string `json:"db_name"`
		InstanceID       string `json:"instance_id"`
		InstanceClass    string `json:"instance_class"`
		Engine           string `json:"engine"`
		Username         string `json:"username"`
		Password         string `json:"password"`
		AllocatedStorage int32  `json:"allocated_storage"`
		SubnetGroupName  string `json:"subnet_group_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Fetch session details
	var session bson.M
	err := db.GetUserSessionCollection().FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&session)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Check if the session is approved
	status := session["status"].(string)
	if status != "ok" {
		http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
		return
	}

	// Configuration details
	config := bson.M{
		"db_name":            req.DBName,
		"instance_id":        req.InstanceID,
		"instance_class":     req.InstanceClass,
		"engine":             req.Engine,
		"username":           req.Username,
		"password":           req.Password,
		"allocated_storage":  req.AllocatedStorage,
		"subnet_group_name":  req.SubnetGroupName,
	}

	// Proceed with RDS instance creation
	result, err := cloud.CreateRDSInstance(req.DBName, req.InstanceID, req.InstanceClass, req.Engine, req.Username, req.Password, req.AllocatedStorage, req.SubnetGroupName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not create RDS instance: %v", err), http.StatusInternalServerError)
		return
	}

	// Store configuration in the user_sessions collection
	filter := bson.M{"session_id": req.SessionID}
	update := bson.M{"$set": bson.M{"config": config}}

	_, err = db.GetUserSessionCollection().UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, "Failed to store configuration in user_sessions", http.StatusInternalServerError)
		return
	}

	// Respond with the creation result (service creation successful)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "RDS instance created successfully",
		"result":         result,
		"config":         config,
		"db_instance_id": req.InstanceID,
	})

}

// // Handler for creating DynamoDB table
// func CreateDynamoDBTableHandler(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		SessionID     string `json:"session_id"`
// 		TableName     string `json:"table_name"`
// 		Region        string `json:"region"` // Region as input
// 		ReadCapacity  int64  `json:"read_capacity"`
// 		WriteCapacity int64  `json:"write_capacity"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid input", http.StatusBadRequest)
// 		return
// 	}

// 	if req.Region == "" {
// 		http.Error(w, "Region is required", http.StatusBadRequest)
// 		return
// 	}

// 	// Fetch session details
// 	var session bson.M
// 	err := db.GetUserSessionCollection().FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&session)
// 	if err != nil {
// 		http.Error(w, "Session not found", http.StatusNotFound)
// 		return
// 	}

// 	// Check if the session is approved
// 	status, ok := session["status"].(string)
// 	if !ok || status != "ok" {
// 		http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
// 		return
// 	}

// 	// Proceed with DynamoDB table creation
// 	result, err := cloud.CreateDynamoDBTable(req.TableName, req.Region, req.ReadCapacity, req.WriteCapacity)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Could not create DynamoDB table: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	// Finalize the session
// 	completeReq := struct {
// 		SessionID string `json:"session_id"`
// 		Status    string `json:"status"`
// 	}{
// 		SessionID: req.SessionID,
// 		Status:    "completed",
// 	}

// 	completeData, _ := json.Marshal(completeReq)
// 	http.Post("http://localhost:8080/complete-session", "application/json", bytes.NewReader(completeData))

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(result)
// }

// Handler for Creating CloudFront Distribution
func CreateCloudFrontDistributionHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        SessionID        string `json:"session_id"`
        OriginDomainName string `json:"origin_domain_name"`
        Comment          string `json:"comment"`
        Region           string `json:"region"`
        MinTTL           int64  `json:"min_ttl"`
        BucketName       string `json:"bucket_name"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    // Fetch session details
    var session bson.M
    err := db.GetUserSessionCollection().FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&session)
    if err != nil {
        http.Error(w, "Session not found", http.StatusNotFound)
        return
    }

    // Check if the session is approved
    status := session["status"].(string)
    if status != "ok" {
        http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
        return
    }

    // Create CloudFront distribution
    distributionResult, oaiCanonicalUserID, err := cloud.CreateCloudFrontDistribution(req.OriginDomainName, req.Comment, req.Region, req.MinTTL)
    if err != nil {
        http.Error(w, fmt.Sprintf("Could not create CloudFront distribution: %v", err), http.StatusInternalServerError)
        return
    }

    // Extract the Distribution ID
    distributionID := *distributionResult.Distribution.Id

    // Create the S3 Bucket and attach the policy
    _, err = cloud.CreateS3BucketWithPolicy(req.BucketName, req.Region, oaiCanonicalUserID)
    if err != nil {
        http.Error(w, fmt.Sprintf("Could not create S3 bucket or attach policy: %v", err), http.StatusInternalServerError)
        return
    }

    // Store configuration in the user_sessions collection
    config := bson.M{
        "origin_domain_name": req.OriginDomainName,
        "comment":            req.Comment,
        "region":             req.Region,
        "min_ttl":            req.MinTTL,
        "bucket_name":        req.BucketName,
        "distribution_id":    distributionID, // Store the distribution ID
    }

    filter := bson.M{"session_id": req.SessionID}
    update := bson.M{"$set": bson.M{"config": config}}

    _, err = db.GetUserSessionCollection().UpdateOne(context.Background(), filter, update)
    if err != nil {
        http.Error(w, "Failed to store configuration in user_sessions", http.StatusInternalServerError)
        return
    }

    // Respond with the creation result (service creation successful)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message":           "CloudFront distribution created successfully",
        "result":            distributionResult,
        "config":            config,
        "origin_domain_name": req.OriginDomainName,
        "region":            req.Region,
    })
}

// Handler for creating VPC with session management
func CreateVPCHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
		CidrBlock string `json:"cidr_block"`
		Region    string `json:"region"`
		Name      string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" || req.Region == "" || req.CidrBlock == "" || req.Name == "" {
		http.Error(w, "Session ID, Region, CIDR block, and Name are required fields", http.StatusBadRequest)
		return
	}

	// Fetch session details
	var session bson.M
	err := db.GetUserSessionCollection().FindOne(context.Background(), bson.M{"session_id": req.SessionID}).Decode(&session)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Check if the session is approved
	status := session["status"].(string)
	if status != "ok" {
		http.Error(w, "Session is not approved for service creation", http.StatusForbidden)
		return
	}

	// Configuration details
	config := bson.M{
		"cidr_block": req.CidrBlock,
		"region":     req.Region,
		"name":       req.Name,
	}

	// Proceed with VPC creation
	result, err := cloud.CreateVPC(req.CidrBlock, req.Region, req.Name)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not create VPC: %v", err), http.StatusInternalServerError)
		return
	}

	// Store configuration in the user_sessions collection
	filter := bson.M{"session_id": req.SessionID}
	update := bson.M{"$set": bson.M{"config": config}}

	_, err = db.GetUserSessionCollection().UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, "Failed to store configuration in user_sessions", http.StatusInternalServerError)
		return
	}

	// Respond with the creation result (service creation successful)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "VPC created successfully",
		"result":    result,
		"config":    config,
		"cidr_block": req.CidrBlock,
		"region":    req.Region,
	})

}

// DeleteAWSServiceHandler handles AWS service deletions
func DeleteAWSServiceHandler(w http.ResponseWriter, r *http.Request) {
    username := r.URL.Query().Get("username")

    var req struct {
        ServiceType string `json:"service_type"`
        ServiceName string `json:"service_name"` // For all services except CloudFront
        ServiceID   string `json:"service_id"`   // For CloudFront only
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    if username == "" || req.ServiceType == "" || (req.ServiceType == "AWS CloudFront" && req.ServiceID == "") || (req.ServiceType != "AWS CloudFront" && req.ServiceName == "") {
        http.Error(w, "Missing required fields", http.StatusBadRequest)
        return
    }

    var instanceID string
    var result interface{}
    var err error
    var message string
    var shouldUpdateStatus bool = false // Default: do not update DB unless explicitly set

    // Fetch the instance ID for services requiring it (EC2, RDS)
    if req.ServiceType == "Amazon EC2 (Elastic Compute Cloud)" || req.ServiceType == "Amazon RDS (Relational Database Service)" {
        instanceID, err = db.GetInstanceIDByInstanceName(username, req.ServiceType, req.ServiceName)
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to fetch instance ID: %v", err), http.StatusInternalServerError)
            return
        }
    }

    // Call the appropriate deletion function based on the service type
    switch req.ServiceType {
    case "AWS Lambda":
        result, err = cloud.DeleteLambdaFunction(req.ServiceName)
        if err == nil {
            shouldUpdateStatus = true
            message = "Service deleted successfully"
        }
    case "Amazon EC2 (Elastic Compute Cloud)":
        result, err = cloud.TerminateEC2Instance(instanceID) // Use instanceID for EC2
        if err == nil {
            shouldUpdateStatus = true
            message = "Service deleted successfully"
        }
    case "Amazon S3 (Simple Storage Service)":
        result, err = cloud.DeleteS3Bucket(req.ServiceName)
        if err == nil {
            shouldUpdateStatus = true
            message = "Service deleted successfully"
        }
    case "Amazon RDS (Relational Database Service)":
        result, err = cloud.DeleteRDSInstance(instanceID) // Use instanceID for RDS
        if err == nil {
            shouldUpdateStatus = true
            message = "Service deleted successfully"
        }
    case "AWS CloudFront":
        result, err = cloud.DisableCloudFrontDistribution(req.ServiceID) // Use ServiceID for CloudFront
        if err == nil {
            shouldUpdateStatus = true
            message = "Service disabled successfully"
        }
    case "Amazon VPC (Virtual Private Cloud)":
        result, message, err = cloud.DeleteVPC(req.ServiceName)
        if err == nil && message == "deleted" {
            err = db.UpdateawsServiceStatus(username, req.ServiceType, req.ServiceName, "deleted")
            if err != nil {
                http.Error(w, fmt.Sprintf("Failed to update service status: %v", err), http.StatusInternalServerError)
                return
            }
        }
    default:
        http.Error(w, "Invalid service type", http.StatusBadRequest)
        return
    }

    if err != nil && req.ServiceType != "Amazon VPC (Virtual Private Cloud)" {
        // If deletion fails for non-VPC services, return the error
        http.Error(w, fmt.Sprintf("Failed to delete service: %v", err), http.StatusInternalServerError)
        return
    }

    if shouldUpdateStatus {
        // Update service status in the database only if deletion was successful
        updateIdentifier := req.ServiceName
        if req.ServiceType == "AWS CloudFront" {
            updateIdentifier = req.ServiceID // Use ServiceID for CloudFront in the update function
        }

        err = db.UpdateawsServiceStatus(username, req.ServiceType, updateIdentifier, "deleted")
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to update service status: %v", err), http.StatusInternalServerError)
            return
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": message,
        "result":  result,
    })
}
