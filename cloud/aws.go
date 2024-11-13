package cloud

import (
	"fmt"
	"os/exec"
	"os"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	// "github.com/aws/aws-sdk-go-v2/service/cloudfront"
	// cloudfronttypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	// "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	// dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types" 
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	// "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var cfg aws.Config

// Initialize AWS configuration
func InitAWS() error {
    var err error
    cfg, err = config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1")) 
	fmt.Println("Region in configuration:", cfg.Region)
    if err != nil {
        return fmt.Errorf("unable to load SDK config, %v", err)
    }
    return nil
}

// Run Terraform Apply
func ApplyTerraform() error {
	cmd := exec.Command("terraform", "apply", "-auto-approve")
	cmd.Dir = "./terraform" // Set the working directory to where Terraform files are located
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error applying Terraform: %v - %s", err, output)
	}
	fmt.Println(string(output))
	return nil
}

// Run Terraform Destroy
func DestroyTerraform() error {
	cmd := exec.Command("terraform", "destroy", "-auto-approve")
	cmd.Dir = "./terraform"
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error destroying resources: %v - %s", err, output)
	}
	fmt.Println(string(output))
	return nil
}

// Retrieve Terraform Outputs (e.g., for instance ID or public IP)
func GetTerraformOutput(outputName string) (string, error) {
	cmd := exec.Command("terraform", "output", "-raw", outputName)
	cmd.Dir = "./terraform"
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error retrieving output '%s': %v", outputName, err)
	}
	return string(output), nil
}

// EC2 Instance Creation
func CreateEC2Instance(instanceType, amiID, keyName, subnetID, securityGroupID string) (*ec2.RunInstancesOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %v", err)
	}
	ec2Client := ec2.NewFromConfig(cfg)


    input := &ec2.RunInstancesInput{
        ImageId:      aws.String(amiID),
        InstanceType: ec2types.InstanceType(instanceType),
        KeyName:      aws.String(keyName),
        SubnetId:     aws.String(subnetID),
        SecurityGroupIds: []string{
            securityGroupID, // Only Security Group IDs should be provided
        },
        MinCount: aws.Int32(1),
        MaxCount: aws.Int32(1),
    }

    result, err := ec2Client.RunInstances(context.Background(), input)
    if err != nil {
        return nil, fmt.Errorf("could not create EC2 instance: %v", err)
    }
    return result, nil
}

// S3 Bucket Creation
func CreateS3Bucket(bucketName string, enableVersioning bool, Region string) (*s3.CreateBucketOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %v", err)
	}
	s3Client := s3.NewFromConfig(cfg)

	// Create bucket
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	result, err := s3Client.CreateBucket(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("could not create S3 bucket: %v", err)
	}

	// Enable versioning if requested
	if enableVersioning {
		_, err = s3Client.PutBucketVersioning(context.Background(), &s3.PutBucketVersioningInput{
			Bucket: aws.String(bucketName),
			VersioningConfiguration: &s3types.VersioningConfiguration{
				Status: s3types.BucketVersioningStatusEnabled,
			},
		})
		if err != nil {
			return result, fmt.Errorf("failed to enable versioning: %v", err)
		}
	}
	return result, nil
}

//Lambda Function Creation
const lambdaExecutionRoleARN = "arn:aws:iam::058264391220:role/LambdaExecutionRole" 
func CreateLambdaFunction(functionName, handler, runtime, zipFilePath, region string) (*lambda.CreateFunctionOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %v", err)
	}
	lambdaClient := lambda.NewFromConfig(cfg)

	// Read zip file contents
	code, err := os.ReadFile(zipFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read zip file: %v", err)
	}

	input := &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionName),
		Role:         aws.String(lambdaExecutionRoleARN), // Use the hardcoded role here
		Handler:      aws.String(handler),
		Runtime:      lambdatypes.Runtime(runtime),
		Code: &lambdatypes.FunctionCode{
			ZipFile: code,
		},
	}

	result, err := lambdaClient.CreateFunction(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("could not create Lambda function: %v", err)
	}
	return result, nil
}

