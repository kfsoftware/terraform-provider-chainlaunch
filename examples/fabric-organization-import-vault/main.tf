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

# Get the Vault key provider
data "chainlaunch_key_providers" "vault" {
  filter_type = "VAULT"
}

locals {
  vault_provider_id = data.chainlaunch_key_providers.vault.providers[0].id
}

# Example 1: Basic Vault import
resource "chainlaunch_fabric_organization_import" "org1_vault" {
  msp_id      = "Org1MSP"
  name        = "organization-1"
  provider_id = local.vault_provider_id
  source_type = "vault"
  description = "Organization 1 with certificates in Vault"

  vault_import {
    # Vault paths must include /data/ for KV v2 mounts
    sign_ca_path = "secret/data/fabric/org1-sign-ca"
    tls_ca_path  = "secret/data/fabric/org1-tls-ca"
  }
}

# Example 2: Import with custom Vault paths
resource "chainlaunch_fabric_organization_import" "org2_vault_custom" {
  msp_id      = "Org2MSP"
  name        = "organization-2"
  provider_id = local.vault_provider_id
  source_type = "vault"
  description = "Organization 2 with custom Vault paths"

  vault_import {
    sign_ca_path = "secret/data/organizations/org2/signing-ca"
    tls_ca_path  = "secret/data/organizations/org2/tls-ca"
  }
}

# Example 3: Import organization with Vault path from variable
resource "chainlaunch_fabric_organization_import" "org3_vault_variable" {
  msp_id      = "Org3MSP"
  name        = "organization-3"
  provider_id = local.vault_provider_id
  source_type = "vault"
  description = "Organization 3 with Vault paths from variables"

  vault_import {
    sign_ca_path = "${var.vault_base_path}/org3-sign-ca"
    tls_ca_path  = "${var.vault_base_path}/org3-tls-ca"
  }
}

# Example 4: Bulk import multiple organizations
locals {
  vault_organizations = {
    org4 = {
      msp_id       = "Org4MSP"
      name         = "organization-4"
      sign_ca_path = "secret/data/fabric/org4-sign-ca"
      tls_ca_path  = "secret/data/fabric/org4-tls-ca"
    }
    org5 = {
      msp_id       = "Org5MSP"
      name         = "organization-5"
      sign_ca_path = "secret/data/fabric/org5-sign-ca"
      tls_ca_path  = "secret/data/fabric/org5-tls-ca"
    }
  }
}

resource "chainlaunch_fabric_organization_import" "org_bulk_vault" {
  for_each = local.vault_organizations

  msp_id      = each.value.msp_id
  name        = each.value.name
  provider_id = local.vault_provider_id
  source_type = "vault"
  description = "Organization imported from Vault: ${each.value.name}"

  vault_import {
    sign_ca_path = each.value.sign_ca_path
    tls_ca_path  = each.value.tls_ca_path
  }
}

# Outputs
output "vault_provider_id" {
  description = "ID of the Vault key provider"
  value       = local.vault_provider_id
}

output "org1_vault_id" {
  description = "Organization 1 ID (from Vault)"
  value       = chainlaunch_fabric_organization_import.org1_vault.id
}

output "org2_vault_id" {
  description = "Organization 2 ID (from Vault)"
  value       = chainlaunch_fabric_organization_import.org2_vault_custom.id
}

output "org3_vault_id" {
  description = "Organization 3 ID (from Vault)"
  value       = chainlaunch_fabric_organization_import.org3_vault_variable.id
}

output "org_bulk_vault_ids" {
  description = "IDs of bulk-imported organizations from Vault"
  value = {
    for k, v in chainlaunch_fabric_organization_import.org_bulk_vault : k => v.id
  }
}
