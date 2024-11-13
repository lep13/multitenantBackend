# EC2 Variables
variable "ami_id" {
  type = string
}

variable "instance_type" {
  type = string
}

variable "key_name" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "security_group_ids" {
  type = list(string)
}

# S3 Variables
variable "bucket_name" {
  type = string
}

variable "enable_versioning" {
  type    = bool
  default = false
}

variable "region" {
  type    = string
  default = "us-east-1"
}

# Lambda Variables
variable "lambda_name" {
  type = string
}

variable "lambda_role_arn" {
  type = string
}

variable "lambda_handler" {
  type = string
}

variable "lambda_runtime" {
  type = string
}

variable "lambda_zip_path" {
  type = string
}

