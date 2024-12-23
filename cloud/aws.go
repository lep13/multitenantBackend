package cloud

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	// "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	// dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cloudfronttypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
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

// EC2 Instance Creation
func CreateEC2Instance(instanceType, amiID, keyName, subnetID, securityGroupID, instanceName string) (*ec2.RunInstancesOutput, error) {
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
            securityGroupID,
        },
        MinCount: aws.Int32(1),
        MaxCount: aws.Int32(1),
        TagSpecifications: []ec2types.TagSpecification{
            {
                ResourceType: ec2types.ResourceTypeInstance,
                Tags: []ec2types.Tag{
                    {
                        Key:   aws.String("Name"),
                        Value: aws.String(instanceName),
                    },
                },
            },
        },
    }

    // Run the instance
    result, err := ec2Client.RunInstances(context.Background(), input)
    if err != nil {
        return nil, fmt.Errorf("could not create EC2 instance: %v", err)
    }

    // Extract the instance ID
    if len(result.Instances) == 0 {
        return nil, fmt.Errorf("no instances were created")
    }
    instanceID := *result.Instances[0].InstanceId

    fmt.Printf("EC2 Instance created successfully with ID: %s\n", instanceID)

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

// Lambda Function Creation
// const lambdaExecutionRoleARN = "arn:aws:iam::173939030599:role/LambdaExecutionRole"
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

// RDS Instance Creation
func CreateRDSInstance(dbName, instanceID, instanceClass, engine, username, password string, allocatedStorage int32, subnetGroupName string) (*rds.CreateDBInstanceOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %v", err)
	}
	rdsClient := rds.NewFromConfig(cfg)

	input := &rds.CreateDBInstanceInput{
		DBName:               aws.String(dbName),
		DBInstanceIdentifier: aws.String(instanceID),
		DBInstanceClass:      aws.String(instanceClass),
		Engine:               aws.String(engine),
		MasterUsername:       aws.String(username),
		MasterUserPassword:   aws.String(password),
		AllocatedStorage:     aws.Int32(allocatedStorage),
		DBSubnetGroupName:    aws.String(subnetGroupName), // Add the subnet group
	}

	result, err := rdsClient.CreateDBInstance(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("could not create RDS instance: %v", err)
	}
	return result, nil
}

// func CreateDynamoDBTable(tableName, region string, readCapacity, writeCapacity int64) (*dynamodb.CreateTableOutput, error) {
// 	// Load AWS configuration with the provided region
// 	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to load AWS config: %v", err)
// 	}

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

// CloudFront Distribution Creation with S3 Integration with OAI
func CreateCloudFrontDistribution(originDomainName, comment, region string, minTTL int64) (*cloudfront.CreateDistributionOutput, string, error) {
    cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
    if err != nil {
        return nil, "", fmt.Errorf("unable to load config: %v", err)
    }
    cloudFrontClient := cloudfront.NewFromConfig(cfg)

    // Create an Origin Access Identity (OAI)
    oaiResult, err := cloudFrontClient.CreateCloudFrontOriginAccessIdentity(context.TODO(), &cloudfront.CreateCloudFrontOriginAccessIdentityInput{
        CloudFrontOriginAccessIdentityConfig: &cloudfronttypes.CloudFrontOriginAccessIdentityConfig{
            CallerReference: aws.String(fmt.Sprintf("caller-ref-%d", time.Now().UnixNano())),
            Comment:         aws.String(comment),
        },
    })
    if err != nil {
        return nil, "", fmt.Errorf("could not create Origin Access Identity: %v", err)
    }

    // Retrieve the OAI Canonical User ID
    canonicalUserID := *oaiResult.CloudFrontOriginAccessIdentity.S3CanonicalUserId

    // Create the CloudFront distribution
    input := &cloudfront.CreateDistributionInput{
        DistributionConfig: &cloudfronttypes.DistributionConfig{
            CallerReference: aws.String(fmt.Sprintf("caller-ref-%d", time.Now().UnixNano())),
            Enabled:         aws.Bool(true),
            Comment:         aws.String(comment),
            Origins: &cloudfronttypes.Origins{
                Quantity: aws.Int32(1),
                Items: []cloudfronttypes.Origin{
                    {
                        Id:         aws.String("Origin1"),
                        DomainName: aws.String(originDomainName),
                        S3OriginConfig: &cloudfronttypes.S3OriginConfig{
                            OriginAccessIdentity: aws.String(fmt.Sprintf("origin-access-identity/cloudfront/%s", *oaiResult.CloudFrontOriginAccessIdentity.Id)),
                        },
                    },
                },
            },
            DefaultCacheBehavior: &cloudfronttypes.DefaultCacheBehavior{
                TargetOriginId:       aws.String("Origin1"),
                ViewerProtocolPolicy: cloudfronttypes.ViewerProtocolPolicyRedirectToHttps,
                AllowedMethods: &cloudfronttypes.AllowedMethods{
                    Quantity: aws.Int32(2),
                    Items:    []cloudfronttypes.Method{cloudfronttypes.MethodGet, cloudfronttypes.MethodHead},
                },
                ForwardedValues: &cloudfronttypes.ForwardedValues{
                    QueryString: aws.Bool(false),
                    Cookies: &cloudfronttypes.CookiePreference{
                        Forward: cloudfronttypes.ItemSelectionNone,
                    },
                },
                MinTTL: aws.Int64(minTTL),
            },
        },
    }

    result, err := cloudFrontClient.CreateDistribution(context.TODO(), input)
    if err != nil {
        return nil, "", fmt.Errorf("could not create CloudFront distribution: %v", err)
    }

    return result, canonicalUserID, nil
}

// CreateS3BucketWithPolicy creates an S3 bucket and attaches a policy to allow CloudFront access using OAI
func CreateS3BucketWithPolicy(bucketName, region, oaiCanonicalUserID string) (*s3.CreateBucketOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load config: %v", err)
	}
	s3Client := s3.NewFromConfig(cfg)

	// Create the bucket
	_, err = s3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create S3 bucket: %v", err)
	}

	// Attach bucket policy
	bucketPolicy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"CanonicalUser": "%s"
				},
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::%s/*"
			}
		]
	}`, oaiCanonicalUserID, bucketName)

	_, err = s3Client.PutBucketPolicy(context.TODO(), &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(bucketPolicy),
	})
	if err != nil {
		return nil, fmt.Errorf("could not attach bucket policy: %v", err)
	}

	return nil, nil
}

// VPC Creation
func CreateVPC(cidrBlock, region, name string) (*ec2.CreateVpcOutput, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("could not load AWS config: %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Create VPC
	input := &ec2.CreateVpcInput{
		CidrBlock: aws.String(cidrBlock),
	}

	result, err := ec2Client.CreateVpc(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("could not create VPC: %v", err)
	}

	// Add a Name tag to the VPC
	_, err = ec2Client.CreateTags(context.TODO(), &ec2.CreateTagsInput{
		Resources: []string{*result.Vpc.VpcId},
		Tags: []ec2types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not tag VPC: %v", err)
	}

	return result, nil
}
