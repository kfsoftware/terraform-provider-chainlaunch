# Configure the Chainlaunch provider
terraform {
  required_providers {
    chainlaunch = {
      source = "registry.terraform.io/chainlaunch/chainlaunch"
    }
  }
}

# Provider configuration with API key
provider "chainlaunch" {
  url     = "http://localhost:8100"
  api_key = "your-api-key-here"
}

# Alternative: Provider configuration with username/password
# provider "chainlaunch" {
#   url      = "http://localhost:8100"
#   username = "admin"
#   password = "admin123"
# }
