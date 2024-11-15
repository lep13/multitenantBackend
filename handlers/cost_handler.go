package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"multitenant/db"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/aws/aws-sdk-go/aws"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/option"
)

// CalculateCostHandler calculates the estimated cost of a service
func CalculateCostHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
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

	// Get service, provider, and budget from session
	service, serviceOk := session["service"].(string)
	provider, providerOk := session["provider"].(string)
	budget, budgetOk := session["group_budget"].(float64)
	if !serviceOk || !providerOk || !budgetOk {
		http.Error(w, "Invalid session data", http.StatusInternalServerError)
		return
	}

	// Calculate cost
	var estimatedCost float64
	if provider == "aws" {
		estimatedCost, err = CalculateAWSCost(service)
	} else if provider == "gcp" {
		estimatedCost, err = CalculateGCPCost(service)
	} else {
		http.Error(w, "Unsupported cloud provider", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to calculate cost: %v", err), http.StatusInternalServerError)
		return
	}

	// Compare the estimated cost with the budget
	var status string
	if estimatedCost > budget {
		status = "denied"
	} else {
		status = "ok"
	}

	// Update session with estimated cost and status
	_, err = db.UpdateSessionWithCost(req.SessionID, estimatedCost, status)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update session: %v", err), http.StatusBadRequest)
		return
	}

	// Respond with the status and estimated cost
	response := map[string]interface{}{
		"status":         status,
		"estimated_cost": estimatedCost,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CalculateAWSCost calculates the quarterly cost for an AWS service based on its pricing unit.
func CalculateAWSCost(service string) (float64, error) {
	// Fetch the hourly price for the service.
	pricePerUnit, err := fetchAWSServicePrice(service)
	if err != nil {
		return 0, err
	}

	var estimatedCost float64

	// Calculate cost based on service-specific units.
	switch service {
	case "Amazon EC2 (Elastic Compute Cloud)":
		// EC2: Unit is USD/hour
		hoursPerQuarter := float64(24 * 90) // 24 hours/day * 90 days
		estimatedCost = pricePerUnit * hoursPerQuarter

	case "Amazon S3 (Simple Storage Service)":
		// S3: Unit is USD/GB/month
		monthlyCost := pricePerUnit * 1024 // assuming 1TB = 1024GB
		estimatedCost = monthlyCost * 3    // Quarterly cost (3 months)

	case "AWS Lambda":
		invocationsPerMonth := float64(1_000_000)   // 1 million invocations per month as float64
		memoryAllocatedGB := 0.125                 // 128 MB memory allocated, converted to GB
		secondsPerInvocation := float64(2)         // 2 seconds per invocation as float64
	
		pricePerGBSecond := pricePerUnit // Fetched using fetchAWSServicePrice
		monthlyCost := pricePerGBSecond * memoryAllocatedGB * secondsPerInvocation * invocationsPerMonth

		estimatedCost = monthlyCost * 3 // Multiply by 3 for quarterly cost
		estimatedCost = math.Round(estimatedCost*100) / 100	

	case "AWS CloudFront":
		// CloudFront: Calculate cost for requests and data transfer
		requestsPerMonth := float64(1_000_000) // assuming 1 million requests per month
		dataTransferPerMonth := float64(100)   // assuming 100 GB data transfer per month
		// CloudFront request pricing
		requestPricePerUnit := 0.0075 / 10_000 // $0.0075 per 10,000 requests
		requestCostPerMonth := (requestsPerMonth * requestPricePerUnit)
		// CloudFront data transfer pricing
		dataTransferPricePerUnit := 0.085 // $0.085 per GB for data transfer
		dataTransferCostPerMonth := dataTransferPerMonth * dataTransferPricePerUnit
		// S3 storage pricing
		s3StoragePerGB := 0.023 // $0.023 per GB for S3 Standard
		s3StorageSize := float64(100) // Assume 100 GB stored
		s3StorageCost := s3StorageSize * s3StoragePerGB
		// S3 GET request costs
		s3GetRequestPricePerUnit := 0.0004 // $0.0004 per 1,000 GET requests
		s3GetRequestCost := (requestsPerMonth / 1_000) * s3GetRequestPricePerUnit
		// Total monthly cost
		monthlyCost := requestCostPerMonth + dataTransferCostPerMonth + s3StorageCost + s3GetRequestCost
		
		estimatedCost = monthlyCost * 3 // Multiply by 3 for quarterly cost
		estimatedCost = math.Round(estimatedCost*100) / 100	

	case "Amazon RDS (Relational Database Service)":
		// RDS: Unit is USD/hour
		hoursPerQuarter := float64(24 * 90) // 24 hours/day * 90 days
		estimatedCost = pricePerUnit * hoursPerQuarter

	// case "Amazon DynamoDB":
	// 	readCostPerMillion := 0.25   // $0.25 per million read requests
	// 	writeCostPerMillion := 1.25 // $1.25 per million write requests
	
	// 	readCapacity := float64(5) // Default or fetched user input 
	// 	writeCapacity := float64(5) // Default or fetched user input
	
	// 	secondsPerMonth := float64(24 * 60 * 60 * 30) // 30 days
	// 	totalReadRequests := readCapacity * secondsPerMonth
	// 	totalWriteRequests := writeCapacity * secondsPerMonth
	
	// 	readCost := (totalReadRequests / 1_000_000) * readCostPerMillion
	// 	writeCost := (totalWriteRequests / 1_000_000) * writeCostPerMillion
	
	// 	monthlyCost := readCost + writeCost
	// 	estimatedCost = monthlyCost * 3
	// 	estimatedCost = math.Round(estimatedCost*100) / 100

	case "Amazon VPC (Virtual Private Cloud)":
		// VPC Endpoint: Unit is USD/hour
		hoursPerQuarter := float64(24 * 90) // 24 hours/day * 90 days
		estimatedCost = pricePerUnit * hoursPerQuarter

	default:
		return 0, fmt.Errorf("unsupported AWS service: %s", service)
	}

	// Round to 2 decimal places for clarity.
	estimatedCost = math.Round(estimatedCost*100) / 100
	return estimatedCost, nil
}

func fetchAWSServicePrice(service string) (float64, error) {
	// Load AWS configuration (pricing data is only available in us-east-1 region)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		return 0, fmt.Errorf("failed to load AWS config: %v", err)
	}

	// Create AWS Pricing client
	client := pricing.NewFromConfig(cfg)

	// Define service-specific filters for default configurations
	var filters []types.Filter
	switch service {
	case "Amazon EC2 (Elastic Compute Cloud)":
		filters = []types.Filter{
			{Field: aws.String("instanceType"), Value: aws.String("t2.micro"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("operatingSystem"), Value: aws.String("Linux"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("tenancy"), Value: aws.String("Shared"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("preInstalledSw"), Value: aws.String("NA"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("location"), Value: aws.String("US East (N. Virginia)"), Type: types.FilterTypeTermMatch},
		}
	case "Amazon S3 (Simple Storage Service)":
		filters = []types.Filter{
			{Field: aws.String("productFamily"), Value: aws.String("Storage"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("storageClass"), Value: aws.String("General Purpose"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("location"), Value: aws.String("US East (N. Virginia)"), Type: types.FilterTypeTermMatch},
		}
	case "AWS Lambda":
		filters = []types.Filter{
			{Field: aws.String("usagetype"), Value: aws.String("Lambda-GB-Second"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("location"), Value: aws.String("US East (N. Virginia)"), Type: types.FilterTypeTermMatch},
		}
	case "Amazon RDS (Relational Database Service)":
		filters = []types.Filter{
			{Field: aws.String("instanceType"), Value: aws.String("db.t3.micro"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("databaseEngine"), Value: aws.String("MySQL"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("deploymentOption"), Value: aws.String("Single-AZ"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("location"), Value: aws.String("US East (N. Virginia)"), Type: types.FilterTypeTermMatch},
		}
	// case "Amazon DynamoDB":
	// 	filters = []types.Filter{
	// 		{Field: aws.String("usagetype"), Value: aws.String("DynamoDB-WriteCapacityUnit-Hrs"), Type: types.FilterTypeTermMatch},
	// 		{Field: aws.String("location"), Value: aws.String("US East (N. Virginia)"), Type: types.FilterTypeTermMatch},
	// 	}
	case "AWS CloudFront":
		filters = []types.Filter{
			{Field: aws.String("productFamily"), Value: aws.String("Request"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("usagetype"), Value: aws.String("USE1-Requests-OriginShield"), Type: types.FilterTypeTermMatch}, // Correct usage type
			{Field: aws.String("location"), Value: aws.String("US East (N. Virginia)"), Type: types.FilterTypeTermMatch},
		}	
	case "Amazon VPC (Virtual Private Cloud)":
		filters = []types.Filter{
			{Field: aws.String("productFamily"), Value: aws.String("VpcEndpoint"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("endpointType"), Value: aws.String("Gateway Load Balancer Endpoint"), Type: types.FilterTypeTermMatch},
			{Field: aws.String("location"), Value: aws.String("US East (N. Virginia)"), Type: types.FilterTypeTermMatch},
		}	
	default:
		return 0, fmt.Errorf("unsupported AWS service: %s", service)
	}

	// Define input for Pricing API
	input := &pricing.GetProductsInput{
		ServiceCode:   aws.String(serviceToCode(service)),
		Filters:       filters,
		FormatVersion: aws.String("aws_v1"),
	}

	// Fetch pricing data from AWS Pricing API
	result, err := client.GetProducts(context.TODO(), input)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch AWS pricing data: %v", err)
	}

	// Parse the pricing data to extract the hourly price
	for _, priceItem := range result.PriceList {
		var priceData map[string]interface{}
		if err := json.Unmarshal([]byte(priceItem), &priceData); err != nil {
			continue // Skip items that fail to parse
		}

		if terms, ok := priceData["terms"].(map[string]interface{}); ok {
			if onDemand, ok := terms["OnDemand"].(map[string]interface{}); ok {
				for _, term := range onDemand {
					if priceDimensions, ok := term.(map[string]interface{})["priceDimensions"].(map[string]interface{}); ok {
						for _, dimension := range priceDimensions {
							if pricePerUnit, ok := dimension.(map[string]interface{})["pricePerUnit"].(map[string]interface{}); ok {
								if usdPrice, ok := pricePerUnit["USD"].(string); ok {
									// Convert price to float64
									price, err := strconv.ParseFloat(usdPrice, 64)
									if err != nil {
										continue
									}
									return price, nil
								}
							}
						}
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("price data not found for service: %s", service)
}

// Helper to map service names to their service codes
func serviceToCode(service string) string {
	switch service {
	case "Amazon EC2 (Elastic Compute Cloud)":
		return "AmazonEC2"
	case "Amazon S3 (Simple Storage Service)":
		return "AmazonS3"
	case "AWS Lambda":
		return "AWSLambda"
	case "Amazon RDS (Relational Database Service)":
		return "AmazonRDS"
	// case "Amazon DynamoDB":
	// 	return "AmazonDynamoDB"
	case "AWS CloudFront":
		return "AmazonCloudFront"
	case "Amazon VPC (Virtual Private Cloud)":
		return "AmazonVPC"
	default:
		return ""
	}
}

// FetchAWSServicePriceHandler fetches and displays AWS service pricing for debugging purposes
func FetchAWSServicePriceHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Service string `json:"service"`
	}

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Service == "" {
		http.Error(w, "Service is required", http.StatusBadRequest)
		return
	}

	// Fetch the service price
	price, err := fetchAWSServicePrice(req.Service)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch AWS pricing data: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the price and debug information
	response := map[string]interface{}{
		"service": req.Service,
		"price":   price,
		"units":   "USD/hour for EC2; USD/GB/month for S3; USD/GB-second for Lambda", // Unit clarification
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CalculateGCPCost dynamically calculates the estimated cost for a GCP service using the Cloud Billing API
func CalculateGCPCost(service string) (float64, error) {
	// Default region
	region := "us-east1"

	// Fetch the hourly price for the service
	pricePerHour, err := fetchGCPServicePrice(service, region)
	if err != nil {
		return 0, err
	}

	// Calculate cost for 90 days (quarter)
	hoursPerQuarter := float64(24 * 90)
	estimatedCost := hoursPerQuarter * pricePerHour

	// Round to 2 decimal places
	estimatedCost = math.Round(estimatedCost*100) / 100
	return estimatedCost, nil
}

// fetchGCPServicePrice dynamically fetches the hourly price of a GCP service for a specific region
func fetchGCPServicePrice(service, region string) (float64, error) {
	ctx := context.Background()
	client, err := cloudbilling.NewService(ctx, option.WithCredentialsFile("path/to/credentials.json"))
	if err != nil {
		return 0, fmt.Errorf("failed to create GCP billing client: %v", err)
	}

	// Get the list of services from the Cloud Billing API
	services, err := client.Services.List().Do()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch GCP services: %v", err)
	}

	// Find the matching service
	var serviceName string
	for _, serviceData := range services.Services {
		if serviceData.DisplayName == service {
			serviceName = serviceData.Name
			break
		}
	}

	if serviceName == "" {
		return 0, fmt.Errorf("service '%s' not found", service)
	}

	// Fetch the pricing details for the service
	skus, err := client.Services.Skus.List(serviceName).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch pricing SKUs: %v", err)
	}

	// Parse the SKUs to find the price for the given region
	for _, sku := range skus.Skus {
		for _, pricingInfo := range sku.PricingInfo {
			if contains(sku.ServiceRegions, region) || contains(sku.ServiceRegions, "global") {
				unitPrice := pricingInfo.PricingExpression.TieredRates[0].UnitPrice
				price := float64(unitPrice.Nanos) / 1e9
				return price, nil
			}
		}
	}
	return 0, fmt.Errorf("price data not found for service '%s' in region '%s'", service, region)
}

// contains checks if a slice contains a specific string
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
