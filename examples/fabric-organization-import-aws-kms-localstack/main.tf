terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# ============================================================================
# NODE 1: Provider for creating the source organization
# ============================================================================
provider "chainlaunch" {
  alias    = "node1"
  url      = var.node1_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# ============================================================================
# NODE 2: Provider for importing the organization
# ============================================================================
provider "chainlaunch" {
  alias    = "node2"
  url      = var.node2_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# ============================================================================
# LocalStack AWS Provider Configuration
# LocalStack emulates AWS services locally for testing
# ============================================================================
provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key
  secret_key = var.aws_secret_key

  endpoints {
    kms = var.localstack_endpoint
  }

  skip_credentials_validation = true
  skip_requesting_account_id  = true
}

# ============================================================================
# STEP 1: Create an organization in Node 1
# This organization will provide the certificates for import
# ============================================================================

# Get database provider on Node 1
data "chainlaunch_key_providers" "node1_database" {
  provider    = chainlaunch.node1
  filter_type = "DATABASE"
}

locals {
  node1_database_provider_id = data.chainlaunch_key_providers.node1_database.providers[0].id
}

# Create source organization in Node 1
resource "chainlaunch_fabric_organization" "source_org" {
  provider    = chainlaunch.node1
  msp_id      = "SourceOrgMSP"
  name        = "source-organization"
  description = "Source organization - certificates will be imported to Node 2"
  provider_id = local.node1_database_provider_id
}

# Fetch the created organization to get its certificates
data "chainlaunch_fabric_organization" "source_org_data" {
  provider = chainlaunch.node1
  id       = chainlaunch_fabric_organization.source_org.id
}

# ============================================================================
# STEP 2: Create AWS KMS Keys via LocalStack
# ============================================================================

# Create signing CA KMS key
resource "aws_kms_key" "sign_ca_key" {
  description             = "KMS key for signing CA in LocalStack"
  deletion_window_in_days = 7
  enable_key_rotation     = false

  tags = {
    Name = "fabric-sign-ca-key"
  }
}

# Create alias for signing key
resource "aws_kms_alias" "sign_ca_alias" {
  name          = "alias/fabric-sign-ca"
  target_key_id = aws_kms_key.sign_ca_key.key_id
}

# Create TLS CA KMS key
resource "aws_kms_key" "tls_ca_key" {
  description             = "KMS key for TLS CA in LocalStack"
  deletion_window_in_days = 7
  enable_key_rotation     = false

  tags = {
    Name = "fabric-tls-ca-key"
  }
}

# Create alias for TLS key
resource "aws_kms_alias" "tls_ca_alias" {
  name          = "alias/fabric-tls-ca"
  target_key_id = aws_kms_key.tls_ca_key.key_id
}

# ============================================================================
# STEP 3: Create AWS KMS Provider in Node 2
# This provider will manage keys in LocalStack's KMS
# ============================================================================

# Get database provider on Node 2 (for reference)
data "chainlaunch_key_providers" "node2_database" {
  provider    = chainlaunch.node2
  filter_type = "DATABASE"
}

# Create AWS KMS provider on Node 2
resource "chainlaunch_key_provider" "localstack_kms" {
  provider    = chainlaunch.node2
  name        = "localstack-kms-provider"
  type        = "AWS_KMS"
  description = "AWS KMS provider pointing to LocalStack for testing"

  awskms_config = {
    operation  = "IMPORT"
    region     = var.aws_region
    access_key = var.aws_access_key
    secret_key = var.aws_secret_key
    endpoint   = var.localstack_endpoint
  }

  depends_on = [
    aws_kms_key.sign_ca_key,
    aws_kms_key.tls_ca_key
  ]
}

# ============================================================================
# STEP 4: Import the source organization into Node 2 using KMS
# Uses certificates from Node 1 organization with LocalStack KMS keys
# ============================================================================

resource "chainlaunch_fabric_organization_import" "imported_org" {
  provider    = chainlaunch.node2
  msp_id      = "ImportedSourceOrgMSP"
  name        = "imported-from-node1"
  provider_id = tonumber(chainlaunch_key_provider.localstack_kms.id)
  source_type = "aws_kms"
  description = "Organization imported from Node 1 using LocalStack KMS"

  aws_kms_import {
    # Use the exact certificates from Node 1 organization
    sign_ca_cert   = data.chainlaunch_fabric_organization.source_org_data.sign_certificate
    sign_ca_key_id = aws_kms_key.sign_ca_key.arn

    tls_ca_cert   = data.chainlaunch_fabric_organization.source_org_data.tls_certificate
    tls_ca_key_id = aws_kms_key.tls_ca_key.arn
  }

  depends_on = [
    chainlaunch_key_provider.localstack_kms,
    aws_kms_alias.sign_ca_alias,
    aws_kms_alias.tls_ca_alias
  ]
}

# ============================================================================
# Outputs
# ============================================================================

# Node 1 Outputs
output "source_organization_id" {
  description = "ID of the source organization created in Node 1"
  value       = chainlaunch_fabric_organization.source_org.id
}

output "source_organization_msp_id" {
  description = "MSP ID of the source organization"
  value       = chainlaunch_fabric_organization.source_org.msp_id
}

output "source_organization_name" {
  description = "Name of the source organization"
  value       = chainlaunch_fabric_organization.source_org.name
}

output "source_sign_certificate" {
  description = "Signing CA certificate from Node 1 organization"
  value       = data.chainlaunch_fabric_organization.source_org_data.sign_certificate
  sensitive   = true
}

output "source_tls_certificate" {
  description = "TLS CA certificate from Node 1 organization"
  value       = data.chainlaunch_fabric_organization.source_org_data.tls_certificate
  sensitive   = true
}

# LocalStack KMS Outputs
output "sign_ca_key_id" {
  description = "LocalStack KMS Key ID for signing CA"
  value       = aws_kms_key.sign_ca_key.key_id
}

output "sign_ca_key_arn" {
  description = "LocalStack KMS Key ARN for signing CA"
  value       = aws_kms_key.sign_ca_key.arn
}

output "tls_ca_key_id" {
  description = "LocalStack KMS Key ID for TLS CA"
  value       = aws_kms_key.tls_ca_key.key_id
}

output "tls_ca_key_arn" {
  description = "LocalStack KMS Key ARN for TLS CA"
  value       = aws_kms_key.tls_ca_key.arn
}

# Node 2 Outputs
output "kms_provider_id" {
  description = "ID of the LocalStack KMS provider in Node 2"
  value       = chainlaunch_key_provider.localstack_kms.id
}

output "imported_organization_id" {
  description = "ID of the imported organization in Node 2"
  value       = chainlaunch_fabric_organization_import.imported_org.id
}

output "imported_organization_msp_id" {
  description = "MSP ID of the imported organization"
  value       = chainlaunch_fabric_organization_import.imported_org.msp_id
}

output "imported_organization_name" {
  description = "Name of the imported organization"
  value       = chainlaunch_fabric_organization_import.imported_org.name
}

# Verification Outputs
output "certificates_match" {
  description = "Verification that imported organization uses same certificates as source"
  value = {
    source_sign_cert   = data.chainlaunch_fabric_organization.source_org_data.sign_certificate
    imported_sign_cert = chainlaunch_fabric_organization_import.imported_org.sign_certificate
    match              = data.chainlaunch_fabric_organization.source_org_data.sign_certificate == chainlaunch_fabric_organization_import.imported_org.sign_certificate
  }
  sensitive = true
}
