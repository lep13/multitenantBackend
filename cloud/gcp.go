package cloud

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"multitenant/models"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/compute/metadata"
	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"

	// functionspb "cloud.google.com/go/functions/apiv1/functionspb"
	// "google.golang.org/protobuf/types/known/durationpb"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/storage"
	"google.golang.org/api/sqladmin/v1"
	"google.golang.org/protobuf/proto"
)

// FetchProjectID dynamically retrieves the GCP project ID
func FetchProjectID() (string, error) {
	// Check if running on GCP
	if metadata.OnGCE() {
		projectID, err := metadata.ProjectID()
		if err != nil {
			return "", fmt.Errorf("failed to fetch project ID from metadata: %v", err)
		}
		return projectID, nil
	}

	// Check environment variable
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID != "" {
		return projectID, nil
	}

	// Fallback to `gcloud` CLI
	cmd := exec.Command("gcloud", "config", "get-value", "project")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to fetch project ID using gcloud CLI: %v", err)
	}

	projectID = strings.TrimSpace(out.String())
	if projectID == "" {
		return "", fmt.Errorf("GCP project ID could not be determined")
	}

	return projectID, nil
}

// CreateComputeEngineInstance creates a GCP Compute Engine instance
func CreateComputeEngineInstance(req models.GCPInstanceRequest) (*compute.Operation, error) {
	ctx := context.Background()

	// Create the Compute Engine client
	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Compute Engine client: %v", err)
	}
	defer client.Close()

	// Define the attached disk
	disk := &computepb.AttachedDisk{
		AutoDelete: proto.Bool(true),
		Boot:       proto.Bool(true),
		Type:       proto.String("PERSISTENT"), // Enum converted to string
		InitializeParams: &computepb.AttachedDiskInitializeParams{
			SourceImage: proto.String(fmt.Sprintf("projects/%s/global/images/family/%s", req.ImageProject, req.ImageFamily)),
		},
	}

	// Define the network interface
	networkInterface := &computepb.NetworkInterface{
		Network:    proto.String(fmt.Sprintf("projects/%s/global/networks/%s", req.ProjectID, req.Network)), // Fixed type: *string
		Subnetwork: proto.String(fmt.Sprintf("regions/%s/subnetworks/%s", req.Region, req.Subnetwork)),      // Fixed type: *string
	}

	// Define the instance
	instance := &computepb.Instance{
		Name:        proto.String(req.Name),                                                           // Fixed type: *string
		MachineType: proto.String(fmt.Sprintf("zones/%s/machineTypes/%s", req.Zone, req.MachineType)), // Fixed type: *string
		Disks:       []*computepb.AttachedDisk{disk},
		NetworkInterfaces: []*computepb.NetworkInterface{
			networkInterface,
		},
		ServiceAccounts: []*computepb.ServiceAccount{
			{
				Email: proto.String(req.ServiceAccount), // Fixed type: *string
				Scopes: []string{
					"https://www.googleapis.com/auth/cloud-platform",
				},
			},
		},
	}

	// Define the request
	insertReq := &computepb.InsertInstanceRequest{
		Project:          req.ProjectID,
		Zone:             req.Zone,
		InstanceResource: instance,
	}

	// Call the API to insert the instance
	op, err := client.Insert(ctx, insertReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %v", err)
	}

	return op, nil
}

// CreateCloudStorage creates a Cloud Storage bucket
func CreateCloudStorage(bucketName, region string) (*storage.BucketHandle, error) {
	// Fetch the project ID dynamically
	projectID, err := FetchProjectID()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch project ID: %v", err)
	}

	// Proceed with bucket creation
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Storage client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)

	// Create bucket with the specified region
	if err := bucket.Create(context.Background(), projectID, &storage.BucketAttrs{
		Location: region,
	}); err != nil {
		return nil, fmt.Errorf("failed to create bucket: %v", err)
	}

	return bucket, nil
}

