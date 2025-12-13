# ==============================================================================
# Bidirectional Node Invitations - Multi-Provider Example
# ==============================================================================
# This example demonstrates bidirectional node sharing between two Chainlaunch
# instances using provider aliases in a single Terraform project.
#
# Setup:
# - Node 1 on localhost:8100 (admin/admin123)
# - Node 2 on localhost:8104 (admin/admin)
# ==============================================================================

terraform {
  required_providers {
    chainlaunch = {
      source  = "registry.terraform.io/kfsoftware/chainlaunch"
      version = "0.1.0"
    }
  }
}

# ==============================================================================
# PROVIDER ALIASES - One for each Chainlaunch instance
# ==============================================================================

provider "chainlaunch" {
  alias    = "node1"
  url      = "http://localhost:8100"
  username = "admin"
  password = "admin123"
}

provider "chainlaunch" {
  alias    = "node2"
  url      = "http://localhost:8104"
  username = "admin"
  password = "admin"
}

# ==============================================================================
# STEP 1: Node 1 generates invitation (bidirectional by default)
# ==============================================================================

resource "chainlaunch_node_invitation" "node1_to_node2" {
  provider = chainlaunch.node1

  # bidirectional defaults to true, so both can share nodes
}

# ==============================================================================
# STEP 2: Node 2 accepts Node 1's invitation
# ==============================================================================

resource "chainlaunch_node_accept_invitation" "node2_accepts_node1" {
  provider = chainlaunch.node2

  invitation_jwt = chainlaunch_node_invitation.node1_to_node2.invitation_jwt
}

# ==============================================================================
# STEP 3: Node 2 generates invitation back to Node 1 (optional for bidirectional)
# ==============================================================================

resource "chainlaunch_node_invitation" "node2_to_node1" {
  provider = chainlaunch.node2

  # This is optional if Node 1's invitation was bidirectional
  # But demonstrates the full handshake
}

# ==============================================================================
# STEP 4: Node 1 accepts Node 2's invitation
# ==============================================================================

resource "chainlaunch_node_accept_invitation" "node1_accepts_node2" {
  provider = chainlaunch.node1

  invitation_jwt = chainlaunch_node_invitation.node2_to_node1.invitation_jwt
}

# ==============================================================================
# STEP 5: Node 2 syncs external nodes from Node 1
# ==============================================================================
# This automatically imports all nodes (peers, orderers, Besu nodes) from Node 1
# and stores them as external nodes in Node 2
# ==============================================================================

resource "chainlaunch_external_nodes_sync" "node2_sync_from_node1" {
  provider = chainlaunch.node2

  peer_node_id = "node1" # The node ID from Node 1's instance

  depends_on = [chainlaunch_node_accept_invitation.node2_accepts_node1]
}

# ==============================================================================
# STEP 6: Node 1 syncs external nodes from Node 2
# ==============================================================================

resource "chainlaunch_external_nodes_sync" "node1_sync_from_node2" {
  provider = chainlaunch.node1

  peer_node_id = "node2" # The node ID from Node 2's instance

  depends_on = [chainlaunch_node_accept_invitation.node1_accepts_node2]
}

# ==============================================================================
# ALTERNATIVE: Sync from ALL connected peers automatically
# ==============================================================================
# Instead of manually syncing from each peer (STEP 5 & 6), you can use
# chainlaunch_sync_all_external_nodes to automatically sync from ALL
# connected peers without specifying peer_node_id for each one.
# ==============================================================================

# Uncomment to use automatic sync for all peers:
# resource "chainlaunch_sync_all_external_nodes" "node2_sync_all" {
#   provider = chainlaunch.node2
#
#   # No configuration needed - automatically discovers and syncs from all connected peers
#
#   depends_on = [
#     chainlaunch_node_accept_invitation.node2_accepts_node1,
#     # Add more node acceptances here as needed
#   ]
# }

# resource "chainlaunch_sync_all_external_nodes" "node1_sync_all" {
#   provider = chainlaunch.node1
#
#   depends_on = [
#     chainlaunch_node_accept_invitation.node1_accepts_node2,
#   ]
# }

