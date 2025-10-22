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

# Query existing plugin definition (must already be registered)
data "chainlaunch_plugin" "hlf_api" {
  name = var.plugin_name
}

# Query organization
data "chainlaunch_fabric_organization" "org1" {
  msp_id = var.organization_msp_id
}

# Query peers by name
data "chainlaunch_fabric_peer" "peer0" {
  name = var.peer0_name
}

data "chainlaunch_fabric_peer" "peer1" {
  count = var.peer1_name != "" ? 1 : 0
  name  = var.peer1_name
}

# Create admin identity for the API
resource "chainlaunch_fabric_identity" "api_admin" {
  organization_id = data.chainlaunch_fabric_organization.org1.id
  name            = var.identity_name
  role            = "admin"
  description     = "Admin identity for HLF API plugin"
}

# Deploy the plugin with parameters
# Expected format: {"channelName":"mychannel","key":{"keyId":508,"orgId":49},"peers":[125],"port":9501}
resource "chainlaunch_plugin_deployment" "hlf_api" {
  plugin_name = data.chainlaunch_plugin.hlf_api.name

  # Parameters are JSON-encoded based on the plugin's parameter schema
  parameters = jsonencode({
    channelName = var.channel_name
    key = {
      keyId = tonumber(chainlaunch_fabric_identity.api_admin.id)
      orgId = tonumber(data.chainlaunch_fabric_organization.org1.id)
    }
    peers = concat(
      [tonumber(data.chainlaunch_fabric_peer.peer0.id)],
      var.peer1_name != "" ? [tonumber(data.chainlaunch_fabric_peer.peer1[0].id)] : []
    )
    port = var.api_port
  })

  depends_on = [chainlaunch_fabric_identity.api_admin]
}
