variable "chainlaunch_url" {
  type        = string
  description = "The URL of the Chainlaunch API server"
  default     = "http://localhost:8100"
}

variable "chainlaunch_username" {
  type        = string
  description = "The username for Chainlaunch API authentication"
  default     = "admin"
}

variable "chainlaunch_password" {
  type        = string
  description = "The password for Chainlaunch API authentication"
  sensitive   = true
  default     = "admin123"
}
