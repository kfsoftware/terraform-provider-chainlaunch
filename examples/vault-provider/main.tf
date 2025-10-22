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

# Example 1: Create a Vault provider (IMPORT mode - use existing Vault)
resource "chainlaunch_key_provider" "vault_existing" {
  name       = "ExistingVaultProvider"
  type       = "VAULT"
  is_default = false

  vault_config = {
    operation = "IMPORT"
    address   = "http://127.0.0.1:8200"
    token     = "root" # Vault root token or appropriate token
    mount     = "secret"
  }
}

# Example 2: Create a Vault provider (CREATE mode - Chainlaunch manages Vault)
resource "chainlaunch_key_provider" "vault_managed" {
  name       = "ManagedVaultProvider"
  type       = "VAULT"
  is_default = false

  vault_config = {
    operation = "CREATE"
    mode      = "docker" # Vault deployment mode: docker or service
    network   = "bridge" # Network mode: host or bridge
    port      = 8200     # Vault server port
  }
}

# Create an organization using the existing Vault provider
resource "chainlaunch_fabric_organization" "with_vault" {
  msp_id      = "VaultOrgMSP"
  description = "Organization using Hashicorp Vault for key management"
  provider_id = tonumber(chainlaunch_key_provider.vault_existing.id)
}

# Outputs
output "vault_existing_provider_id" {
  description = "The ID of the existing Vault provider"
  value       = chainlaunch_key_provider.vault_existing.id
}

output "vault_existing_provider_type" {
  description = "The type of the Vault provider"
  value       = chainlaunch_key_provider.vault_existing.type
}

output "vault_managed_provider_id" {
  description = "The ID of the managed Vault provider"
  value       = chainlaunch_key_provider.vault_managed.id
}

output "organization_id" {
  description = "The ID of the created organization"
  value       = chainlaunch_fabric_organization.with_vault.id
}

output "organization_provider_id" {
  description = "The provider ID used by the organization"
  value       = chainlaunch_fabric_organization.with_vault.provider_id
}
