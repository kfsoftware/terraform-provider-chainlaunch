variable "chainlaunch_url" {
  description = "Chainlaunch instance URL"
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
  sensitive   = true
  default     = "admin123"
}

variable "local_orderer_id" {
  description = "ID of local orderer node (for creating channels with external orgs)"
  type        = string
  default     = "1"
}

variable "local_orderer_tls_cert" {
  description = "TLS CA certificate for local orderer"
  type        = string
  default     = ""
}

variable "local_org_id" {
  description = "ID of local organization"
  type        = string
  default     = "1"
}
