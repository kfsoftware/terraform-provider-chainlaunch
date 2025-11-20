variable "chainlaunch_url" {
  description = "URL of the Chainlaunch API"
  type        = string
  default     = "http://localhost:8100"
}

variable "chainlaunch_username" {
  description = "Chainlaunch username"
  type        = string
  default     = "admin"
  sensitive   = true
}

variable "chainlaunch_password" {
  description = "Chainlaunch password"
  type        = string
  default     = "admin123"
  sensitive   = true
}

variable "aws_region" {
  description = "AWS region for KMS keys"
  type        = string
  default     = "us-east-1"
}

variable "aws_account_id" {
  description = "AWS account ID (used for ARN construction)"
  type        = string
  sensitive   = true
}

variable "org1_sign_kms_key_id" {
  description = "KMS key ID for Org1 signing CA"
  type        = string
  sensitive   = true
}

variable "org1_tls_kms_key_id" {
  description = "KMS key ID for Org1 TLS CA"
  type        = string
  sensitive   = true
}

variable "org2_sign_kms_key_id" {
  description = "KMS key ID or ARN for Org2 signing CA"
  type        = string
  sensitive   = true
}

variable "org2_tls_kms_key_id" {
  description = "KMS key ID or ARN for Org2 TLS CA"
  type        = string
  sensitive   = true
}
