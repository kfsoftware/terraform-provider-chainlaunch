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

# Example: Share a Fabric network with connected peer nodes
# This demonstrates sharing network configuration (genesis block, etc.) with other Chainlaunch instances

# Step 1: Create a Fabric network locally
resource "chainlaunch_fabric_network" "mychannel" {
  name        = "mychannel"
  description = "Shared channel for multi-organization collaboration"

  orderer_endpoints = [
    {
      node_id = tonumber(var.orderer_node_id)
      tls_ca_cert = var.orderer_tls_ca_cert
    }
  ]

  organizations = [
    {
      organization_id = tonumber(var.org1_id)
    }
  ]
}

# Step 2: Share the network with connected peer nodes
resource "chainlaunch_network_share" "share_with_org2" {
  network_id   = chainlaunch_fabric_network.mychannel.id
  network_type = "fabric"

  # List of peer node IDs (connection IDs) to share with
  # These must be peers you've connected to via node invitations
  recipients = [
    var.peer_node_id_org2,  # Connected peer from Org2's Chainlaunch instance
  ]

  metadata = {
    purpose     = "multi-org-channel"
    shared_date = timestamp()
    note        = "Shared for collaboration on supply chain project"
  }
}

# Example: Share a Besu network
resource "chainlaunch_besu_network" "besu_network" {
  name            = "besu-shared-network"
  description     = "Shared Besu network"
  chain_id        = 1337
  consensus       = "qbft"
  block_period    = 5
  epoch_length    = 30000
  request_timeout = 10

  initial_validator_key_ids = [tonumber(var.validator_key_id)]
}

resource "chainlaunch_network_share" "share_besu" {
  network_id   = chainlaunch_besu_network.besu_network.id
  network_type = "besu"

  recipients = [
    var.peer_node_id_partner,
  ]

  metadata = {
    purpose = "besu-consortium"
  }
}
