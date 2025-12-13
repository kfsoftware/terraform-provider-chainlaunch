terraform {
  required_providers {
    chainlaunch = {
      source  = "kfsoftware/chainlaunch"
      version = "~> 0.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

provider "aws" {
  region = var.aws_region
}

# Get the AWS KMS key provider
data "chainlaunch_key_providers" "awskms" {
  filter_type = "AWS_KMS"
}

locals {
  awskms_provider_id = data.chainlaunch_key_providers.awskms.providers[0].id
}

# Example 1: Import organization with existing KMS keys
resource "chainlaunch_fabric_organization_import" "org1_kms" {
  msp_id      = "Org1MSP"
  name        = "organization-1"
  provider_id = local.awskms_provider_id
  source_type = "aws_kms"
  description = "Organization 1 with keys in AWS KMS"

  aws_kms_import {
    sign_ca_cert   = file("${path.module}/certs/org1-sign-ca.pem")
    sign_ca_key_id = "arn:aws:kms:${var.aws_region}:${var.aws_account_id}:key/${var.org1_sign_kms_key_id}"

    tls_ca_cert   = file("${path.module}/certs/org1-tls-ca.pem")
    tls_ca_key_id = "arn:aws:kms:${var.aws_region}:${var.aws_account_id}:key/${var.org1_tls_kms_key_id}"
  }
}

# Example 2: Import with short-form KMS key IDs
resource "chainlaunch_fabric_organization_import" "org2_kms_keyid" {
  msp_id      = "Org2MSP"
  name        = "organization-2"
  provider_id = local.awskms_provider_id
  source_type = "aws_kms"
  description = "Organization 2 with KMS key IDs"

  aws_kms_import {
    sign_ca_cert   = file("${path.module}/certs/org2-sign-ca.pem")
    sign_ca_key_id = var.org2_sign_kms_key_id

    tls_ca_cert   = file("${path.module}/certs/org2-tls-ca.pem")
    tls_ca_key_id = var.org2_tls_kms_key_id
  }
}

# Example 3: Create KMS keys and import organization
resource "aws_kms_key" "org3_sign_key" {
  description             = "KMS key for Org3 signing CA"
  deletion_window_in_days = 10
  enable_key_rotation     = true

  tags = {
    Name = "org3-sign-ca-key"
  }
}

resource "aws_kms_key" "org3_tls_key" {
  description             = "KMS key for Org3 TLS CA"
  deletion_window_in_days = 10
  enable_key_rotation     = true

  tags = {
    Name = "org3-tls-ca-key"
  }
}

resource "aws_kms_alias" "org3_sign_alias" {
  name          = "alias/org3-sign-ca"
  target_key_id = aws_kms_key.org3_sign_key.key_id
}

resource "aws_kms_alias" "org3_tls_alias" {
  name          = "alias/org3-tls-ca"
  target_key_id = aws_kms_key.org3_tls_key.key_id
}

resource "chainlaunch_fabric_organization_import" "org3_kms_created" {
  msp_id      = "Org3MSP"
  name        = "organization-3"
  provider_id = local.awskms_provider_id
  source_type = "aws_kms"
  description = "Organization 3 with newly created KMS keys"

  aws_kms_import {
    sign_ca_cert   = file("${path.module}/certs/org3-sign-ca.pem")
    sign_ca_key_id = aws_kms_key.org3_sign_key.arn

    tls_ca_cert   = file("${path.module}/certs/org3-tls-ca.pem")
    tls_ca_key_id = aws_kms_key.org3_tls_key.arn
  }

  depends_on = [
    aws_kms_alias.org3_sign_alias,
    aws_kms_alias.org3_tls_alias
  ]
}

# Outputs
output "awskms_provider_id" {
  description = "ID of the AWS KMS key provider"
  value       = local.awskms_provider_id
}

output "org1_kms_id" {
  description = "Organization 1 ID (KMS-backed)"
  value       = chainlaunch_fabric_organization_import.org1_kms.id
}

output "org2_kms_id" {
  description = "Organization 2 ID (KMS-backed)"
  value       = chainlaunch_fabric_organization_import.org2_kms_keyid.id
}

output "org3_kms_id" {
  description = "Organization 3 ID (KMS-backed with created keys)"
  value       = chainlaunch_fabric_organization_import.org3_kms_created.id
}

output "org3_kms_key_arns" {
  description = "ARNs of KMS keys created for Organization 3"
  value = {
    sign_key = aws_kms_key.org3_sign_key.arn
    tls_key  = aws_kms_key.org3_tls_key.arn
  }
}