# Benefits of chainlaunch_sync_all_external_nodes:
# - Automatic discovery of all connected peers
# - No need to specify peer_node_id for each peer
# - Aggregated results across all peers
# - Per-peer sync status and error tracking
# - Continues syncing even if one peer fails

# ==============================================================================
# STEP 7: Query synced external nodes
# ==============================================================================
# After syncing, you can query the external nodes that have been imported
# ==============================================================================

# Query external Fabric organizations synced to Node 2
data "chainlaunch_external_fabric_organizations" "node2_orgs" {
  provider = chainlaunch.node2

  depends_on = [chainlaunch_external_nodes_sync.node2_sync_from_node1]
}

# Query external Fabric peers synced to Node 2
data "chainlaunch_external_fabric_peers" "node2_peers" {
  provider = chainlaunch.node2

  depends_on = [chainlaunch_external_nodes_sync.node2_sync_from_node1]
}

# Query external Fabric orderers synced to Node 2
data "chainlaunch_external_fabric_orderers" "node2_orderers" {
  provider = chainlaunch.node2

  depends_on = [chainlaunch_external_nodes_sync.node2_sync_from_node1]
}

# Query external Besu nodes synced to Node 2
data "chainlaunch_external_besu_nodes" "node2_besu" {
  provider = chainlaunch.node2

  depends_on = [chainlaunch_external_nodes_sync.node2_sync_from_node1]
}

# Similarly for Node 1
data "chainlaunch_external_fabric_organizations" "node1_orgs" {
  provider = chainlaunch.node1

  depends_on = [chainlaunch_external_nodes_sync.node1_sync_from_node2]
}

data "chainlaunch_external_fabric_peers" "node1_peers" {
  provider = chainlaunch.node1

  depends_on = [chainlaunch_external_nodes_sync.node1_sync_from_node2]
}

data "chainlaunch_external_fabric_orderers" "node1_orderers" {
  provider = chainlaunch.node1

  depends_on = [chainlaunch_external_nodes_sync.node1_sync_from_node2]
}

data "chainlaunch_external_besu_nodes" "node1_besu" {
  provider = chainlaunch.node1

  depends_on = [chainlaunch_external_nodes_sync.node1_sync_from_node2]
}

# ==============================================================================
# EXAMPLE: Using External Nodes Data
# ==============================================================================
# This section demonstrates how to use the synced external nodes in your
# Terraform configurations
# ==============================================================================

# Example 1: Get list of all external peer endpoints
locals {
  # Extract peer endpoints from Node 2's external peers
  node2_external_peer_endpoints = [
    for peer in data.chainlaunch_external_fabric_peers.node2_peers.peers :
    peer.external_endpoint
  ]

  # Extract orderer endpoints from Node 2's external orderers
  node2_external_orderer_endpoints = [
    for orderer in data.chainlaunch_external_fabric_orderers.node2_orderers.orderers :
    orderer.external_endpoint
  ]

  # Extract MSP IDs from Node 2's external organizations
  node2_external_msp_ids = [
    for org in data.chainlaunch_external_fabric_organizations.node2_orgs.organizations :
    org.msp_id
  ]

  # Extract Besu enode URLs from Node 2's external Besu nodes
  node2_external_besu_enodes = [
    for node in data.chainlaunch_external_besu_nodes.node2_besu.nodes :
    node.enode_url
  ]
}

# Example 2: Filter external nodes by criteria
locals {
  # Get only peers from a specific organization (e.g., "Org1MSP")
  org1_peers = [
    for peer in data.chainlaunch_external_fabric_peers.node2_peers.peers :
    peer if peer.msp_id == "Org1MSP"
  ]

  # Get peers running a specific Fabric version
  fabric_v2_peers = [
    for peer in data.chainlaunch_external_fabric_peers.node2_peers.peers :
    peer if can(regex("^2\\.", peer.version))
  ]

  # Get Besu nodes with metrics enabled
  besu_nodes_with_metrics = [
    for node in data.chainlaunch_external_besu_nodes.node2_besu.nodes :
    node if node.metrics_enabled
  ]
}

