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

variable "vault_base_path" {
  description = "Base path in Vault for organization certificates"
  type        = string
  default     = "secret/data/fabric"
}
