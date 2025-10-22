# ==============================================================================
# Chainlaunch Configuration
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
  default     = "admin123"
  sensitive   = true
}

# ==============================================================================
# Mailpit Configuration
# ==============================================================================

variable "mailpit_version" {
  description = "Mailpit Docker image version"
  type        = string
  default     = "latest"
}

variable "mailpit_container_name" {
  description = "Name of the Mailpit Docker container"
  type        = string
  default     = "chainlaunch-mailpit"
}

variable "mailpit_host" {
  description = "Mailpit SMTP host (use 'host.docker.internal' if Chainlaunch is in Docker)"
  type        = string
  default     = "localhost"
}

variable "mailpit_smtp_port" {
  description = "Mailpit SMTP port"
  type        = number
  default     = 1026
}

variable "mailpit_web_port" {
  description = "Mailpit web UI port"
  type        = number
  default     = 8026
}

# ==============================================================================
# Email Configuration
# ==============================================================================

variable "from_email" {
  description = "Sender email address"
  type        = string
  default     = "chainlaunch@example.com"
}

variable "from_name" {
  description = "Sender display name"
  type        = string
  default     = "Chainlaunch Alerts"
}

variable "to_email" {
  description = "Recipient email address"
  type        = string
  default     = "admin@example.com"
}

variable "smtp_username" {
  description = "SMTP username (optional for Mailpit)"
  type        = string
  default     = ""
}

variable "smtp_password" {
  description = "SMTP password (optional for Mailpit)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "use_tls" {
  description = "Use TLS for SMTP connection"
  type        = bool
  default     = false
}

variable "skip_tls_verify" {
  description = "Skip TLS certificate verification"
  type        = bool
  default     = false
}

# ==============================================================================
# Notification Triggers
# ==============================================================================

variable "notify_backup_success" {
  description = "Send notifications on successful backups"
  type        = bool
  default     = false
}