// CreateGKECluster creates a GKE cluster
func CreateGKECluster(clusterName, zone, region, machineType, network, subnetwork string, nodeCount int) (*containerpb.Operation, error) {
	ctx := context.Background()

	// Fetch project ID dynamically
	projectID, err := FetchProjectID()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project ID: %v", err)
	}

	// Create the GKE client
	client, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GKE client: %v", err)
	}
	defer client.Close()

	// Define the cluster configuration
	cluster := &containerpb.Cluster{
		Name:             clusterName,
		InitialNodeCount: int32(nodeCount),
		Location:         region,
		NodeConfig: &containerpb.NodeConfig{
			MachineType: machineType,
		},
		Network:    fmt.Sprintf("projects/%s/global/networks/%s", projectID, network),
		Subnetwork: fmt.Sprintf("regions/%s/subnetworks/%s", region, subnetwork),
	}

	// Create the request
	req := &containerpb.CreateClusterRequest{
		Parent:  fmt.Sprintf("projects/%s/locations/%s", projectID, zone),
		Cluster: cluster,
	}

	// Call the GKE API to create the cluster
	op, err := client.CreateCluster(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create GKE cluster: %v", err)
	}

	return op, nil
}

// CreateBigQueryDataset creates a BigQuery dataset
func CreateBigQueryDataset(datasetID, region string) (*bigquery.Dataset, error) {
	projectID, err := FetchProjectID() // Dynamically fetch the project ID
	if err != nil {
		return nil, fmt.Errorf("unable to fetch project ID: %v", err)
	}

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create BigQuery client: %v", err)
	}
	defer client.Close()

	// Define the dataset
	dataset := client.Dataset(datasetID)
	meta := &bigquery.DatasetMetadata{
		Location: region,
	}

	// Create the dataset
	err = dataset.Create(ctx, meta)
	if err != nil {
		return nil, fmt.Errorf("failed to create BigQuery dataset: %v", err)
	}

	return dataset, nil
}

// CreateCloudSQLInstance creates a Cloud SQL instance
func CreateCloudSQLInstance(instanceName, projectID, region, tier, databaseVersion string) (*sqladmin.Operation, error) {
	ctx := context.Background()
	client, err := sqladmin.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud SQL client: %v", err)
	}

	req := &sqladmin.DatabaseInstance{
		Name:            instanceName,
		Project:         projectID,
		Region:          region,
		DatabaseVersion: databaseVersion,
		Settings: &sqladmin.Settings{
			Tier: tier,
		},
	}

	op, err := client.Instances.Insert(projectID, req).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud SQL instance: %v", err)
	}
	return op, nil
}

// // DeployCloudFunction deploys a Google Cloud Function
// func DeployCloudFunction(functionName, region, runtime, entryPoint, bucketName, objectName string, environmentVariables map[string]string, triggerHTTP bool) (*functionspb.OperationMetadataV1, error) {
// 	projectID, err := FetchProjectID() // Dynamically fetch the project ID
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to fetch project ID: %v", err)
// 	}

// 	ctx := context.Background()
// 	client, err := functionspb.NewCloudFunctionsClient(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create Cloud Functions client: %v", err)
// 	}
// 	defer client.Close()

// 	// Define the Cloud Function
// 	function := &functionspb.CloudFunction{
// 		Name:        fmt.Sprintf("projects/%s/locations/%s/functions/%s", projectID, region, functionName),
// 		Runtime:     runtime,
// 		EntryPoint:  entryPoint,
// 		EnvironmentVariables: environmentVariables,
// 		SourceCode: &functionspb.SourceCode{
// 			SourceCode: &functionspb.SourceCode_StorageSource{
// 				StorageSource: &functionspb.StorageSource{
// 					Bucket: bucketName,
// 					Object: objectName,
// 				},
// 			},
// 		},
// 		Timeout: durationpb.New(60 * time.Second), // Function timeout
// 	}

// 	// Define the trigger
// 	if triggerHTTP {
// 		function.HttpsTrigger = &functionspb.HttpsTrigger{}
// 	} else {
// 		function.EventTrigger = &functionspb.EventTrigger{
// 			EventType: "google.storage.object.finalize",
// 			Resource:  fmt.Sprintf("projects/_/buckets/%s", bucketName),
// 		}
// 	}

// 	// Create the request
// 	req := &functionspb.CreateFunctionRequest{
// 		Parent:   fmt.Sprintf("projects/%s/locations/%s", projectID, region),
// 		Function: function,
// 	}

// 	// Call the API to create the function
// 	op, err := client.CreateFunction(ctx, req)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to deploy Cloud Function: %v", err)
// 	}

// 	// Return operation metadata for additional details
// 	return op.Metadata.(*functionspb.OperationMetadataV1), nil
// }