terraform {
  required_providers {
    chainlaunch = {
      source = "registry.terraform.io/chainlaunch/chainlaunch"
    }
  }
}

provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# Create the plugin definition from a YAML file
resource "chainlaunch_plugin" "hlf_api" {
  yaml_file_path = var.plugin_yaml_path
}

# Alternative: Create plugin from inline YAML content
# resource "chainlaunch_plugin" "hlf_api" {
#   yaml_content = file("${path.module}/plugin.yaml")
# }
