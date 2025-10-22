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

# Data Sources

# Fetch the default key provider (Database)
data "chainlaunch_key_providers" "default" {}

# Resources - Create various types of keys using the database provider

# RSA 2048-bit key
resource "chainlaunch_key" "rsa_2048" {
  name        = "database-rsa-2048"
  algorithm   = "RSA"
  key_size    = 2048
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
  description = "RSA 2048-bit key stored in database"
  is_ca       = false
}

# RSA 4096-bit key
resource "chainlaunch_key" "rsa_4096" {
  name        = "database-rsa-4096"
  algorithm   = "RSA"
  key_size    = 4096
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
  description = "RSA 4096-bit key stored in database"
  is_ca       = false
}

# EC P-256 key (NIST curve)
resource "chainlaunch_key" "ec_p256" {
  name        = "database-ec-p256"
  algorithm   = "EC"
  curve       = "P-256"
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
  description = "EC P-256 key stored in database"
  is_ca       = false
}

# EC P-384 key (NIST curve)
resource "chainlaunch_key" "ec_p384" {
  name        = "database-ec-p384"
  algorithm   = "EC"
  curve       = "P-384"
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
  description = "EC P-384 key stored in database"
  is_ca       = false
}

# EC P-521 key (NIST curve)
resource "chainlaunch_key" "ec_p521" {
  name        = "database-ec-p521"
  algorithm   = "EC"
  curve       = "P-521"
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
  description = "EC P-521 key stored in database"
  is_ca       = false
}

# EC secp256k1 key (Bitcoin/Ethereum curve)
resource "chainlaunch_key" "ec_secp256k1" {
  name        = "database-ec-secp256k1"
  algorithm   = "EC"
  curve       = "secp256k1"
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
  description = "EC secp256k1 key stored in database"
  is_ca       = false
}

# ED25519 key
resource "chainlaunch_key" "ed25519" {
  name        = "database-ed25519"
  algorithm   = "ED25519"
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
  description = "ED25519 key stored in database"
  is_ca       = false
}

# Certificate Authority key (RSA 4096)
resource "chainlaunch_key" "ca_rsa" {
  name        = "database-ca-rsa"
  algorithm   = "RSA"
  key_size    = 4096
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
  description = "Certificate Authority RSA key"
  is_ca       = true
}

# Outputs

output "default_provider_details" {
  description = "Details of the default key provider"
  value = {
    id   = data.chainlaunch_key_providers.default.default_provider_id
    name = data.chainlaunch_key_providers.default.default_provider_name
    type = data.chainlaunch_key_providers.default.default_provider_type
  }
}

output "rsa_2048_key" {
  description = "RSA 2048 key details"
  value = {
    id         = chainlaunch_key.rsa_2048.id
    name       = chainlaunch_key.rsa_2048.name
    algorithm  = chainlaunch_key.rsa_2048.algorithm
    key_size   = chainlaunch_key.rsa_2048.key_size
    created_at = chainlaunch_key.rsa_2048.created_at
  }
}

output "rsa_4096_key" {
  description = "RSA 4096 key details"
  value = {
    id         = chainlaunch_key.rsa_4096.id
    name       = chainlaunch_key.rsa_4096.name
    algorithm  = chainlaunch_key.rsa_4096.algorithm
    key_size   = chainlaunch_key.rsa_4096.key_size
    created_at = chainlaunch_key.rsa_4096.created_at
  }
}

output "ec_p256_key" {
  description = "EC P-256 key details"
  value = {
    id         = chainlaunch_key.ec_p256.id
    name       = chainlaunch_key.ec_p256.name
    algorithm  = chainlaunch_key.ec_p256.algorithm
    curve      = chainlaunch_key.ec_p256.curve
    created_at = chainlaunch_key.ec_p256.created_at
  }
}

output "ec_p384_key" {
  description = "EC P-384 key details"
  value = {
    id         = chainlaunch_key.ec_p384.id
    name       = chainlaunch_key.ec_p384.name
    algorithm  = chainlaunch_key.ec_p384.algorithm
    curve      = chainlaunch_key.ec_p384.curve
    created_at = chainlaunch_key.ec_p384.created_at
  }
}

output "ec_p521_key" {
  description = "EC P-521 key details"
  value = {
    id         = chainlaunch_key.ec_p521.id
    name       = chainlaunch_key.ec_p521.name
    algorithm  = chainlaunch_key.ec_p521.algorithm
    curve      = chainlaunch_key.ec_p521.curve
    created_at = chainlaunch_key.ec_p521.created_at
  }
}

output "ec_secp256k1_key" {
  description = "EC secp256k1 key details"
  value = {
    id         = chainlaunch_key.ec_secp256k1.id
    name       = chainlaunch_key.ec_secp256k1.name
    algorithm  = chainlaunch_key.ec_secp256k1.algorithm
    curve      = chainlaunch_key.ec_secp256k1.curve
    created_at = chainlaunch_key.ec_secp256k1.created_at
  }
}

output "ed25519_key" {
  description = "ED25519 key details"
  value = {
    id         = chainlaunch_key.ed25519.id
    name       = chainlaunch_key.ed25519.name
    algorithm  = chainlaunch_key.ed25519.algorithm
    created_at = chainlaunch_key.ed25519.created_at
  }
}

output "ca_key" {
  description = "Certificate Authority key details"
  value = {
    id         = chainlaunch_key.ca_rsa.id
    name       = chainlaunch_key.ca_rsa.name
    algorithm  = chainlaunch_key.ca_rsa.algorithm
    key_size   = chainlaunch_key.ca_rsa.key_size
    is_ca      = chainlaunch_key.ca_rsa.is_ca
    created_at = chainlaunch_key.ca_rsa.created_at
  }
}
