terraform {
  required_providers {
    chainlaunch = {
      source  = "kfsoftware/chainlaunch"
      version = "~> 0.0"
    }
  }
}

provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# Get the default database key provider
data "chainlaunch_key_providers" "database" {
  filter_type = "DATABASE"
}

locals {
  database_provider_id = data.chainlaunch_key_providers.database.providers[0].id
}

# Example 1: Import organization with all certificates (including private keys)
# This allows generating new identities for the organization
resource "chainlaunch_fabric_organization_import" "org1_with_keys" {
  msp_id      = "Org1MSP"
  name        = "organization-1"
  provider_id = local.database_provider_id
  source_type = "raw"
  description = "Organization 1 imported with private keys - can generate new identities"

  raw_import {
    # Read certificates from files
    sign_ca_cert        = file("${path.module}/certs/org1-sign-ca.pem")
    sign_ca_private_key = file("${path.module}/certs/org1-sign-ca-key.pem")
    tls_ca_cert         = file("${path.module}/certs/org1-tls-ca.pem")
    tls_ca_private_key  = file("${path.module}/certs/org1-tls-ca-key.pem")
  }
}

# Example 2: Import organization with only certificates (no private keys)
# Use this when private keys are not available or managed elsewhere
resource "chainlaunch_fabric_organization_import" "org2_certs_only" {
  msp_id      = "Org2MSP"
  name        = "organization-2"
  provider_id = local.database_provider_id
  source_type = "raw"
  description = "Organization 2 imported with certificates only - no private keys available"

  raw_import {
    # Only public certificates, no private keys
    sign_ca_cert = file("${path.module}/certs/org2-sign-ca.pem")
    tls_ca_cert  = file("${path.module}/certs/org2-tls-ca.pem")
    # sign_ca_private_key and tls_ca_private_key omitted
  }
}

# Example 3: Create identities after import
# This demonstrates using the imported organization to create admin identities
resource "chainlaunch_fabric_identity" "org1_admin" {
  organization_id = chainlaunch_fabric_organization_import.org1_with_keys.id
  name            = "admin"
  role            = "admin"
  description     = "Admin identity for Organization 1"
  dns_names       = ["admin.org1.example.com"]
}

resource "chainlaunch_fabric_identity" "org1_client" {
  organization_id = chainlaunch_fabric_organization_import.org1_with_keys.id
  name            = "client"
  role            = "client"
  description     = "Client identity for Organization 1 applications"
  dns_names       = ["app.org1.example.com"]
}

# Output the imported organization IDs
output "org1_id" {
  description = "Organization 1 ID"
  value       = chainlaunch_fabric_organization_import.org1_with_keys.id
}

output "org2_id" {
  description = "Organization 2 ID"
  value       = chainlaunch_fabric_organization_import.org2_certs_only.id
}

output "org1_admin_identity_id" {
  description = "Admin identity ID for Organization 1"
  value       = chainlaunch_fabric_identity.org1_admin.id
}
