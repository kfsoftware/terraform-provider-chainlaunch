terraform {
  required_providers {
    chainlaunch = {
      source = "registry.terraform.io/kfsoftware/chainlaunch"
    }
  }
}

# Configure the Chainlaunch Provider
provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# ==============================================================================
# KEY PROVIDER
# ==============================================================================

# Fetch the key provider (database by default, supports secp256k1 for Ethereum)
data "chainlaunch_key_providers" "provider" {
  type_filter = var.key_provider_type
}

# ==============================================================================
# CRYPTOGRAPHIC KEYS (secp256k1 for Ethereum/Besu)
# ==============================================================================

# Create secp256k1 keys for each Besu node
resource "chainlaunch_key" "node_keys" {
  for_each = var.nodes

  name        = "${var.network_name}-${each.key}-key"
  algorithm   = "EC"
  curve       = "secp256k1" # Ethereum curve
  provider_id = tonumber(data.chainlaunch_key_providers.provider.providers[0].id)
  # is_ca defaults to false for node keys
}

# ==============================================================================
# BESU NETWORK
# ==============================================================================

# Create the Besu network using dedicated Besu network resource
resource "chainlaunch_besu_network" "besu" {
  name            = var.network_name
  description     = var.network_description
  chain_id        = var.chain_id
  consensus       = var.consensus_mechanism
  block_period    = var.block_period
  epoch_length    = var.epoch_length
  request_timeout = var.request_timeout
  gas_limit       = var.gas_limit

  initial_validator_key_ids = [
    for key_name, key in chainlaunch_key.node_keys :
    tonumber(key.id)
  ]

  depends_on = [chainlaunch_key.node_keys]
}

# ==============================================================================
# BESU NODES
# ==============================================================================

# Create Besu nodes
resource "chainlaunch_besu_node" "nodes" {
  for_each = var.nodes

  name       = "${var.network_name}-${each.key}"
  network_id = tonumber(chainlaunch_besu_network.besu.id)
  key_id     = tonumber(chainlaunch_key.node_keys[each.key].id)

  # Deployment configuration
  mode    = each.value.mode
  version = lookup(each.value, "version", "24.5.1")

  # Network endpoints - use defaults if not specified per node
  external_ip = coalesce(each.value.external_ip, var.external_ip)
  internal_ip = coalesce(each.value.internal_ip, var.internal_ip)
  p2p_host    = each.value.p2p_host
  p2p_port    = each.value.p2p_port
  rpc_host    = each.value.rpc_host
  rpc_port    = each.value.rpc_port

  # Boot nodes configuration - use the first node as bootnode for others
  boot_nodes = (each.value.is_bootnode != null ? each.value.is_bootnode : false) ? [] : [
    # Reference boot node (node0) enode URL will be constructed by Chainlaunch
    # Empty for now as boot nodes are discovered after creation
  ]

  # Optional: Minimum gas price
  min_gas_price = each.value.min_gas_price

  # Optional: Metrics configuration
  metrics_enabled  = each.value.metrics_enabled != null ? each.value.metrics_enabled : true
  metrics_port     = each.value.metrics_port
  metrics_protocol = each.value.metrics_protocol != null ? each.value.metrics_protocol : "prometheus"

  # Optional: Host allow list
  host_allow_list = each.value.host_allow_list != null ? each.value.host_allow_list : "*"

  # Optional: JWT authentication
  jwt_enabled                  = each.value.jwt_enabled != null ? each.value.jwt_enabled : false
  jwt_authentication_algorithm = each.value.jwt_algorithm

  # Optional: Environment variables
  environment = each.value.environment != null ? each.value.environment : {
    BESU_LOGGING = "INFO"
  }

  depends_on = [
    chainlaunch_besu_network.besu,
    chainlaunch_key.node_keys
  ]
}

# ==============================================================================
# OUTPUTS
# ==============================================================================

output "network_id" {
  description = "The ID of the created Besu network"
  value       = chainlaunch_besu_network.besu.id
}

output "network_name" {
  description = "The name of the Besu network"
  value       = chainlaunch_besu_network.besu.name
}

output "network_status" {
  description = "The status of the Besu network"
  value       = chainlaunch_besu_network.besu.status
}

output "network_config" {
  description = "The Besu network configuration"
  value = {
    chain_id        = chainlaunch_besu_network.besu.chain_id
    consensus       = chainlaunch_besu_network.besu.consensus
    block_period    = chainlaunch_besu_network.besu.block_period
    epoch_length    = chainlaunch_besu_network.besu.epoch_length
    request_timeout = chainlaunch_besu_network.besu.request_timeout
  }
}

output "node_keys" {
  description = "Created cryptographic keys for nodes"
  value = {
    for key_name, key in chainlaunch_key.node_keys :
    key_name => {
      id        = key.id
      name      = key.name
      algorithm = key.algorithm
      curve     = key.curve
    }
  }
}

output "nodes" {
  description = "Created Besu nodes"
  value = {
    for node_key, node in chainlaunch_besu_node.nodes :
    node_key => {
      id           = node.id
      name         = node.name
      status       = node.status
      external_ip  = node.external_ip
      rpc_port     = node.rpc_port
      p2p_port     = node.p2p_port
      rpc_endpoint = "http://${node.external_ip}:${node.rpc_port}"
    }
  }
}

output "bootnode_info" {
  description = "Information about the bootnode"
  value = {
    for node_key, node_config in var.nodes :
    node_key => {
      name         = "${var.network_name}-${node_key}"
      is_bootnode  = node_config.is_bootnode != null ? node_config.is_bootnode : false
      p2p_endpoint = "${coalesce(node_config.external_ip, var.external_ip)}:${node_config.p2p_port}"
    }
    if(node_config.is_bootnode != null ? node_config.is_bootnode : false)
  }
}

output "rpc_endpoints" {
  description = "RPC endpoints for all nodes"
  value = {
    for node_key, node in chainlaunch_besu_node.nodes :
    node_key => "http://${node.external_ip}:${node.rpc_port}"
  }
}

output "metrics_endpoints" {
  description = "Metrics endpoints for all nodes (if enabled)"
  value = {
    for node_key, node in chainlaunch_besu_node.nodes :
    node_key => (var.nodes[node_key].metrics_enabled != null ? var.nodes[node_key].metrics_enabled : true) ?
    "http://${node.external_ip}:${var.nodes[node_key].metrics_port != null ? var.nodes[node_key].metrics_port : 9545}" :
    "metrics disabled"
  }
}
