# ==============================================================================
# PROVIDER CONFIGURATION
# ==============================================================================

variable "chainlaunch_url" {
  description = "Chainlaunch API URL"
  type        = string
  default     = "http://localhost:8100"
}

variable "chainlaunch_username" {
  description = "Chainlaunch username"
  type        = string
  default     = "admin"
}

variable "chainlaunch_password" {
  description = "Chainlaunch password"
  type        = string
  sensitive   = true
  default     = "admin123"
}

# ==============================================================================
# NETWORK CONFIGURATION
# ==============================================================================

variable "network_name" {
  description = "Name of the Besu network"
  type        = string
  default     = "besu-dev-network"
}

variable "network_description" {
  description = "Description of the Besu network"
  type        = string
  default     = "A complete Hyperledger Besu network"
}

variable "chain_id" {
  description = "Chain ID for the Besu network"
  type        = number
  default     = 1337
}

variable "consensus_mechanism" {
  description = "Consensus mechanism (qbft, ibft2, clique)"
  type        = string
  default     = "qbft"

  validation {
    condition     = contains(["qbft", "ibft2", "clique"], var.consensus_mechanism)
    error_message = "Consensus mechanism must be one of: qbft, ibft2, clique"
  }
}

variable "block_period" {
  description = "Block time in seconds"
  type        = number
  default     = 2
}

variable "epoch_length" {
  description = "Number of blocks after which to reset all votes"
  type        = number
  default     = 30000
}

variable "request_timeout" {
  description = "Timeout for each consensus round in seconds"
  type        = number
  default     = 10
}

variable "gas_limit" {
  description = "Block gas limit (in hexadecimal format)"
  type        = string
  default     = "0x1fffffffffffff"
}

# ==============================================================================
# NETWORK ENDPOINTS
# ==============================================================================

variable "external_ip" {
  description = "Default external IP address for all nodes (can be overridden per node)"
  type        = string
  default     = "localhost"
}

variable "internal_ip" {
  description = "Default internal IP address for all nodes (can be overridden per node)"
  type        = string
  default     = "127.0.0.1"
}

# ==============================================================================
# NODE CONFIGURATION
# ==============================================================================

variable "nodes" {
  description = "Map of Besu nodes to create"
  type = map(object({
    external_ip      = optional(string) # Uses var.external_ip if not specified
    internal_ip      = optional(string) # Uses var.internal_ip if not specified
    p2p_host         = string
    p2p_port         = number
    rpc_host         = string
    rpc_port         = number
    mode             = string
    version          = optional(string)
    min_gas_price    = optional(number)
    metrics_enabled  = optional(bool)
    metrics_port     = optional(number)
    metrics_protocol = optional(string)
    host_allow_list  = optional(string)
    jwt_enabled      = optional(bool)
    jwt_algorithm    = optional(string)
    environment      = optional(map(string))
    is_bootnode      = optional(bool) # Mark this node as a bootnode
  }))

  default = {
    node0 = {
      # external_ip and internal_ip will use defaults (localhost/127.0.0.1)
      p2p_host         = "0.0.0.0"
      p2p_port         = 30303
      rpc_host         = "0.0.0.0"
      rpc_port         = 8545
      mode             = "docker"
      version          = "24.5.1"
      min_gas_price    = 1000
      metrics_enabled  = true
      metrics_port     = 9545
      metrics_protocol = "prometheus"
      host_allow_list  = "*"
      is_bootnode      = true
    }
    node1 = {
      # external_ip and internal_ip will use defaults
      p2p_host         = "0.0.0.0"
      p2p_port         = 30304
      rpc_host         = "0.0.0.0"
      rpc_port         = 8546
      mode             = "docker"
      version          = "24.5.1"
      metrics_enabled  = true
      metrics_port     = 9546
      metrics_protocol = "prometheus"
      host_allow_list  = "*"
    }
    node2 = {
      # external_ip and internal_ip will use defaults
      p2p_host         = "0.0.0.0"
      p2p_port         = 30305
      rpc_host         = "0.0.0.0"
      rpc_port         = 8547
      mode             = "docker"
      version          = "24.5.1"
      metrics_enabled  = true
      metrics_port     = 9547
      metrics_protocol = "prometheus"
      host_allow_list  = "*"
    }
  }
}

# ==============================================================================
# KEY PROVIDER
# ==============================================================================

variable "key_provider_type" {
  description = "Type of key provider to use (database, vault, awskms)"
  type        = string
  default     = "database"

  validation {
    condition     = contains(["database", "vault", "awskms"], var.key_provider_type)
    error_message = "Key provider type must be one of: database, vault, awskms"
  }
}
