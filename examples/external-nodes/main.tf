terraform {
  required_providers {
    chainlaunch = {
      source = "kfsoftware/chainlaunch"
    }
  }
}

provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# Step 1: Sync all external nodes from connected peers
resource "chainlaunch_sync_all_external_nodes" "sync" {}

# Step 2: Query ALL external nodes with a single data source
# This is more efficient than using individual data sources for each node type
data "chainlaunch_external_nodes" "all" {
  depends_on = [chainlaunch_sync_all_external_nodes.sync]
}

# Output all Fabric peers
output "fabric_peers" {
  description = "All external Fabric peer nodes"
  value = {
    count = length(data.chainlaunch_external_nodes.all.fabric_peers)
    peers = [
      for peer in data.chainlaunch_external_nodes.all.fabric_peers : {
        name             = peer.name
        msp_id           = peer.msp_id
        endpoint         = peer.external_endpoint
        version          = peer.version
        external_node_id = peer.external_node_id
      }
    ]
  }
}

# Output all Fabric orderers
output "fabric_orderers" {
  description = "All external Fabric orderer nodes"
  value = {
    count    = length(data.chainlaunch_external_nodes.all.fabric_orderers)
    orderers = [
      for orderer in data.chainlaunch_external_nodes.all.fabric_orderers : {
        name             = orderer.name
        msp_id           = orderer.msp_id
        endpoint         = orderer.external_endpoint
        version          = orderer.version
        external_node_id = orderer.external_node_id
      }
    ]
  }
}

# Output all Besu nodes
output "besu_nodes" {
  description = "All external Besu nodes"
  value = {
    count = length(data.chainlaunch_external_nodes.all.besu_nodes)
    nodes = [
      for node in data.chainlaunch_external_nodes.all.besu_nodes : {
        name             = node.name
        enode_url        = node.enode_url
        p2p_endpoint     = "${node.p2p_host}:${node.p2p_port}"
        version          = node.version
        metrics_enabled  = node.metrics_enabled
        external_node_id = node.external_node_id
      }
    ]
  }
}

# Example: Use external peers in a Fabric network configuration
resource "chainlaunch_fabric_network" "shared_channel" {
  name        = "multi-org-channel"
  description = "Channel with external organizations"

  # Use local orderer
  orderer_endpoints = [
    {
      node_id     = tonumber(var.local_orderer_id)
      tls_ca_cert = var.local_orderer_tls_cert
    }
  ]

  # Include local organization
  organizations = [
    {
      organization_id = tonumber(var.local_org_id)
    }
  ]

  depends_on = [data.chainlaunch_external_nodes.all]
}

# Example: Filter peers by MSP ID
output "org1_peers" {
  description = "All Org1MSP peers"
  value = [
    for peer in data.chainlaunch_external_nodes.all.fabric_peers :
    peer if peer.msp_id == "Org1MSP"
  ]
}

# Example: Get all peer endpoints for a specific organization
output "org2_peer_endpoints" {
  description = "Endpoints for Org2MSP peers"
  value = [
    for peer in data.chainlaunch_external_nodes.all.fabric_peers :
    peer.external_endpoint if peer.msp_id == "Org2MSP"
  ]
}

# Example: Count nodes by type
output "node_counts" {
  description = "Count of external nodes by type"
  value = {
    fabric_peers    = length(data.chainlaunch_external_nodes.all.fabric_peers)
    fabric_orderers = length(data.chainlaunch_external_nodes.all.fabric_orderers)
    besu_nodes      = length(data.chainlaunch_external_nodes.all.besu_nodes)
    total           = length(data.chainlaunch_external_nodes.all.fabric_peers) + length(data.chainlaunch_external_nodes.all.fabric_orderers) + length(data.chainlaunch_external_nodes.all.besu_nodes)
  }
}

# Example: Get all unique MSP IDs
output "all_msp_ids" {
  description = "All unique MSP IDs from external nodes"
  value = distinct(concat(
    [for peer in data.chainlaunch_external_nodes.all.fabric_peers : peer.msp_id],
    [for orderer in data.chainlaunch_external_nodes.all.fabric_orderers : orderer.msp_id]
  ))
}

# Example: Export certificates for external peers (for configuration files)
output "peer_certificates" {
  description = "TLS certificates for all external peers"
  sensitive   = true
  value = {
    for peer in data.chainlaunch_external_nodes.all.fabric_peers :
    peer.name => {
      tls_cert  = peer.tls_certificate
      sign_cert = peer.sign_certificate
    }
  }
}
