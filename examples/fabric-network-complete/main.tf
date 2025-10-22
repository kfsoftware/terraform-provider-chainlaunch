terraform {
  required_providers {
    chainlaunch = {
      source = "registry.terraform.io/chainlaunch/chainlaunch"
    }
  }
}

# Configure the Chainlaunch Provider
provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# Fetch the default key provider
data "chainlaunch_key_providers" "default" {}

# ==============================================================================
# PEER ORGANIZATIONS AND NODES
# ==============================================================================

# Create peer organizations dynamically
resource "chainlaunch_fabric_organization" "peer_orgs" {
  for_each = var.peer_organizations

  msp_id      = each.value.msp_id
  description = each.value.description
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
}

# Create peer nodes for each organization
resource "chainlaunch_fabric_peer" "peers" {
  for_each = merge([
    for org_key, org in var.peer_organizations : {
      for peer_key, peer in org.peers :
      "${org_key}-${peer_key}" => merge(peer, {
        org_key = org_key
        msp_id  = org.msp_id
      })
    }
  ]...)

  name            = each.value.name
  organization_id = chainlaunch_fabric_organization.peer_orgs[each.value.org_key].id
  msp_id          = each.value.msp_id
  mode            = each.value.mode
  version         = each.value.version

  # Network endpoints
  external_endpoint         = each.value.external_endpoint
  listen_address            = each.value.listen_address
  chaincode_address         = each.value.chaincode_address
  events_address            = each.value.events_address
  operations_listen_address = each.value.operations_listen_address

  # Optional: Address overrides (for mapping external to internal addresses)
  address_overrides = lookup(each.value, "address_overrides", [])

  # Optional: Domain names
  domain_names = lookup(each.value, "domain_names", [each.value.name, "localhost"])

  # Certificate configuration
  certificate_expiration = lookup(each.value, "certificate_expiration", 365)
  auto_renewal_enabled   = lookup(each.value, "auto_renewal_enabled", true)
  auto_renewal_days      = lookup(each.value, "auto_renewal_days", 30)

  # Environment variables
  environment = lookup(each.value, "environment", {
    CORE_PEER_GOSSIP_USELEADERELECTION = "true"
    CORE_PEER_GOSSIP_ORGLEADER         = "false"
    CORE_PEER_PROFILE_ENABLED          = "true"
    FABRIC_LOGGING_SPEC                = "INFO"
  })
}

# ==============================================================================
# ORDERER ORGANIZATIONS AND NODES
# ==============================================================================

# Create orderer organizations dynamically
resource "chainlaunch_fabric_organization" "orderer_orgs" {
  for_each = var.orderer_organizations

  msp_id      = each.value.msp_id
  description = each.value.description
  provider_id = data.chainlaunch_key_providers.default.default_provider_id
}

# Create orderer nodes (consenters) for each organization
resource "chainlaunch_fabric_orderer" "orderers" {
  for_each = merge([
    for org_key, org in var.orderer_organizations : {
      for orderer_key, orderer in org.orderers :
      "${org_key}-${orderer_key}" => merge(orderer, {
        org_key = org_key
        msp_id  = org.msp_id
      })
    }
  ]...)

  name            = each.value.name
  organization_id = chainlaunch_fabric_organization.orderer_orgs[each.value.org_key].id
  msp_id          = each.value.msp_id
  mode            = each.value.mode
  version         = each.value.version

  # Network endpoints
  external_endpoint         = each.value.external_endpoint
  listen_address            = each.value.listen_address
  admin_address             = each.value.admin_address
  operations_listen_address = each.value.operations_listen_address

  # Optional: Domain names
  domain_names = lookup(each.value, "domain_names", [each.value.name, "localhost"])

  # Certificate configuration
  certificate_expiration = lookup(each.value, "certificate_expiration", 365)
  auto_renewal_enabled   = lookup(each.value, "auto_renewal_enabled", true)
  auto_renewal_days      = lookup(each.value, "auto_renewal_days", 30)

  # Environment variables
  environment = lookup(each.value, "environment", {
    ORDERER_GENERAL_LOGLEVEL         = "INFO"
    ORDERER_OPERATIONS_LISTENADDRESS = each.value.operations_listen_address
    FABRIC_LOGGING_SPEC              = "INFO"
  })
}

# ==============================================================================
# FABRIC NETWORK (CHANNEL)
# ==============================================================================

