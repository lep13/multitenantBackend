provider "aws" {
  region = "us-east-1" # Default region
}

# Fetch available EC2 instance types
data "aws_ec2_instance_type_offerings" "available_types" {
  location_type = "availability-zone"
}

# Fetch available subnets
data "aws_subnets" "all" {}

# Fetch available security groups
data "aws_security_groups" "all" {}

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
locals {
  lambda_supported_runtimes = [
    "nodejs14.x",
    "nodejs16.x",
    "python3.9",
    "python3.8",
    "go1.x",
    "java11",
    "java8",
    "ruby2.7",
    "dotnet6"
  ]
}




provider "google" {
  project = "gifted-fragment-436605-u0"
  region  = "us-east1"
  zone    = "us-east1-b"
}

# Fetch available machine types in the zone
data "google_compute_machine_types" "types" {
  zone = "us-east1-b"
}

# Fetch available zones in the region
data "google_compute_zones" "available_zones" {}

# Fetch public images (e.g., Debian)
data "google_compute_image" "debian

data "google_compute_image" "debian" {
  family  = "debian-11"
  project = "debian-cloud"
}

# Fetch available networks
data "google_compute_networks" "networks" {}

# Fetch available subnetworks for the selected network
data "google_compute_subnetworks" "subnetworks" {
  region = "us-east1"
}