# Example 3: Count external nodes by type
locals {
  node2_stats = {
    total_organizations = length(data.chainlaunch_external_fabric_organizations.node2_orgs.organizations)
    total_peers         = length(data.chainlaunch_external_fabric_peers.node2_peers.peers)
    total_orderers      = length(data.chainlaunch_external_fabric_orderers.node2_orderers.orderers)
    total_besu_nodes    = length(data.chainlaunch_external_besu_nodes.node2_besu.nodes)
  }
}

# ==============================================================================
# PRACTICAL USE CASE: Using External Nodes in Network Configuration
# ==============================================================================
# The following commented examples show how you would use the external nodes
# data in real Terraform resources (commented out to avoid requiring actual
# network setup)
# ==============================================================================

# Example 4: Configure Prometheus monitoring for external peers
# resource "chainlaunch_metrics_job" "external_peers_monitoring" {
#   provider = chainlaunch.node2
#
#   job_name = "external-fabric-peers"
#   targets = [
#     for peer in data.chainlaunch_external_fabric_peers.node2_peers.peers :
#     "${peer.external_endpoint}:9443"  # Assuming metrics on port 9443
#   ]
#   metrics_path    = "/metrics"
#   scrape_interval = "15s"
# }

# Example 5: Create a Besu node using external nodes as bootnodes
# resource "chainlaunch_besu_node" "node2_besu_node" {
#   provider   = chainlaunch.node2
#   name       = "besu-node-2"
#   network_id = var.besu_network_id
#
#   # Use external Besu nodes as bootnodes
#   bootnodes = local.node2_external_besu_enodes
#
#   p2p_host = "0.0.0.0"
#   p2p_port = 30303
# }

# Example 6: Join a Fabric channel with both local and external orderers
# resource "chainlaunch_fabric_join_node" "join_orderer_to_channel" {
#   provider   = chainlaunch.node2
#   network_id = var.fabric_channel_id
#   node_type  = "orderer"
#
#   # Combine local orderer with external orderers from Node 1
#   node_ids = concat(
#     [chainlaunch_fabric_orderer.local_orderer.id],
#     [
#       for orderer in data.chainlaunch_external_fabric_orderers.node2_orderers.orderers :
#       orderer.id
#     ]
#   )
# }

# Example 7: Configure chaincode endorsement policy using external organizations
# resource "chainlaunch_fabric_chaincode_definition" "shared_chaincode" {
#   provider     = chainlaunch.node2
#   chaincode_id = chainlaunch_fabric_chaincode.mycc.id
#   version      = "1.0"
#   sequence     = 1
#   docker_image = "myregistry/mycc:1.0"
#
#   # Endorsement policy requiring signatures from both local and external orgs
#   endorsement_policy = "OR(${join(",", concat(
#     ["'Org2MSP.peer'"],  # Local organization
#     [
#       for org in data.chainlaunch_external_fabric_organizations.node2_orgs.organizations :
#       "'${org.msp_id}.peer'"
#     ]
#   ))})"
# }

# Example 8: Create network configuration map for application
# locals {
#   network_config = {
#     organizations = {
#       for org in data.chainlaunch_external_fabric_organizations.node2_orgs.organizations :
#       org.msp_id => {
#         msp_id           = org.msp_id
#         sign_certificate = org.sign_certificate
#         tls_certificate  = org.tls_certificate
#       }
#     }
#     peers = {
#       for peer in data.chainlaunch_external_fabric_peers.node2_peers.peers :
#       peer.name => {
#         endpoint        = peer.external_endpoint
#         msp_id          = peer.msp_id
#         tls_certificate = peer.tls_certificate
#       }
#     }
#     orderers = {
#       for orderer in data.chainlaunch_external_fabric_orderers.node2_orderers.orderers :
#       orderer.name => {
#         endpoint        = orderer.external_endpoint
#         msp_id          = orderer.msp_id
#         tls_certificate = orderer.tls_certificate
#       }
#     }
#   }
# }

# You can then output this configuration for your application to use
# output "network_config" {
#   description = "Complete network configuration including external nodes"
#   value       = local.network_config
#   sensitive   = false
# }
