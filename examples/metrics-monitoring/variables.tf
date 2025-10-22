# ==============================================================================
# PROVIDER CONFIGURATION
# ==============================================================================

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
  sensitive   = true
  default     = "admin123"
}

# ==============================================================================
# PROMETHEUS CONFIGURATION
# ==============================================================================

variable "prometheus_version" {
  description = "Prometheus version to deploy"
  type        = string
  default     = "v2.45.0"
}

variable "prometheus_port" {
  description = "Port for Prometheus server"
  type        = number
  default     = 9090
}

variable "scrape_interval" {
  description = "Default scrape interval in seconds"
  type        = number
  default     = 15
}

variable "deployment_mode" {
  description = "Deployment mode: docker or binary"
  type        = string
  default     = "docker"
}

variable "network_mode" {
  description = "Docker network mode: bridge or host"
  type        = string
  default     = "bridge"
}

# ==============================================================================
# METRICS TARGETS
# ==============================================================================

variable "peer_metrics_targets" {
  description = "List of Fabric peer metrics endpoints"
  type        = list(string)
  default     = []
}

variable "orderer_metrics_targets" {
  description = "List of Fabric orderer metrics endpoints"
  type        = list(string)
  default     = []
}

variable "besu_metrics_targets" {
  description = "List of Besu node metrics endpoints"
  type        = list(string)
  default     = []
}

variable "custom_monitoring_targets" {
  description = "List of custom service endpoints to monitor"
  type        = list(string)
  default     = []
}

variable "custom_metrics_path" {
  description = "Metrics path for custom services"
  type        = string
  default     = "/metrics"
}

variable "custom_scrape_interval" {
  description = "Scrape interval for custom services"
  type        = string
  default     = "30s"
}
