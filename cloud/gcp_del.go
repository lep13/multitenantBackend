package cloud

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/sqladmin/v1"
)

// DeleteComputeEngineInstance deletes a GCP Compute Engine instance
func DeleteComputeEngineInstance(instanceName, zone string) (string, error) {
	ctx := context.Background()

	client, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create Compute Engine client: %v", err)
	}
	defer client.Close()

	// Fetch the project ID dynamically
	projectID, err := FetchProjectID()
	if err != nil {
		return "", fmt.Errorf("unable to fetch project ID: %v", err)
	}

	req := &computepb.DeleteInstanceRequest{
		Project:  projectID,
		Zone:     zone,
		Instance: instanceName,
	}

	_, err = client.Delete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to delete Compute Engine instance: %v", err)
	}

	return fmt.Sprintf("Compute Engine instance '%s' deleted successfully", instanceName), nil
}

// DeleteCloudStorage deletes a GCP Cloud Storage bucket
func DeleteCloudStorage(bucketName string) (string, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create Cloud Storage client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)

	// Delete all objects in the bucket
	it := bucket.Objects(ctx, nil)
	for {
		obj, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to iterate over bucket objects: %v", err)
		}

		if err := bucket.Object(obj.Name).Delete(ctx); err != nil {
			return "", fmt.Errorf("failed to delete object '%s': %v", obj.Name, err)
		}
	}

	// Delete the bucket itself
	if err := bucket.Delete(ctx); err != nil {
		return "", fmt.Errorf("failed to delete bucket: %v", err)
	}

	return fmt.Sprintf("Cloud Storage bucket '%s' deleted successfully", bucketName), nil
}

// DeleteGKECluster deletes a GCP GKE cluster
func DeleteGKECluster(clusterName, zone string) (string, error) {
	ctx := context.Background()

	client, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create GKE client: %v", err)
	}
	defer client.Close()

	// Fetch project ID dynamically
	projectID, err := FetchProjectID()
	if err != nil {
		return "", fmt.Errorf("unable to fetch project ID: %v", err)
	}

	req := &containerpb.DeleteClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, zone, clusterName),
	}

	_, err = client.DeleteCluster(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to delete GKE cluster: %v", err)
	}

	return fmt.Sprintf("GKE cluster '%s' deleted successfully", clusterName), nil
}

// DeleteBigQueryDataset deletes a GCP BigQuery dataset
func DeleteBigQueryDataset(datasetID string) (string, error) {
	projectID, err := FetchProjectID()
	if err != nil {
		return "", fmt.Errorf("unable to fetch project ID: %v", err)
	}

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to create BigQuery client: %v", err)
	}
	defer client.Close()

	dataset := client.Dataset(datasetID)

	// Delete the dataset and its contents
	err = dataset.DeleteWithContents(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to delete BigQuery dataset: %v", err)
	}

	return fmt.Sprintf("BigQuery dataset '%s' deleted successfully", datasetID), nil
}

// DeleteCloudSQLInstance deletes a GCP Cloud SQL instance
func DeleteCloudSQLInstance(instanceName string) (string, error) {
	ctx := context.Background()
	client, err := sqladmin.NewService(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create Cloud SQL client: %v", err)
	}

	// Fetch the project ID dynamically
	projectID, err := FetchProjectID()
	if err != nil {
		return "", fmt.Errorf("unable to fetch project ID: %v", err)
	}

	_, err = client.Instances.Delete(projectID, instanceName).Do()
	if err != nil {
		return "", fmt.Errorf("failed to delete Cloud SQL instance: %v", err)
	}

	return fmt.Sprintf("Cloud SQL instance '%s' deleted successfully", instanceName), nil
}
