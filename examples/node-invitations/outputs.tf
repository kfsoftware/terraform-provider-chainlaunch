# ==============================================================================
# Node 1 Outputs
# ==============================================================================

output "node1_invitation_id" {
  description = "Node 1's invitation to Node 2"
  value       = chainlaunch_node_invitation.node1_to_node2.id
}

output "node1_invitation_jwt" {
  description = "Node 1's invitation JWT"
  value       = chainlaunch_node_invitation.node1_to_node2.invitation_jwt
  sensitive   = true
}

output "node1_acceptance_status" {
  description = "Node 1's acceptance of Node 2's invitation"
  value = {
    success = chainlaunch_node_accept_invitation.node1_accepts_node2.success
    error   = chainlaunch_node_accept_invitation.node1_accepts_node2.error
  }
}

# ==============================================================================
# Node 2 Outputs
# ==============================================================================

output "node2_invitation_id" {
  description = "Node 2's invitation to Node 1"
  value       = chainlaunch_node_invitation.node2_to_node1.id
}

output "node2_invitation_jwt" {
  description = "Node 2's invitation JWT"
  value       = chainlaunch_node_invitation.node2_to_node1.invitation_jwt
  sensitive   = true
}

output "node2_acceptance_status" {
  description = "Node 2's acceptance of Node 1's invitation"
  value = {
    success = chainlaunch_node_accept_invitation.node2_accepts_node1.success
    error   = chainlaunch_node_accept_invitation.node2_accepts_node1.error
  }
}

# ==============================================================================
# Summary
# ==============================================================================

output "bidirectional_connection_established" {
  description = "Whether bidirectional connection is fully established"
  value = (
    chainlaunch_node_accept_invitation.node1_accepts_node2.success &&
    chainlaunch_node_accept_invitation.node2_accepts_node1.success
  )
}

output "connection_summary" {
  description = "Summary of the bidirectional connection"
  value       = <<-EOT

    ╔══════════════════════════════════════════════════════════════╗
    ║        Bidirectional Node Connection Summary                ║
    ╚══════════════════════════════════════════════════════════════╝

    Node 1 → Node 2:
      Invitation:  ${chainlaunch_node_invitation.node1_to_node2.id}
      Accepted:    ${chainlaunch_node_accept_invitation.node2_accepts_node1.success ? "✅ YES" : "❌ NO"}

    Node 2 → Node 1:
      Invitation:  ${chainlaunch_node_invitation.node2_to_node1.id}
      Accepted:    ${chainlaunch_node_accept_invitation.node1_accepts_node2.success ? "✅ YES" : "❌ NO"}

    Connection Status: ${chainlaunch_node_accept_invitation.node1_accepts_node2.success && chainlaunch_node_accept_invitation.node2_accepts_node1.success ? "✅ FULLY ESTABLISHED" : "⚠️  INCOMPLETE"}

  EOT
}

# ==============================================================================
# External Nodes Sync Outputs
# ==============================================================================

output "node1_sync_summary" {
  description = "Summary of nodes synced from Node 2 to Node 1"
  value = {
    organizations_added     = chainlaunch_external_nodes_sync.node1_sync_from_node2.organizations_added
    fabric_peers_added      = chainlaunch_external_nodes_sync.node1_sync_from_node2.fabric_peers_added
    fabric_peers_deleted    = chainlaunch_external_nodes_sync.node1_sync_from_node2.fabric_peers_deleted
    fabric_orderers_added   = chainlaunch_external_nodes_sync.node1_sync_from_node2.fabric_orderers_added
    fabric_orderers_deleted = chainlaunch_external_nodes_sync.node1_sync_from_node2.fabric_orderers_deleted
    besu_nodes_added        = chainlaunch_external_nodes_sync.node1_sync_from_node2.besu_nodes_added
    besu_nodes_deleted      = chainlaunch_external_nodes_sync.node1_sync_from_node2.besu_nodes_deleted
    last_sync_at            = chainlaunch_external_nodes_sync.node1_sync_from_node2.last_sync_at
  }
}

