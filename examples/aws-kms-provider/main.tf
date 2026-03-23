terraform {
  required_providers {
    chainlaunch = {
      source = "registry.terraform.io/kfsoftware/chainlaunch"
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

# ---------------------------------------------------------------------------
# Example 1: EC2 Instance Role / EKS IRSA (recommended for production)
#
# When ChainLaunch runs on EC2 or EKS, omit aws_access_key_id and
# aws_secret_access_key entirely. The AWS SDK picks up credentials
# automatically from the instance metadata service or IRSA token.
#
# Prerequisites:
#   - EC2 instance profile or EKS service account with kms:* permissions
#   - ChainLaunch server running on that EC2/EKS workload
# ---------------------------------------------------------------------------
resource "chainlaunch_key_provider" "aws_kms_iam_role" {
  name       = "aws-kms-iam-role"
  type       = "AWS_KMS"
  is_default = false

  aws_kms_config = {
    operation            = "IMPORT"
    aws_region           = "us-east-1"
    kms_key_alias_prefix = "chainlaunch/"
    # No credentials — SDK uses instance role / IRSA automatically
  }
}

# ---------------------------------------------------------------------------
# Example 2: STS AssumeRole (cross-account or least-privilege)
#
# ChainLaunch assumes a dedicated IAM role via STS. Works with both
# instance role base credentials and static credentials.
#
# Prerequisites:
#   - Target role trust policy allows the ChainLaunch principal to assume it
#   - Target role has kms:* permissions on the required keys
# ---------------------------------------------------------------------------
resource "chainlaunch_key_provider" "aws_kms_assume_role" {
  name       = "aws-kms-assume-role"
  type       = "AWS_KMS"
  is_default = false

  aws_kms_config = {
    operation            = "IMPORT"
    aws_region           = "us-east-1"
    assume_role_arn      = "arn:aws:iam::123456789012:role/ChainLaunchKMSRole"
    external_id          = "chainlaunch-unique-id" # optional, for extra security
    kms_key_alias_prefix = "chainlaunch/"
    # No static credentials — assumes role using instance role as base
  }
}

# ---------------------------------------------------------------------------
# Example 3: Static credentials (development / LocalStack only)
#
# Provide explicit access key and secret key. Avoid in production —
# use IAM roles instead.
# ---------------------------------------------------------------------------
resource "chainlaunch_key_provider" "aws_kms_static" {
  name       = "aws-kms-static-creds"
  type       = "AWS_KMS"
  is_default = false

  aws_kms_config = {
    operation             = "IMPORT"
    aws_region            = "us-east-1"
    aws_access_key_id     = "test"                 # LocalStack test credentials
    aws_secret_access_key = "test"                 # LocalStack test credentials
    endpoint_url          = "http://localhost:4566" # LocalStack endpoint
    kms_key_alias_prefix  = "chainlaunch/"
  }
}

# ---------------------------------------------------------------------------
# Example 4: Static credentials + AssumeRole (cross-account from outside AWS)
#
# When ChainLaunch runs outside AWS (on-prem, other cloud), provide
# base credentials that are allowed to assume the target role.
# ---------------------------------------------------------------------------
resource "chainlaunch_key_provider" "aws_kms_static_assume_role" {
  name       = "aws-kms-static-assume-role"
  type       = "AWS_KMS"
  is_default = false

  aws_kms_config = {
    operation             = "IMPORT"
    aws_region            = "us-east-1"
    aws_access_key_id     = var.aws_access_key_id
    aws_secret_access_key = var.aws_secret_access_key
    assume_role_arn       = "arn:aws:iam::123456789012:role/ChainLaunchKMSRole"
    external_id           = "chainlaunch-unique-id"
    kms_key_alias_prefix  = "chainlaunch/"
  }
}

variable "aws_access_key_id" {
  description = "AWS access key ID (only for non-AWS environments)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "aws_secret_access_key" {
  description = "AWS secret access key (only for non-AWS environments)"
  type        = string
  sensitive   = true
  default     = ""
}

# Create an organization using the IAM role provider
resource "chainlaunch_fabric_organization" "with_aws_kms" {
  msp_id      = "AWSKMSOrgMSP"
  description = "Organization using AWS KMS for key management"
  provider_id = tonumber(chainlaunch_key_provider.aws_kms_iam_role.id)
}

# Outputs
output "iam_role_provider_id" {
  description = "The ID of the IAM role-based AWS KMS key provider"
  value       = chainlaunch_key_provider.aws_kms_iam_role.id
}

output "assume_role_provider_id" {
  description = "The ID of the AssumeRole-based AWS KMS key provider"
  value       = chainlaunch_key_provider.aws_kms_assume_role.id
}

output "static_provider_id" {
  description = "The ID of the static credentials AWS KMS key provider"
  value       = chainlaunch_key_provider.aws_kms_static.id
}

output "organization_id" {
  description = "The ID of the created organization"
  value       = chainlaunch_fabric_organization.with_aws_kms.id
}
