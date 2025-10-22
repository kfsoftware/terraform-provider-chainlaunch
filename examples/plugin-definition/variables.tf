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

variable "plugin_yaml_path" {
  description = "Path to the plugin YAML file"
  type        = string
  default     = "./plugin.yaml"
}