output "node2_sync_summary" {
  description = "Summary of nodes synced from Node 1 to Node 2"
  value = {
    organizations_added     = chainlaunch_external_nodes_sync.node2_sync_from_node1.organizations_added
    fabric_peers_added      = chainlaunch_external_nodes_sync.node2_sync_from_node1.fabric_peers_added
    fabric_peers_deleted    = chainlaunch_external_nodes_sync.node2_sync_from_node1.fabric_peers_deleted
    fabric_orderers_added   = chainlaunch_external_nodes_sync.node2_sync_from_node1.fabric_orderers_added
    fabric_orderers_deleted = chainlaunch_external_nodes_sync.node2_sync_from_node1.fabric_orderers_deleted
    besu_nodes_added        = chainlaunch_external_nodes_sync.node2_sync_from_node1.besu_nodes_added
    besu_nodes_deleted      = chainlaunch_external_nodes_sync.node2_sync_from_node1.besu_nodes_deleted
    last_sync_at            = chainlaunch_external_nodes_sync.node2_sync_from_node1.last_sync_at
  }
}

# ==============================================================================
# External Nodes Data Outputs
# ==============================================================================

output "node1_external_organizations" {
  description = "External organizations synced to Node 1"
  value       = data.chainlaunch_external_fabric_organizations.node1_orgs.organizations
}

output "node1_external_peers" {
  description = "External peers synced to Node 1"
  value       = data.chainlaunch_external_fabric_peers.node1_peers.peers
}

output "node1_external_orderers" {
  description = "External orderers synced to Node 1"
  value       = data.chainlaunch_external_fabric_orderers.node1_orderers.orderers
}

output "node1_external_besu_nodes" {
  description = "External Besu nodes synced to Node 1"
  value       = data.chainlaunch_external_besu_nodes.node1_besu.nodes
}

output "node2_external_organizations" {
  description = "External organizations synced to Node 2"
  value       = data.chainlaunch_external_fabric_organizations.node2_orgs.organizations
}

output "node2_external_peers" {
  description = "External peers synced to Node 2"
  value       = data.chainlaunch_external_fabric_peers.node2_peers.peers
}

output "node2_external_orderers" {
  description = "External orderers synced to Node 2"
  value       = data.chainlaunch_external_fabric_orderers.node2_orderers.orderers
}

output "node2_external_besu_nodes" {
  description = "External Besu nodes synced to Node 2"
  value       = data.chainlaunch_external_besu_nodes.node2_besu.nodes
}

# ==============================================================================
# Example Usage Outputs
# ==============================================================================

output "node2_external_peer_endpoints" {
  description = "List of all external peer endpoints synced to Node 2"
  value       = local.node2_external_peer_endpoints
}

output "node2_external_orderer_endpoints" {
  description = "List of all external orderer endpoints synced to Node 2"
  value       = local.node2_external_orderer_endpoints
}

output "node2_external_msp_ids" {
  description = "List of all external MSP IDs synced to Node 2"
  value       = local.node2_external_msp_ids
}

output "node2_external_besu_enodes" {
  description = "List of all external Besu enode URLs synced to Node 2"
  value       = local.node2_external_besu_enodes
}

output "node2_stats" {
  description = "Statistics of external nodes synced to Node 2"
  value       = local.node2_stats
}

output "org1_peers" {
  description = "Peers from Org1MSP (filtered example)"
  value       = local.org1_peers
}

output "fabric_v2_peers" {
  description = "Peers running Fabric v2.x (filtered example)"
  value       = local.fabric_v2_peers
}

output "besu_nodes_with_metrics" {
  description = "Besu nodes with metrics enabled (filtered example)"
  value       = local.besu_nodes_with_metrics
}
