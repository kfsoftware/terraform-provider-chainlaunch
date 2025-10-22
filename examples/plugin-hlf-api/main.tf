terraform {
  required_providers {
    chainlaunch = {
      source = "registry.terraform.io/kfsoftware/chainlaunch"
    }
  }
}

provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}


# Data source to query an existing Fabric organization
data "chainlaunch_fabric_organization" "org1" {
  msp_id = var.organization_msp_id
}

# Create an admin identity for the API to use
# This generates a key pair and certificate for accessing the Fabric network
resource "chainlaunch_fabric_identity" "api_admin" {
  organization_id = data.chainlaunch_fabric_organization.org1.id
  name            = var.identity_name
  role            = "admin"
  description     = "Admin identity for HLF API plugin"
}

# Data source to query existing Fabric peers by name
data "chainlaunch_fabric_peer" "peer0" {
  name = var.peer0_name
}

data "chainlaunch_fabric_peer" "peer1" {
  count = var.peer1_name != "" ? 1 : 0
  name  = var.peer1_name
}

# Create the plugin from YAML file
resource "chainlaunch_plugin" "hlf_api" {
  yaml_file_path = "${path.module}/plugin.yaml"
}

# Deploy the plugin with parameters
# Parameters match the plugin's expected schema:
# {"channelName":"mychannel","key":{"keyId":508,"orgId":49},"peers":[125],"port":9501}
resource "chainlaunch_plugin_deployment" "hlf_api" {
  plugin_name = chainlaunch_plugin.hlf_api.name

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

  depends_on = [
    chainlaunch_plugin.hlf_api,
    chainlaunch_fabric_identity.api_admin,
    data.chainlaunch_fabric_organization.org1,
    data.chainlaunch_fabric_peer.peer0
  ]
}
