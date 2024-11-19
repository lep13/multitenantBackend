package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"multitenant/db"
	"net/http"
	"strconv"
	"strings"

	billing "cloud.google.com/go/billing/apiv1"
	"cloud.google.com/go/billing/apiv1/billingpb"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/aws/aws-sdk-go/aws"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/api/iterator"
)

// calculates the estimated cost of a service
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
	var status string
	var message string

	if provider == "aws" {
		switch service {
		case "Amazon EC2 (Elastic Compute Cloud)", "Amazon S3 (Simple Storage Service)", "Amazon RDS (Relational Database Service)":
			// For EC2, S3, and RDS, calculate estimated cost
			estimatedCost, err = CalculateAWSCost(service)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to calculate cost: %v", err), http.StatusInternalServerError)
				return
			}
			// Compare the estimated cost with the budget
			if estimatedCost > budget {
				status = "denied"
				message = fmt.Sprintf(
					"Estimated cost of this service is $%.2f, which exceeds your budget of $%.2f. Request denied.",
					estimatedCost, budget,
				)
			} else {
				status = "ok"
				message = fmt.Sprintf(
					"Estimated cost of this service is $%.2f. Your budget is $%.2f. You can proceed with the service creation.",
					estimatedCost, budget,
				)
			}

		case "Amazon VPC (Virtual Private Cloud)", "AWS Lambda", "Amazon DynamoDB", "AWS CloudFront":
			// No cost calculation for these services
			status = "ok"
			message = fmt.Sprintf(
				"Estimated cost cannot be calculated for this service. Your budget is $%.2f. Do you want to create it?",
				budget,
			)

		default:
			http.Error(w, "Unsupported AWS service", http.StatusBadRequest)
			return
		}
	} else if provider == "gcp" {
		switch service {
		case "Compute Engine", "Cloud Storage", "Cloud SQL":
			// For Compute Engine, Cloud Storage, and Cloud SQL, calculate estimated cost
			estimatedCost, err = CalculateGCPCost(service)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to calculate cost: %v", err), http.StatusInternalServerError)
				return
			}
			// Compare the estimated cost with the budget
			if estimatedCost > budget {
				status = "denied"
				message = fmt.Sprintf(
					"Estimated cost of this service is $%.2f, which exceeds your budget of $%.2f. Request denied.",
					estimatedCost, budget,
				)
			} else {
				status = "ok"
				message = fmt.Sprintf(
					"Estimated cost of this service is $%.2f. Your budget is $%.2f. You can proceed with the service creation.",
					estimatedCost, budget,
				)
			}

		case "Google Kubernetes Engine (GKE)", "BigQuery", "Cloud Functions":
			// No cost calculation for these services
			status = "ok"
			message = fmt.Sprintf(
				"Estimated cost cannot be calculated for this service. Your budget is $%.2f. Do you want to create it?",
				budget,
			)

		default:
			http.Error(w, "Unsupported GCP service", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Unsupported cloud provider", http.StatusBadRequest)
		return
	}

	// Update session with estimated cost and status
	_, err = db.UpdateSessionWithCost(req.SessionID, estimatedCost, status)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update session: %v", err), http.StatusBadRequest)
		return
	}

	// Respond with the status, estimated cost, budget, and message
	response := map[string]interface{}{
		"status":         status,
		"estimated_cost": estimatedCost,
		"budget":         budget,
		"message":        message,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CalculateAWSCost calculates the quarterly cost for an AWS service based on its pricing unit.
func CalculateAWSCost(service string) (float64, error) {
	// Fetch the hourly price for the service.
	pricePerUnit, err := fetchAWSServicePrice(service)
	if err != nil {
		// Return 0 cost and no error for services where cost calculation isn't supported.
		if strings.Contains(err.Error(), "cost fetching is not supported") {
			return 0, nil
		}
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
		monthlyCost := pricePerUnit * 1024 // 1TB = 1024GB
		estimatedCost = monthlyCost * 3    // Quarterly cost (3 months)

	case "Amazon RDS (Relational Database Service)":
		// RDS: Unit is USD/hour
		hoursPerQuarter := float64(24 * 90) // 24 hours/day * 90 days
		estimatedCost = pricePerUnit * hoursPerQuarter

	default:
		// For services like Lambda, CloudFront, DynamoDB, and VPC:
		return 0, fmt.Errorf("cost calculation is not supported for this service: %s", service)
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
	// case "AWS Lambda":
	// 	filters = []types.Filter{
	// 		{Field: aws.String("usagetype"), Value: aws.String("Lambda-GB-Second"), Type: types.FilterTypeTermMatch},
	// 		{Field: aws.String("location"), Value: aws.String("US East (N. Virginia)"), Type: types.FilterTypeTermMatch},
	// 	}
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
	// case "AWS CloudFront":
	// 	filters = []types.Filter{
	// 		{Field: aws.String("productFamily"), Value: aws.String("Request"), Type: types.FilterTypeTermMatch},
	// 		{Field: aws.String("usagetype"), Value: aws.String("USE1-Requests-OriginShield"), Type: types.FilterTypeTermMatch}, // Correct usage type
	// 		{Field: aws.String("location"), Value: aws.String("US East (N. Virginia)"), Type: types.FilterTypeTermMatch},
	// 	}
	// case "Amazon VPC (Virtual Private Cloud)":
	// 	filters = []types.Filter{
	// 		{Field: aws.String("productFamily"), Value: aws.String("VpcEndpoint"), Type: types.FilterTypeTermMatch},
	// 		{Field: aws.String("endpointType"), Value: aws.String("Gateway Load Balancer Endpoint"), Type: types.FilterTypeTermMatch},
	// 		{Field: aws.String("location"), Value: aws.String("US East (N. Virginia)"), Type: types.FilterTypeTermMatch},
	// 	}
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
	case "Amazon DynamoDB":
		return "AmazonDynamoDB"
	case "AWS CloudFront":
		return "AmazonCloudFront"
	case "Amazon VPC (Virtual Private Cloud)":
		return "AmazonVPC"
	default:
		return ""
	}
}

func CalculateGCPCost(service string) (float64, error) {
	// Fetch the price and unit from the GCP pricing API
	pricePerUnit, _, err := FetchGCPServicePrice(service)
	if err != nil {
		// Return 0 cost and no error for services where cost calculation isn't supported.
		if strings.Contains(err.Error(), "cost fetching is not supported") {
			return 0, nil
		}
		return 0, err
	}

	var estimatedCost float64

	// Calculate cost based on service-specific units
	switch service {
	case "Compute Engine":
		// Unit: USD/hour
		hoursPerQuarter := float64(24 * 90) // 24 hours/day * 90 days
		estimatedCost = pricePerUnit * hoursPerQuarter

	case "Cloud Storage":
		// Unit: USD/GB/month
		monthlyCost := pricePerUnit * 1024 // assuming 1TB = 1024GB
		estimatedCost = monthlyCost * 3    // Quarterly cost (3 months)

	case "Cloud SQL":
		// Unit: USD/hour
		hoursPerQuarter := float64(24 * 90) // 24 hours/day * 90 days
		estimatedCost = pricePerUnit * hoursPerQuarter

	default:
		// For services like GKE, BigQuery, and Cloud Functions
		return 0, fmt.Errorf("cost calculation is not supported for this service: %s", service)
	}

	// Round to 2 decimal places for clarity
	estimatedCost = math.Round(estimatedCost*100) / 100
	return estimatedCost, nil
}

// fetches the dynamic pricing for a GCP service using Billing API.
func FetchGCPServicePrice(service string) (float64, string, error) {
	ctx := context.Background()

	// Create a Cloud Catalog client
	client, err := billing.NewCloudCatalogClient(ctx)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create Cloud Catalog client: %v", err)
	}
	defer client.Close()

	// Map service names to their display names in the Cloud Catalog
	serviceMap := map[string]string{
		"Compute Engine":                "Compute Engine",
		"Cloud Storage":                 "Cloud Storage",
		"Cloud SQL":                     "Cloud SQL",
	}

	displayName, exists := serviceMap[service]
	if !exists {
		return 0, "", fmt.Errorf("unsupported GCP service: %s", service)
	}

	// List all services and find the matching one
	req := &billingpb.ListServicesRequest{}
	it := client.ListServices(ctx, req)
	for {
		svc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, "", fmt.Errorf("error while iterating services: %v", err)
		}

		// Match service by display name
		if svc.DisplayName == displayName {
			// List SKUs for the matched service
			skuReq := &billingpb.ListSkusRequest{
				Parent: svc.Name,
			}
			skuIt := client.ListSkus(ctx, skuReq)
			for {
				sku, err := skuIt.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					return 0, "", fmt.Errorf("error while iterating SKUs: %v", err)
				}

				// Compute Engine
				if service == "Compute Engine" && sku.Category.ResourceFamily == "Compute" {
					if sku.PricingInfo != nil && len(sku.PricingInfo) > 0 {
						pricing := sku.PricingInfo[0].PricingExpression
						if pricing != nil && len(pricing.TieredRates) > 0 && pricing.UsageUnit == "h" {
							price := pricing.TieredRates[0].UnitPrice
							return float64(price.Units) + float64(price.Nanos)/1e9, pricing.UsageUnitDescription, nil
						}
					}
				}

				// Cloud Storage
				if sku.Category.ResourceFamily == "Storage" && service == "Cloud Storage" {
					if sku.PricingInfo != nil && len(sku.PricingInfo) > 0 {
						pricing := sku.PricingInfo[0].PricingExpression
						if pricing != nil && len(pricing.TieredRates) > 0 && pricing.UsageUnit == "GiBy.mo" {
							price := pricing.TieredRates[0].UnitPrice
							return float64(price.Units) + float64(price.Nanos)/1e9, pricing.UsageUnitDescription, nil
						}
					}
				}

				// Cloud SQL
				if service == "Cloud SQL" && strings.Contains(sku.Description, "MySQL") && strings.Contains(sku.Description, "vCPU") {
					if sku.PricingInfo != nil && len(sku.PricingInfo) > 0 {
						pricing := sku.PricingInfo[0].PricingExpression
						if pricing != nil && len(pricing.TieredRates) > 0 {
							price := pricing.TieredRates[0].UnitPrice
							return float64(price.Units) + float64(price.Nanos)/1e9, pricing.UsageUnitDescription, nil
						}
					}
				}
			}
		}
	}

	return 0, "", fmt.Errorf("pricing data not found for service: %s", service)
}



