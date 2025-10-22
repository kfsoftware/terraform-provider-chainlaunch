variable "chainlaunch_url" {
  description = "Chainlaunch API URL"
  type        = string
  default     = "http://localhost:8100"
}

variable "chainlaunch_username" {
  description = "Chainlaunch username"
  type        = string
  default     = "admin"
}

variable "chainlaunch_password" {
  description = "Chainlaunch password"
  type        = string
  default     = "admin123"
  sensitive   = true
}

variable "plugin_name" {
  description = "Name of the plugin to deploy (must already be registered)"
  type        = string
  default     = "hlf-plugin-api"
}

variable "organization_msp_id" {
  description = "MSP ID of the Fabric organization"
  type        = string
  default     = "Org1MSP"
}

variable "peer0_name" {
  description = "Name of the first peer (e.g., 'peer0.org1.example.com')"
  type        = string
}

variable "peer1_name" {
  description = "Name of the second peer (optional, e.g., 'peer1.org1.example.com')"
  type        = string
  default     = ""
}

variable "identity_name" {
  description = "Name for the admin identity that will be created"
  type        = string
  default     = "hlf-api-admin"
}

variable "channel_name" {
  description = "Name of the Fabric channel"
  type        = string
  default     = "mychannel"
}

variable "api_port" {
  description = "Port for the API server"
  type        = number
  default     = 8080
}
