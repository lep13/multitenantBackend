# Dropdown Data
output "available_instance_types" {
  value       = data.aws_ec2_instance_type_offerings.available_types.instance_types
  description = "List of available EC2 instance types for dropdown"
}

output "available_subnets" {
  value       = data.aws_subnets.all.ids
  description = "List of available subnets for dropdown"
}

output "available_security_groups" {
  value       = data.aws_security_groups.all.ids
  description = "List of available security groups for dropdown"
}

output "available_amis" {
  value       = data.aws_ami.amazon_linux.id
  description = "List of available AMIs for dropdown"
}

output "lambda_supported_runtimes" {
  value       = local.lambda_supported_runtimes
  description = "List of supported Lambda runtimes for dropdown"
}






# Machine Types
output "available_machine_types" {
  value       = data.google_compute_machine_types.types.names
  description = "List of available machine types for dropdown"
}

# Zones
output "available_zones" {
  value       = data.google_compute_zones.available_zones.names
  description = "List of available zones for dropdown"
}

# Images
output "debian_image" {
  value       = data.google_compute_image.debian.self_link
  description = "Debian 11 image link for dropdown"
}

# Networks
output "available_networks" {
  value       = data.google_compute_networks.networks.names
  description = "List of available networks for dropdown"
}

# Subnetworks
output "available_subnetworks" {
  value       = data.google_compute_subnetworks.subnetworks.names
  description = "List of available subnetworks for dropdown"
}
