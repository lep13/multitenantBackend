# EC2 Instance Outputs
output "ec2_instance_id" {
  value = aws_instance.ec2_instance.id
  description = "The ID of the created EC2 instance"
}

output "ec2_public_ip" {
  value = aws_instance.ec2_instance.public_ip
  description = "The public IP address of the created EC2 instance"
}

# Dropdown Data
# Instance Types
output "available_instance_types" {
  value = data.aws_ec2_instance_type_offerings.available_types.instance_types
  description = "List of available EC2 instance types for dropdown"
}

# Subnets
output "available_subnets" {
  value = data.aws_subnets.all.ids
  description = "List of available subnets for dropdown"
}

# Security Groups
output "available_security_groups" {
  value = data.aws_security_groups.all.ids
  description = "List of available security groups for dropdown"
}

# Key Pairs
output "available_key_pairs" {
  value = data.aws_key_pairs.all.key_names
  description = "List of available key pairs for dropdown"
}

# S3 Bucket Outputs
output "s3_bucket_name" {
  value = aws_s3_bucket.example_bucket.bucket
  description = "The name of the created S3 bucket"
}

output "s3_bucket_arn" {
  value = aws_s3_bucket.example_bucket.arn
  description = "The ARN of the created S3 bucket"
}

# Lambda Outputs
output "lambda_function_name" {
  value = aws_lambda_function.example_lambda.function_name
  description = "The name of the created Lambda function"
}

# Lambda Runtimes
output "lambda_supported_runtimes" {
  value = data.aws_lambda_runtimes.supported.runtimes
  description = "List of supported Lambda runtimes for dropdown"
}
