terraform {
  required_providers {
    chainlaunch = {
      source = "registry.terraform.io/chainlaunch/chainlaunch"
    }
  }
}

# Configure the Chainlaunch Provider
provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# Variables
variable "chainlaunch_url" {
  description = "Chainlaunch API URL"
  type        = string
  default     = "http://localhost:8100"
}

variable "chainlaunch_username" {
  description = "Chainlaunch Username"
  type        = string
  default     = "admin"
}

variable "chainlaunch_password" {
  description = "Chainlaunch Password"
  type        = string
  sensitive   = true
  default     = "admin123"
}

# Create an AWS KMS Key Provider
resource "chainlaunch_key_provider" "aws_kms" {
  name       = "MyAWSKMSProvider"
  type       = "AWS_KMS"
  is_default = false

  aws_kms_config = {
    operation             = "IMPORT"
    aws_region            = "us-east-1"
    aws_access_key_id     = "test"                  # LocalStack test credentials
    aws_secret_access_key = "test"                  # LocalStack test credentials
    endpoint_url          = "http://localhost:4566" # LocalStack endpoint
    kms_key_alias_prefix  = "chainlaunch/"
  }
}

# Create an organization using the AWS KMS provider
resource "chainlaunch_fabric_organization" "with_aws_kms" {
  msp_id      = "AWSKMSOrgMSP"
  description = "Organization using AWS KMS for key management"
  provider_id = tonumber(chainlaunch_key_provider.aws_kms.id)
}

# Outputs
output "key_provider_id" {
  description = "The ID of the created AWS KMS key provider"
  value       = chainlaunch_key_provider.aws_kms.id
}

output "key_provider_type" {
  description = "The type of the key provider"
  value       = chainlaunch_key_provider.aws_kms.type
}

output "organization_id" {
  description = "The ID of the created organization"
  value       = chainlaunch_fabric_organization.with_aws_kms.id
}

output "organization_provider_id" {
  description = "The provider ID used by the organization"
  value       = chainlaunch_fabric_organization.with_aws_kms.provider_id
}