resource "chainlaunch_fabric_network" "channel" {
  name        = var.channel_name
  description = var.channel_description

  # Peer organizations - Group peers by their organization
  peer_organizations = [
    for org_key, org in var.peer_organizations : {
      id = chainlaunch_fabric_organization.peer_orgs[org_key].id
      node_ids = [
        for peer_key, peer in org.peers :
        chainlaunch_fabric_peer.peers["${org_key}-${peer_key}"].id
      ]
    }
  ]

  # Orderer organizations - Group orderers by their organization
  orderer_organizations = [
    for org_key, org in var.orderer_organizations : {
      id = chainlaunch_fabric_organization.orderer_orgs[org_key].id
      node_ids = [
        for orderer_key, orderer in org.orderers :
        chainlaunch_fabric_orderer.orderers["${org_key}-${orderer_key}"].id
      ]
    }
  ]

  # Consensus configuration
  consensus_type = var.consensus_type

  etcdraft_options = var.etcdraft_options

  # Capabilities
  channel_capabilities     = var.channel_capabilities
  application_capabilities = var.application_capabilities
  orderer_capabilities     = var.orderer_capabilities

  # Batch configuration
  batch_size = var.batch_size

  batch_timeout = var.batch_timeout

  # Optional: Custom policies
  configure_policies = var.configure_policies

  depends_on = [
    chainlaunch_fabric_peer.peers,
    chainlaunch_fabric_orderer.orderers
  ]
}

# ==============================================================================
# JOIN PEERS TO CHANNEL
# ==============================================================================

resource "chainlaunch_fabric_join_node" "peer_joins" {
  for_each = merge([
    for org_key, org in var.peer_organizations : {
      for peer_key, peer in org.peers :
      "${org_key}-${peer_key}" => {
        peer_id = chainlaunch_fabric_peer.peers["${org_key}-${peer_key}"].id
      }
    }
  ]...)

  network_id = chainlaunch_fabric_network.channel.id
  node_id    = each.value.peer_id
  role       = "peer"

  depends_on = [chainlaunch_fabric_network.channel]
}

# ==============================================================================
# JOIN ORDERERS TO CHANNEL
# ==============================================================================

resource "chainlaunch_fabric_join_node" "orderer_joins" {
  for_each = merge([
    for org_key, org in var.orderer_organizations : {
      for orderer_key, orderer in org.orderers :
      "${org_key}-${orderer_key}" => {
        orderer_id = chainlaunch_fabric_orderer.orderers["${org_key}-${orderer_key}"].id
      }
    }
  ]...)

  network_id = chainlaunch_fabric_network.channel.id
  node_id    = each.value.orderer_id
  role       = "orderer"

  depends_on = [chainlaunch_fabric_network.channel]
}

# ==============================================================================
# SET ANCHOR PEERS
# ==============================================================================

# Set anchor peers for each organization (first peer of each org by default)
resource "chainlaunch_fabric_anchor_peers" "anchors" {
  for_each = var.peer_organizations

  network_id      = chainlaunch_fabric_network.channel.id
  organization_id = chainlaunch_fabric_organization.peer_orgs[each.key].id

  # Use the anchor_peer_indices variable or default to first peer [0]
  anchor_peer_ids = [
    for idx in lookup(each.value, "anchor_peer_indices", [0]) :
    chainlaunch_fabric_peer.peers["${each.key}-peer${idx}"].id
  ]

  depends_on = [chainlaunch_fabric_join_node.peer_joins]
}

# ==============================================================================
# CHAINCODE DEPLOYMENT (OPTIONAL)
# ==============================================================================

# Create chaincode resources (only if deploy_chaincode is enabled)
resource "chainlaunch_fabric_chaincode" "chaincodes" {
  for_each = var.deploy_chaincode ? var.chaincodes : {}

  name       = each.value.name
  network_id = chainlaunch_fabric_network.channel.id

  depends_on = [
    chainlaunch_fabric_join_node.peer_joins,
    chainlaunch_fabric_anchor_peers.anchors
  ]
}

# Create chaincode definitions
resource "chainlaunch_fabric_chaincode_definition" "definitions" {
  for_each = var.deploy_chaincode ? var.chaincodes : {}

  chaincode_id       = chainlaunch_fabric_chaincode.chaincodes[each.key].id
  version            = each.value.version
  sequence           = each.value.sequence
  docker_image       = each.value.docker_image
  chaincode_address  = each.value.chaincode_address
  endorsement_policy = each.value.endorsement_policy != "" ? each.value.endorsement_policy : null

  depends_on = [chainlaunch_fabric_chaincode.chaincodes]
}

# Install chaincode on specified peers
resource "chainlaunch_fabric_chaincode_install" "installs" {
  for_each = var.deploy_chaincode ? {
    for cc_key, cc in var.chaincodes :
    cc_key => cc
  } : {}

  definition_id = chainlaunch_fabric_chaincode_definition.definitions[each.key].id

  # Get all peers from the specified organizations
  peer_ids = flatten([
    for org_key in each.value.install_on_orgs : [
      for peer_key, peer in var.peer_organizations[org_key].peers :
      chainlaunch_fabric_peer.peers["${org_key}-${peer_key}"].id
    ]
  ])

  depends_on = [chainlaunch_fabric_chaincode_definition.definitions]
}

