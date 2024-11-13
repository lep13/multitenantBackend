package handlers

import (
	"encoding/json"
	"fmt"
	"multitenant/cloud"
	"net/http"
)

// Handler for creating EC2 instance
func CreateEC2InstanceHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InstanceType    string `json:"instance_type"`
		AmiID           string `json:"ami_id"`
		KeyName         string `json:"key_name"`
		SubnetID        string `json:"subnet_id"`
		SecurityGroupID string `json:"security_group_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	result, err := cloud.CreateEC2Instance(req.InstanceType, req.AmiID, req.KeyName, req.SubnetID, req.SecurityGroupID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not create EC2 instance: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// S3BucketRequest represents the expected request body for creating an S3 bucket
type S3BucketRequest struct {
	BucketName string `json:"bucket_name"`
	Versioning bool   `json:"versioning"`
	Region     string `json:"region"`
}

// CreateS3BucketHandler handles requests to create an S3 bucket
func CreateS3BucketHandler(w http.ResponseWriter, r *http.Request) {
	var req S3BucketRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	result, err := cloud.CreateS3Bucket(req.BucketName, req.Versioning, req.Region)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not create S3 bucket: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Handler for creating Lambda function
func CreateLambdaFunctionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
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

	// Call the updated cloud function with a hardcoded role
	result, err := cloud.CreateLambdaFunction(req.FunctionName, req.Handler, req.Runtime, req.ZipFilePath, req.Region)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not create Lambda function: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// // Handler for creating RDS instance
// func CreateRDSInstanceHandler(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		DBName           string `json:"db_name"`
// 		InstanceID       string `json:"instance_id"`
// 		InstanceClass    string `json:"instance_class"`
// 		Engine           string `json:"engine"`
// 		Username         string `json:"username"`
// 		Password         string `json:"password"`
// 		AllocatedStorage int32  `json:"allocated_storage"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid input", http.StatusBadRequest)
// 		return
// 	}

// 	result, err := cloud.CreateRDSInstance(req.DBName, req.InstanceID, req.InstanceClass, req.Engine, req.Username, req.Password, req.AllocatedStorage)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Could not create RDS instance: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(result)
// }

// // Handler for creating DynamoDB table
// func CreateDynamoDBTableHandler(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		TableName     string `json:"table_name"`
// 		ReadCapacity  int64  `json:"read_capacity"`
// 		WriteCapacity int64  `json:"write_capacity"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid input", http.StatusBadRequest)
// 		return
// 	}

// 	result, err := cloud.CreateDynamoDBTable(req.TableName, req.ReadCapacity, req.WriteCapacity)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Could not create DynamoDB table: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(result)
// }

// // Handler for creating CloudFront distribution
// func CreateCloudFrontDistributionHandler(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		OriginDomainName string `json:"origin_domain_name"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid input", http.StatusBadRequest)
// 		return
// 	}

// 	result, err := cloud.CreateCloudFrontDistribution(req.OriginDomainName)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Could not create CloudFront distribution: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(result)
// }

// // Handler for creating VPC
// func CreateVPCHandler(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		CidrBlock string `json:"cidr_block"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid input", http.StatusBadRequest)
// 		return
// 	}

// 	result, err := cloud.CreateVPC(req.CidrBlock)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Could not create VPC: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(result)
// }

// // GetDropdownData serves dropdown data for the UI
// func GetDropdownData(w http.ResponseWriter, r *http.Request) {
// 	// Load the JSON file
// 	file, err := os.Open("terraform_outputs.json")
// 	if err != nil {
// 		http.Error(w, "Failed to load dropdown data", http.StatusInternalServerError)
// 		return
// 	}
// 	defer file.Close()

// 	// Parse the JSON file
// 	var data map[string]interface{}
// 	if err := json.NewDecoder(file).Decode(&data); err != nil {
// 		http.Error(w, "Failed to parse dropdown data", http.StatusInternalServerError)
// 		return
// 	}

// 	// Add hardcoded Lambda runtimes
// 	data["lambda_supported_runtimes"] = []string{
// 		"nodejs18.x", "python3.9", "go1.x", "java11", "ruby2.7", "dotnet6",
// 	}

// 	// Send the combined response
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(data)
// }
