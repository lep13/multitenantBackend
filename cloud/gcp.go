package cloud

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/protobuf/proto"
	"multitenant/models"
)

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
		Network:    proto.String(fmt.Sprintf("projects/%s/global/networks/%s", req.ProjectID, req.Network)),       // Fixed type: *string
		Subnetwork: proto.String(fmt.Sprintf("regions/%s/subnetworks/%s", req.Region, req.Subnetwork)),           // Fixed type: *string
	}

	// Define the instance
	instance := &computepb.Instance{
		Name:        proto.String(req.Name),                                                                   // Fixed type: *string
		MachineType: proto.String(fmt.Sprintf("zones/%s/machineTypes/%s", req.Zone, req.MachineType)),         // Fixed type: *string
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
