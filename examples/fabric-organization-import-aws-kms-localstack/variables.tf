variable "node1_url" {
  description = "URL of Node 1 Chainlaunch API (where organizations are created)"
  type        = string
  default     = "http://localhost:8100"
}

variable "node2_url" {
  description = "URL of Node 2 Chainlaunch API (where organizations are imported)"
  type        = string
  default     = "http://localhost:8101"
}

variable "chainlaunch_username" {
  description = "Chainlaunch username (same for both nodes)"
  type        = string
  default     = "admin"
  sensitive   = true
}

variable "chainlaunch_password" {
  description = "Chainlaunch password (same for both nodes)"
  type        = string
  default     = "admin123"
  sensitive   = true
}

variable "aws_region" {
  description = "AWS region (LocalStack emulates this)"
  type        = string
  default     = "us-east-1"
}

variable "localstack_endpoint" {
  description = "LocalStack endpoint for KMS service"
  type        = string
  default     = "http://localhost:4566"
}

variable "aws_access_key" {
  description = "AWS access key for LocalStack (test/test for LocalStack)"
  type        = string
  default     = "test"
  sensitive   = true
}

variable "aws_secret_key" {
  description = "AWS secret key for LocalStack (test/test for LocalStack)"
  type        = string
  default     = "test"
  sensitive   = true
}
