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

# Data Sources

# Fetch all key providers
data "chainlaunch_key_providers" "all" {}

# Fetch only database key providers
data "chainlaunch_key_providers" "database" {
  type_filter = "database"
}

# Fetch existing organizations (commented out - update ID to match an existing organization)
# data "chainlaunch_fabric_organization" "existing" {
#   id = "1"
# }

# Resources

# Create multiple organizations using different approaches

# Organization 1: Using default key provider
resource "chainlaunch_fabric_organization" "org1" {
  msp_id      = "CompleteOrg1MSP"
  description = "First organization using default provider"
  provider_id = data.chainlaunch_key_providers.all.default_provider_id
}

# Organization 2: Using database provider specifically
resource "chainlaunch_fabric_organization" "org2" {
  msp_id      = "CompleteOrg2MSP"
  description = "Second organization using database provider"
  provider_id = tonumber(data.chainlaunch_key_providers.database.providers[0].id)
}

# Outputs

output "all_key_providers" {
  description = "All available key providers"
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

# Output for existing organization (commented out - uncomment when data source is enabled)
# output "existing_organization" {
#   description = "Details of existing organization"
#   value = {
#     id          = data.chainlaunch_fabric_organization.existing.id
#     msp_id      = data.chainlaunch_fabric_organization.existing.msp_id
#     description = data.chainlaunch_fabric_organization.existing.description
#   }
# }

output "org1_details" {
  description = "Created organization 1 details"
  value = {
    id          = chainlaunch_fabric_organization.org1.id
    msp_id      = chainlaunch_fabric_organization.org1.msp_id
    provider_id = chainlaunch_fabric_organization.org1.provider_id
  }
}

output "org2_details" {
  description = "Created organization 2 details"
  value = {
    id          = chainlaunch_fabric_organization.org2.id
    msp_id      = chainlaunch_fabric_organization.org2.msp_id
    provider_id = chainlaunch_fabric_organization.org2.provider_id
  }
}
