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
# TSDB RETENTION (optional)
# ==============================================================================

variable "retention_time" {
  description = "TSDB time-based retention (e.g. \"30d\", \"90d\"). Empty keeps the Prometheus 15d default."
  type        = string
  default     = ""
}

variable "retention_size" {
  description = "TSDB size-based retention (e.g. \"50GB\", \"512MB\"). Empty enforces no size limit."
  type        = string
  default     = ""
}

# ==============================================================================
# REMOTE WRITE (optional) - ship metrics to an external long-term store
# ==============================================================================

variable "remote_write" {
  description = "Prometheus remote_write endpoints (Grafana Cloud, Thanos, Mimir, etc.). Empty renders no remote_write block."
  type = list(object({
    url          = string
    name         = optional(string)
    bearer_token = optional(string)
    basic_auth = optional(object({
      username = optional(string)
      password = optional(string)
    }))
    tls = optional(object({
      ca_file              = optional(string)
      cert_file            = optional(string)
      key_file             = optional(string)
      server_name          = optional(string)
      insecure_skip_verify = optional(bool)
    }))
  }))
  default   = []
  sensitive = true
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
