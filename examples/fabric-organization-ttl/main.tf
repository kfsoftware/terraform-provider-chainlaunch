terraform {
  required_providers {
    chainlaunch = {
      source = "kfsoftware/chainlaunch"
    }
  }
}

provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# Get the default database provider
data "chainlaunch_key_providers" "database" {
  type_filter = "DATABASE"
}

locals {
  database_provider_id = data.chainlaunch_key_providers.database.default_provider_id
}

# ============================================================================
# Organization with default TTL settings
# ============================================================================
resource "chainlaunch_fabric_organization" "org1_default" {
  msp_id       = "Org1DefaultTTLMSP"
  description  = "Organization with default TTL settings"
  provider_id  = local.database_provider_id
}

# ============================================================================
# Organization with custom CA certificate TTL (10 years)
# ============================================================================
resource "chainlaunch_fabric_organization" "org2_long_ca_ttl" {
  msp_id              = "Org2LongCATTLMSP"
  description         = "Organization with long-lived CA certificates (10 years)"
  provider_id         = local.database_provider_id
  ca_cert_valid_for   = "87600h"  # 10 years (365 days * 24 hours * 10 years)
}

# ============================================================================
# Organization with custom certificate TTL (2 years)
# ============================================================================
resource "chainlaunch_fabric_organization" "org3_long_cert_ttl" {
  msp_id              = "Org3LongCertTTLMSP"
  description         = "Organization with long-lived certificates (2 years)"
  provider_id         = local.database_provider_id
  cert_valid_for      = "17520h"  # 2 years (365 days * 24 hours * 2 years)
}

# ============================================================================
# Organization with both CA and certificate TTL settings
# ============================================================================
resource "chainlaunch_fabric_organization" "org4_custom_ttl" {
  msp_id              = "Org4CustomTTLMSP"
  description         = "Organization with custom CA and certificate TTL"
  provider_id         = local.database_provider_id
  ca_cert_valid_for   = "175200h"  # 20 years
  cert_valid_for      = "8760h"    # 1 year (default but explicitly set)
}

# ============================================================================
# Organization with short TTL for development/testing
# ============================================================================
resource "chainlaunch_fabric_organization" "org5_short_ttl" {
  msp_id              = "Org5DevTTLMSP"
  description         = "Organization with short TTL for development (30 days)"
  provider_id         = local.database_provider_id
  cert_valid_for      = "720h"     # 30 days (30 days * 24 hours)
}

# ============================================================================
# Outputs
# ============================================================================

output "org1_id" {
  description = "Organization 1 ID (default TTL)"
  value       = chainlaunch_fabric_organization.org1_default.id
}

output "org2_id" {
  description = "Organization 2 ID (10-year CA TTL)"
  value       = chainlaunch_fabric_organization.org2_long_ca_ttl.id
}

output "org2_ca_cert_valid_for" {
  description = "CA certificate validity for Organization 2"
  value       = chainlaunch_fabric_organization.org2_long_ca_ttl.ca_cert_valid_for
}

output "org3_id" {
  description = "Organization 3 ID (2-year certificate TTL)"
  value       = chainlaunch_fabric_organization.org3_long_cert_ttl.id
}

output "org3_cert_valid_for" {
  description = "Certificate validity for Organization 3"
  value       = chainlaunch_fabric_organization.org3_long_cert_ttl.cert_valid_for
}

output "org4_id" {
  description = "Organization 4 ID (custom TTL)"
  value       = chainlaunch_fabric_organization.org4_custom_ttl.id
}

output "org4_ca_ttl" {
  description = "CA certificate TTL for Organization 4"
  value       = chainlaunch_fabric_organization.org4_custom_ttl.ca_cert_valid_for
}

output "org4_cert_ttl" {
  description = "Certificate TTL for Organization 4"
  value       = chainlaunch_fabric_organization.org4_custom_ttl.cert_valid_for
}

output "org5_id" {
  description = "Organization 5 ID (short TTL for development)"
  value       = chainlaunch_fabric_organization.org5_short_ttl.id
}

output "org5_cert_valid_for" {
  description = "Certificate validity for Organization 5 (development)"
  value       = chainlaunch_fabric_organization.org5_short_ttl.cert_valid_for
}