# Approve chaincode for each organization
resource "chainlaunch_fabric_chaincode_approve" "approvals" {
  for_each = var.deploy_chaincode ? merge([
    for cc_key, cc in var.chaincodes : {
      for org_key in cc.approve_with_orgs :
      "${cc_key}-${org_key}" => {
        cc_key        = cc_key
        org_key       = org_key
        definition_id = chainlaunch_fabric_chaincode_definition.definitions[cc_key].id
      }
    }
  ]...) : {}

  definition_id = each.value.definition_id
  # Use first peer of the organization for approval
  peer_id = chainlaunch_fabric_peer.peers["${each.value.org_key}-peer0"].id

  depends_on = [chainlaunch_fabric_chaincode_install.installs]
}

# Commit chaincode to the channel
resource "chainlaunch_fabric_chaincode_commit" "commits" {
  for_each = var.deploy_chaincode ? var.chaincodes : {}

  definition_id = chainlaunch_fabric_chaincode_definition.definitions[each.key].id
  # Use first peer of the commit organization
  peer_id = chainlaunch_fabric_peer.peers["${each.value.commit_with_org}-peer0"].id

  depends_on = [chainlaunch_fabric_chaincode_approve.approvals]
}

# Deploy chaincode (start containers)
resource "chainlaunch_fabric_chaincode_deploy" "deploys" {
  for_each = var.deploy_chaincode ? var.chaincodes : {}

  definition_id         = chainlaunch_fabric_chaincode_definition.definitions[each.key].id
  environment_variables = each.value.environment_variables

  depends_on = [chainlaunch_fabric_chaincode_commit.commits]
}

# ==============================================================================
# OUTPUTS
# ==============================================================================

output "channel_id" {
  description = "The ID of the created Fabric channel"
  value       = chainlaunch_fabric_network.channel.id
}

output "channel_name" {
  description = "The name of the Fabric channel"
  value       = chainlaunch_fabric_network.channel.name
}

output "channel_status" {
  description = "The status of the Fabric channel"
  value       = chainlaunch_fabric_network.channel.status
}

output "peer_organizations" {
  description = "Created peer organizations"
  value = {
    for org_key, org in chainlaunch_fabric_organization.peer_orgs :
    org_key => {
      id          = org.id
      msp_id      = org.msp_id
      description = org.description
    }
  }
}

output "orderer_organizations" {
  description = "Created orderer organizations"
  value = {
    for org_key, org in chainlaunch_fabric_organization.orderer_orgs :
    org_key => {
      id          = org.id
      msp_id      = org.msp_id
      description = org.description
    }
  }
}

output "peers" {
  description = "Created peer nodes"
  value = {
    for key, peer in chainlaunch_fabric_peer.peers :
    key => {
      id                = peer.id
      name              = peer.name
      status            = peer.status
      external_endpoint = peer.external_endpoint
      msp_id            = peer.msp_id
    }
  }
}

output "orderers" {
  description = "Created orderer nodes (consenters)"
  value = {
    for key, orderer in chainlaunch_fabric_orderer.orderers :
    key => {
      id                = orderer.id
      name              = orderer.name
      status            = orderer.status
      external_endpoint = orderer.external_endpoint
      msp_id            = orderer.msp_id
    }
  }
}

output "anchor_peers" {
  description = "Anchor peers configuration"
  value = {
    for org_key, anchor in chainlaunch_fabric_anchor_peers.anchors :
    org_key => {
      organization_id = anchor.organization_id
      anchor_peer_ids = anchor.anchor_peer_ids
    }
  }
}

output "chaincodes" {
  description = "Deployed chaincodes"
  value = var.deploy_chaincode ? {
    for cc_key, cc in chainlaunch_fabric_chaincode.chaincodes :
    cc_key => {
      id         = cc.id
      name       = cc.name
      network_id = cc.network_id
      definition = {
        id                = chainlaunch_fabric_chaincode_definition.definitions[cc_key].id
        version           = chainlaunch_fabric_chaincode_definition.definitions[cc_key].version
        sequence          = chainlaunch_fabric_chaincode_definition.definitions[cc_key].sequence
        docker_image      = chainlaunch_fabric_chaincode_definition.definitions[cc_key].docker_image
        chaincode_address = chainlaunch_fabric_chaincode_definition.definitions[cc_key].chaincode_address
      }
    }
  } : {}
}
