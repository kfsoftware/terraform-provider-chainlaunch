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

# Create AWS KMS key provider (using LocalStack for local testing)
resource "chainlaunch_key_provider" "aws_kms" {
  name       = "LocalStack-KMS"
  type       = "AWS_KMS"
  is_default = false

  aws_kms_config = {
    operation             = "IMPORT"
    aws_region            = "us-east-1"
    aws_access_key_id     = "test"                  # LocalStack test credentials
    aws_secret_access_key = "test"                  # LocalStack test credentials
    endpoint_url          = "http://localhost:4566" # LocalStack endpoint
    kms_key_alias_prefix  = "chainlaunch/"
  }
}

# Create various types of keys using AWS KMS

# RSA 2048-bit key
resource "chainlaunch_key" "rsa_2048" {
  name        = "kms-rsa-2048"
  algorithm   = "RSA"
  key_size    = 2048
  provider_id = tonumber(chainlaunch_key_provider.aws_kms.id)
  description = "RSA 2048-bit key stored in AWS KMS"
  is_ca       = false
}

# RSA 4096-bit key
resource "chainlaunch_key" "rsa_4096" {
  name        = "kms-rsa-4096"
  algorithm   = "RSA"
  key_size    = 4096
  provider_id = tonumber(chainlaunch_key_provider.aws_kms.id)
  description = "RSA 4096-bit key stored in AWS KMS"
  is_ca       = false
}

# EC P-256 key (NIST curve)
resource "chainlaunch_key" "ec_p256" {
  name        = "kms-ec-p256"
  algorithm   = "EC"
  curve       = "P-256"
  provider_id = tonumber(chainlaunch_key_provider.aws_kms.id)
  description = "EC P-256 key stored in AWS KMS"
  is_ca       = false
}

# EC P-384 key (NIST curve)
resource "chainlaunch_key" "ec_p384" {
  name        = "kms-ec-p384"
  algorithm   = "EC"
  curve       = "P-384"
  provider_id = tonumber(chainlaunch_key_provider.aws_kms.id)
  description = "EC P-384 key stored in AWS KMS"
  is_ca       = false
}

# EC P-521 key (NIST curve)
resource "chainlaunch_key" "ec_p521" {
  name        = "kms-ec-p521"
  algorithm   = "EC"
  curve       = "P-521"
  provider_id = tonumber(chainlaunch_key_provider.aws_kms.id)
  description = "EC P-521 key stored in AWS KMS"
  is_ca       = false
}

# EC secp256k1 key (Bitcoin/Ethereum curve)
# Note: AWS KMS supports secp256k1
resource "chainlaunch_key" "ec_secp256k1" {
  name        = "kms-ec-secp256k1"
  algorithm   = "EC"
  curve       = "secp256k1"
  provider_id = tonumber(chainlaunch_key_provider.aws_kms.id)
  description = "EC secp256k1 key stored in AWS KMS (blockchain curve)"
  is_ca       = false
}

# Outputs

output "aws_kms_provider" {
  description = "AWS KMS provider details"
  value = {
    id   = chainlaunch_key_provider.aws_kms.id
    name = chainlaunch_key_provider.aws_kms.name
    type = chainlaunch_key_provider.aws_kms.type
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

output "summary" {
  description = "Summary of created keys"
  value = {
    provider = "AWS KMS (LocalStack)"
    keys_created = [
      "RSA 2048",
      "RSA 4096",
      "EC P-256",
      "EC P-384",
      "EC P-521",
      "EC secp256k1"
    ]
    total_keys = 6
  }
}
