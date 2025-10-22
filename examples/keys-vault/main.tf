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

# Resources

# Create HashiCorp Vault key provider (Chainlaunch manages Vault)
resource "chainlaunch_key_provider" "vault" {
  name       = "HashiCorp-Vault1"
  type       = "VAULT"
  is_default = false

  vault_config = {
    operation = "CREATE"
    mode      = "docker" # Vault deployment mode: docker or service
    network   = "bridge" # Network mode: host or bridge
    port      = 8265     # Vault server port
    version   = "1.20.2" # Vault version to deploy
  }
}

# Create various types of keys using HashiCorp Vault
# Note: Vault does NOT support secp256k1 curve

# RSA 2048-bit key
resource "chainlaunch_key" "rsa_2048" {
  name        = "vault-rsa-2048"
  algorithm   = "RSA"
  key_size    = 2048
  provider_id = tonumber(chainlaunch_key_provider.vault.id)
  description = "RSA 2048-bit key stored in Vault"
  is_ca       = false
}

# RSA 4096-bit key
resource "chainlaunch_key" "rsa_4096" {
  name        = "vault-rsa-4096"
  algorithm   = "RSA"
  key_size    = 4096
  provider_id = tonumber(chainlaunch_key_provider.vault.id)
  description = "RSA 4096-bit key stored in Vault"
  is_ca       = false
}

# EC P-256 key (NIST curve)
resource "chainlaunch_key" "ec_p256" {
  name        = "vault-ec-p256"
  algorithm   = "EC"
  curve       = "P-256"
  provider_id = tonumber(chainlaunch_key_provider.vault.id)
  description = "EC P-256 key stored in Vault"
  is_ca       = false
}

# EC P-384 key (NIST curve)
resource "chainlaunch_key" "ec_p384" {
  name        = "vault-ec-p384"
  algorithm   = "EC"
  curve       = "P-384"
  provider_id = tonumber(chainlaunch_key_provider.vault.id)
  description = "EC P-384 key stored in Vault"
  is_ca       = false
}

# EC P-521 key (NIST curve)
resource "chainlaunch_key" "ec_p521" {
  name        = "vault-ec-p521"
  algorithm   = "EC"
  curve       = "P-521"
  provider_id = tonumber(chainlaunch_key_provider.vault.id)
  description = "EC P-521 key stored in Vault"
  is_ca       = false
}

# NOTE: secp256k1 is NOT supported by HashiCorp Vault
# Attempting to create a secp256k1 key with Vault will fail
# Use AWS KMS or database provider for secp256k1 keys

# Certificate Authority key (RSA 4096)
resource "chainlaunch_key" "ca_rsa" {
  name        = "vault-ca-rsa"
  algorithm   = "RSA"
  key_size    = 4096
  provider_id = tonumber(chainlaunch_key_provider.vault.id)
  description = "Certificate Authority RSA key stored in Vault"
  is_ca       = true
}

# Outputs

output "vault_provider" {
  description = "Vault provider details"
  value = {
    id   = chainlaunch_key_provider.vault.id
    name = chainlaunch_key_provider.vault.name
    type = chainlaunch_key_provider.vault.type
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

output "summary" {
  description = "Summary of created keys"
  value = {
    provider = "HashiCorp Vault"
    keys_created = [
      "RSA 2048",
      "RSA 4096",
      "EC P-256",
      "EC P-384",
      "EC P-521",
      "CA RSA 4096"
    ]
    total_keys          = 6
    secp256k1_supported = false
    note                = "Vault does not support secp256k1 curve - use AWS KMS or database provider for blockchain keys"
  }
}
