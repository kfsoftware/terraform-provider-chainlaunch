terraform {
  required_providers {
    chainlaunch = {
      source = "registry.terraform.io/chainlaunch/chainlaunch"
    }
  }
}

provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

variable "chainlaunch_url" {
  type    = string
  default = "http://localhost:8100"
}

variable "chainlaunch_username" {
  type    = string
  default = "admin"
}

variable "chainlaunch_password" {
  type      = string
  sensitive = true
  default   = "admin123"
}

# Fetch all key providers
data "chainlaunch_key_providers" "all" {}

# Fetch only database key providers
data "chainlaunch_key_providers" "database" {
  type_filter = "database"
}

# Fetch providers by name (partial match)
data "chainlaunch_key_providers" "by_name" {
  name_filter = "Default"
}

# Create an organization using the default key provider
resource "chainlaunch_fabric_organization" "example" {
  msp_id      = "ExampleOrgMSP"
  description = "Organization using default key provider"

  # Reference the default provider ID from the data source
  provider_id = data.chainlaunch_key_providers.all.default_provider_id
}

# Outputs
output "all_providers" {
  description = "All key providers"
  value       = data.chainlaunch_key_providers.all.providers
}

output "default_provider" {
  description = "Default key provider details"
  value = {
    id   = data.chainlaunch_key_providers.all.default_provider_id
    name = data.chainlaunch_key_providers.all.default_provider_name
    type = data.chainlaunch_key_providers.all.default_provider_type
  }
}

output "database_providers" {
  description = "Database key providers only"
  value       = data.chainlaunch_key_providers.database.providers
}

output "providers_by_name" {
  description = "Providers matching name filter"
  value       = data.chainlaunch_key_providers.by_name.providers
}

# Example: Using the default provider in a locals block
locals {
  default_provider_id  = data.chainlaunch_key_providers.all.default_provider_id
  has_default_provider = data.chainlaunch_key_providers.all.default_provider_id != null
}

output "using_default_provider" {
  value = local.has_default_provider ? "Default provider ID: ${local.default_provider_id}" : "No default provider found"
}
