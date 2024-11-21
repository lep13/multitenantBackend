package cloud

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// LoadAWSConfig loads the AWS configuration
func LoadAWSConfig() (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return cfg, nil
}

// DeleteLambdaFunction deletes an AWS Lambda function
func DeleteLambdaFunction(functionName string) (string, error) {
	cfg, err := LoadAWSConfig()
	if err != nil {
		return "", err
	}

	client := lambda.NewFromConfig(cfg)

	_, err = client.DeleteFunction(context.TODO(), &lambda.DeleteFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to delete Lambda function: %w", err)
	}

	return fmt.Sprintf("Lambda function '%s' deleted successfully", functionName), nil
}

// TerminateEC2Instance terminates an EC2 instance
func TerminateEC2Instance(instanceID string) (interface{}, error) {
	cfg, err := LoadAWSConfig()
	if err != nil {
		return nil, err
	}

	client := ec2.NewFromConfig(cfg)

	_, err = client.TerminateInstances(context.TODO(), &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to terminate EC2 instance: %w", err)
	}

	return fmt.Sprintf("EC2 instance '%s' terminated successfully", instanceID), nil
}

// DeleteS3Bucket deletes an S3 bucket
func DeleteS3Bucket(bucketName string) (string, error) {
	cfg, err := LoadAWSConfig()
	if err != nil {
		return "", err
	}

	client := s3.NewFromConfig(cfg)

	// List and delete all objects in the bucket
	listObjects, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list objects in S3 bucket: %w", err)
	}

	for _, obj := range listObjects.Contents {
		_, err = client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    obj.Key,
		})
		if err != nil {
			return "", fmt.Errorf("failed to delete object '%s': %w", *obj.Key, err)
		}
	}

	// Delete the bucket
	_, err = client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to delete S3 bucket: %w", err)
	}

	return fmt.Sprintf("S3 bucket '%s' deleted successfully", bucketName), nil
}

// DeleteRDSInstance deletes an RDS instance
func DeleteRDSInstance(instanceID string) (interface{}, error) {
	cfg, err := LoadAWSConfig()
	if err != nil {
		return nil, err
	}

	client := rds.NewFromConfig(cfg)

	_, err = client.DeleteDBInstance(context.TODO(), &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(instanceID),
		SkipFinalSnapshot:    aws.Bool(true), // Skip final snapshot for immediate deletion
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete RDS instance: %w", err)
	}

	return fmt.Sprintf("RDS instance '%s' deleted successfully", instanceID), nil
}

// DeleteCloudFrontDistribution deletes a CloudFront distribution
func DisableCloudFrontDistribution(distributionID string) (string, error) {
	cfg, err := LoadAWSConfig()
	if err != nil {
		return "", err
	}

	client := cloudfront.NewFromConfig(cfg)

	// Get the distribution configuration
	distConfig, err := client.GetDistributionConfig(context.TODO(), &cloudfront.GetDistributionConfigInput{
		Id: aws.String(distributionID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get CloudFront distribution config: %w", err)
	}

	// Disable the distribution
	distConfig.DistributionConfig.Enabled = aws.Bool(false)
	_, err = client.UpdateDistribution(context.TODO(), &cloudfront.UpdateDistributionInput{
		Id:                 aws.String(distributionID),
		IfMatch:            distConfig.ETag,
		DistributionConfig: distConfig.DistributionConfig,
	})
	if err != nil {
		return "", fmt.Errorf("failed to disable CloudFront distribution: %w", err)
	}

	return fmt.Sprintf("CloudFront distribution '%s' disabled successfully", distributionID), nil
}

// DeleteVPC deletes a VPC
func DeleteVPC(vpcID string) (string, error) {
	cfg, err := LoadAWSConfig()
	if err != nil {
		return "", err
	}

	client := ec2.NewFromConfig(cfg)

	// Check and detach any internet gateways associated with the VPC
	igwOutput, err := client.DescribeInternetGateways(context.TODO(), &ec2.DescribeInternetGatewaysInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("attachment.vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe internet gateways: %w", err)
	}

	for _, igw := range igwOutput.InternetGateways {
		_, err := client.DetachInternetGateway(context.TODO(), &ec2.DetachInternetGatewayInput{
			InternetGatewayId: igw.InternetGatewayId,
			VpcId:             aws.String(vpcID),
		})
		if err != nil {
			return "", fmt.Errorf("failed to detach internet gateway: %w", err)
		}

		_, err = client.DeleteInternetGateway(context.TODO(), &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: igw.InternetGatewayId,
		})
		if err != nil {
			return "", fmt.Errorf("failed to delete internet gateway: %w", err)
		}
	}

	// Delete subnets associated with the VPC
	subnetsOutput, err := client.DescribeSubnets(context.TODO(), &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe subnets: %w", err)
	}

	for _, subnet := range subnetsOutput.Subnets {
		_, err = client.DeleteSubnet(context.TODO(), &ec2.DeleteSubnetInput{
			SubnetId: subnet.SubnetId,
		})
		if err != nil {
			return "", fmt.Errorf("failed to delete subnet: %w", err)
		}
	}

	// Delete route tables associated with the VPC
	routeTablesOutput, err := client.DescribeRouteTables(context.TODO(), &ec2.DescribeRouteTablesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcID},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe route tables: %w", err)
	}

	for _, routeTable := range routeTablesOutput.RouteTables {
		// Skip the main route table
		isMain := false
		for _, assoc := range routeTable.Associations {
			if assoc.Main != nil && *assoc.Main {
				isMain = true
				break
			}
		}
		if isMain {
			continue
		}

		_, err = client.DeleteRouteTable(context.TODO(), &ec2.DeleteRouteTableInput{
			RouteTableId: routeTable.RouteTableId,
		})
		if err != nil {
			return "", fmt.Errorf("failed to delete route table: %w", err)
		}
	}

	// Finally, delete the VPC
	_, err = client.DeleteVpc(context.TODO(), &ec2.DeleteVpcInput{
		VpcId: aws.String(vpcID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to delete VPC: %w", err)
	}

	return fmt.Sprintf("VPC '%s' deleted successfully", vpcID), nil
}