// // RDS Instance Creation
// func CreateRDSInstance(dbName, instanceID, instanceClass, engine, username, password string, allocatedStorage int32) (*rds.CreateDBInstanceOutput, error) {
// 	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to load config: %v", err)
// 	}
// 	rdsClient := rds.NewFromConfig(cfg)

// 	input := &rds.CreateDBInstanceInput{
// 		DBName:               aws.String(dbName),
// 		DBInstanceIdentifier: aws.String(instanceID),
// 		DBInstanceClass:      aws.String(instanceClass),
// 		Engine:               aws.String(engine),
// 		MasterUsername:       aws.String(username),
// 		MasterUserPassword:   aws.String(password),
// 		AllocatedStorage:     aws.Int32(allocatedStorage),
// 	}

// 	result, err := rdsClient.CreateDBInstance(context.TODO(), input)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not create RDS instance: %v", err)
// 	}
// 	return result, nil
// }

// // DynamoDB Table Creation
// func CreateDynamoDBTable(tableName string, readCapacity, writeCapacity int64) (*dynamodb.CreateTableOutput, error) {
// 	dynamoClient := dynamodb.NewFromConfig(cfg)

// 	input := &dynamodb.CreateTableInput{
// 		TableName: aws.String(tableName),
// 		AttributeDefinitions: []dynamodbtypes.AttributeDefinition{
// 			{
// 				AttributeName: aws.String("PrimaryKey"),
// 				AttributeType: dynamodbtypes.ScalarAttributeTypeS,
// 			},
// 		},
// 		KeySchema: []dynamodbtypes.KeySchemaElement{
// 			{
// 				AttributeName: aws.String("PrimaryKey"),
// 				KeyType:       dynamodbtypes.KeyTypeHash,
// 			},
// 		},
// 		ProvisionedThroughput: &dynamodbtypes.ProvisionedThroughput{
// 			ReadCapacityUnits:  aws.Int64(readCapacity),
// 			WriteCapacityUnits: aws.Int64(writeCapacity),
// 		},
// 	}

// 	result, err := dynamoClient.CreateTable(context.TODO(), input)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not create DynamoDB table: %v", err)
// 	}
// 	return result, nil
// }

// // CloudFront Distribution Creation
// func CreateCloudFrontDistribution(originDomainName string) (*cloudfront.CreateDistributionOutput, error) {
// 	cloudFrontClient := cloudfront.NewFromConfig(cfg)

// 	input := &cloudfront.CreateDistributionInput{
// 		DistributionConfig: &cloudfronttypes.DistributionConfig{
// 			Enabled: aws.Bool(true),
// 			Origins: &cloudfronttypes.Origins{
// 				Quantity: aws.Int32(1),
// 				Items: []cloudfronttypes.Origin{
// 					{
// 						Id:         aws.String("Origin1"),
// 						DomainName: aws.String(originDomainName),
// 					},
// 				},
// 			},
// 			DefaultCacheBehavior: &cloudfronttypes.DefaultCacheBehavior{
// 				TargetOriginId:       aws.String("Origin1"),
// 				ViewerProtocolPolicy: cloudfronttypes.ViewerProtocolPolicyAllowAll,
// 				AllowedMethods: &cloudfronttypes.AllowedMethods{
// 					Quantity: aws.Int32(2),
// 					Items:    []cloudfronttypes.Method{cloudfronttypes.MethodGet, cloudfronttypes.MethodHead},
// 				},
// 			},
// 		},
// 	}

// 	result, err := cloudFrontClient.CreateDistribution(context.TODO(), input)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not create CloudFront distribution: %v", err)
// 	}
// 	return result, nil
// }

// // VPC Creation
// func CreateVPC(cidrBlock string) (*ec2.CreateVpcOutput, error) {
// 	ec2Client := ec2.NewFromConfig(cfg)

// 	input := &ec2.CreateVpcInput{
// 		CidrBlock: aws.String(cidrBlock),
// 	}

// 	result, err := ec2Client.CreateVpc(context.TODO(), input)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not create VPC: %v", err)
// 	}
// 	return result, nil
// }