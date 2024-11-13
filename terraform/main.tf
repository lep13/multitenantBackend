provider "aws" {
  region = "us-east-1" # Default region; you can override with specific region
}

# Fetch available EC2 instance types
data "aws_ec2_instance_type_offerings" "available_types" {
  location_type = "availability-zone"
}

# Fetch available subnets
data "aws_subnets" "all" {}

# Fetch available security groups
data "aws_security_groups" "all" {}

# Fetch available key pairs
data "aws_key_pairs" "all" {}

# Fetch available AMIs (example: Amazon Linux 2)
data "aws_ami" "amazon_linux" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*"]
  }

  owners = ["amazon"] # Filter by Amazon-owned AMIs
}

# Fetch supported Lambda runtimes
data "aws_lambda_runtimes" "supported" {}

# EC2 Instance Configuration
resource "aws_instance" "ec2_instance" {
  ami           = var.ami_id
  instance_type = var.instance_type
  key_name      = var.key_name
  subnet_id     = var.subnet_id
  vpc_security_group_ids = var.security_group_ids

  tags = {
    Name = "MultitenantAppEC2Instance"
  }
}

# S3 Bucket Configuration
resource "aws_s3_bucket" "example_bucket" {
  bucket = var.bucket_name
  acl    = "private"

  versioning {
    enabled = var.enable_versioning
  }

  tags = {
    Name = "MultitenantAppS3Bucket"
  }
}

# Lambda Function Configuration
resource "aws_lambda_function" "example_lambda" {
  function_name = var.lambda_name
  role          = var.lambda_role_arn
  handler       = var.lambda_handler
  runtime       = var.lambda_runtime
  filename      = var.lambda_zip_path

  tags = {
    Name = "MultitenantAppLambda"
  }
}