// FetchGCPServicePriceHandler handles requests to fetch the minimum pricing for a GCP service.
func FetchGCPServicePriceHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Service string `json:"service"`
	}

	// Decode the incoming request payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Service == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	// Fetch the minimum price for the specified GCP service
	price, unit, err := FetchGCPServicePrice(req.Service)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch GCP pricing data: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the fetched price and unit in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": req.Service,
		"price":   price,
		"unit":    unit,
		"currency": "USD", // Assuming the pricing is in USD
	})
}

// // FetchAWSServicePriceHandler fetches and displays AWS service pricing for debugging purposes
// func FetchAWSServicePriceHandler(w http.ResponseWriter, r *http.Request) {
// 	var req struct {
// 		Service string `json:"service"`
// 	}

// 	// Decode the request body
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	if req.Service == "" {
// 		http.Error(w, "Service is required", http.StatusBadRequest)
// 		return
// 	}

// 	// Fetch the service price
// 	price, err := fetchAWSServicePrice(req.Service)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to fetch AWS pricing data: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	// Respond with the price and debug information
// 	response := map[string]interface{}{
// 		"service": req.Service,
// 		"price":   price,
// 		"units":   "USD/hour for EC2; USD/GB/month for S3; USD/GB-second for Lambda", // Unit clarification
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(response)
// }