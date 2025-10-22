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
# CHANNEL CONFIGURATION
# ==============================================================================

variable "channel_name" {
  description = "Name of the Fabric channel to create"
  type        = string
  default     = "mychannel"
}

variable "channel_description" {
  description = "Description of the Fabric channel"
  type        = string
  default     = "A complete Fabric network with configurable organizations"
}

# ==============================================================================
# PEER ORGANIZATIONS
# ==============================================================================

variable "peer_organizations" {
  description = "Map of peer organizations and their peers"
  type = map(object({
    msp_id      = string
    description = string
    peers = map(object({
      name                      = string
      mode                      = string
      version                   = string
      external_endpoint         = string
      listen_address            = string
      chaincode_address         = string
      events_address            = string
      operations_listen_address = string
      domain_names              = optional(list(string))
      certificate_expiration    = optional(number)
      auto_renewal_enabled      = optional(bool)
      auto_renewal_days         = optional(number)
      environment               = optional(map(string))
    }))
    anchor_peer_indices = optional(list(number))
  }))

  default = {
    org1 = {
      msp_id      = "Org1MSP"
      description = "First peer organization"
      peers = {
        peer0 = {
          name                      = "peer0-org1"
          mode                      = "service"
          version                   = "3.1.2"
          external_endpoint         = "localhost:7051"
          listen_address            = "0.0.0.0:7051"
          chaincode_address         = "0.0.0.0:7052"
          events_address            = "0.0.0.0:7053"
          operations_listen_address = "0.0.0.0:9443"
        }
        peer1 = {
          name                      = "peer1-org1"
          mode                      = "service"
          version                   = "3.1.2"
          external_endpoint         = "localhost:7151"
          listen_address            = "0.0.0.0:7151"
          chaincode_address         = "0.0.0.0:7152"
          events_address            = "0.0.0.0:7153"
          operations_listen_address = "0.0.0.0:9543"
        }
      }
      anchor_peer_indices = [0] # peer0 is anchor
    }
    org2 = {
      msp_id      = "Org2MSP"
      description = "Second peer organization"
      peers = {
        peer0 = {
          name                      = "peer0-org2"
          mode                      = "service"
          version                   = "3.1.2"
          external_endpoint         = "localhost:8051"
          listen_address            = "0.0.0.0:8051"
          chaincode_address         = "0.0.0.0:8052"
          events_address            = "0.0.0.0:8053"
          operations_listen_address = "0.0.0.0:10443"
        }
      }
      anchor_peer_indices = [0] # peer0 is anchor
    }
  }
}

# ==============================================================================
# ORDERER ORGANIZATIONS
# ==============================================================================

variable "orderer_organizations" {
  description = "Map of orderer organizations and their orderer nodes (consenters)"
  type = map(object({
    msp_id      = string
    description = string
    orderers = map(object({
      name                      = string
      mode                      = string
      version                   = string
      external_endpoint         = string
      listen_address            = string
      admin_address             = string
      operations_listen_address = string
      domain_names              = optional(list(string))
      certificate_expiration    = optional(number)
      auto_renewal_enabled      = optional(bool)
      auto_renewal_days         = optional(number)
      environment               = optional(map(string))
    }))
  }))

  default = {
    orderer_org = {
      msp_id      = "OrdererOrgMSP"
      description = "Orderer organization with 3 consenters for Raft consensus"
      orderers = {
        orderer0 = {
          name                      = "orderer0-ordererorg"
          mode                      = "service"
          version                   = "3.1.1"
          external_endpoint         = "localhost:17050"
          listen_address            = "0.0.0.0:17050"
          admin_address             = "0.0.0.0:17053"
          operations_listen_address = "0.0.0.0:17443"
        }
        orderer1 = {
          name                      = "orderer1-ordererorg"
          mode                      = "service"
          version                   = "3.1.1"
          external_endpoint         = "localhost:17150"
          listen_address            = "0.0.0.0:17150"
          admin_address             = "0.0.0.0:17153"
          operations_listen_address = "0.0.0.0:17543"
        }
        orderer2 = {
          name                      = "orderer2-ordererorg"
          mode                      = "service"
          version                   = "3.1.1"
          external_endpoint         = "localhost:17250"
          listen_address            = "0.0.0.0:17250"
          admin_address             = "0.0.0.0:17253"
          operations_listen_address = "0.0.0.0:17643"
        }
      }
    }
  }
}

# ==============================================================================
# CONSENSUS CONFIGURATION
# ==============================================================================

variable "consensus_type" {
  description = "Consensus type (etcdraft is recommended for production)"
  type        = string
  default     = "etcdraft"

  validation {
    condition     = contains(["etcdraft", "solo"], var.consensus_type)
    error_message = "Consensus type must be either 'etcdraft' or 'solo'."
  }
}

variable "etcdraft_options" {
  description = "Etcd Raft consensus configuration options"
  type = object({
    tick_interval          = string
    election_tick          = number
    heartbeat_tick         = number
    max_inflight_blocks    = number
    snapshot_interval_size = number
  })

  default = {
    tick_interval          = "500ms"
    election_tick          = 10
    heartbeat_tick         = 1
    max_inflight_blocks    = 5
    snapshot_interval_size = 20971520 # 20MB
  }
}

# ==============================================================================
# FABRIC CAPABILITIES
# ==============================================================================

variable "channel_capabilities" {
  description = "Channel capability versions"
  type        = list(string)
  default     = ["V2_0"]
}

variable "application_capabilities" {
  description = "Application capability versions"
  type        = list(string)
  default     = ["V2_0"]
}

variable "orderer_capabilities" {
  description = "Orderer capability versions"
  type        = list(string)
  default     = ["V2_0"]
}

# ==============================================================================
# BATCH CONFIGURATION
# ==============================================================================

variable "batch_size" {
  description = "Batch size configuration for transaction batching"
  type = object({
    max_message_count   = number
    absolute_max_bytes  = number
    preferred_max_bytes = number
  })

  default = {
    max_message_count   = 500
    absolute_max_bytes  = 103809024 # 99MB
    preferred_max_bytes = 524288    # 512KB
  }
}

variable "batch_timeout" {
  description = "Batch timeout for transaction batching"
  type        = string
  default     = "2s"
}

# ==============================================================================
# POLICIES
# ==============================================================================

variable "configure_policies" {
  description = "Whether to configure custom channel policies"
  type        = bool
  default     = false
}

# ==============================================================================
# CHAINCODE CONFIGURATION
# ==============================================================================

variable "deploy_chaincode" {
  description = "Whether to deploy chaincode to the network"
  type        = bool
  default     = false
}

variable "chaincodes" {
  description = "Map of chaincodes to deploy on the network"
  type = map(object({
    name                  = string
    version               = string
    sequence              = number
    docker_image          = string
    chaincode_address     = string
    endorsement_policy    = optional(string)
    install_on_orgs       = list(string)           # List of org keys (e.g., ["org1", "org2"])
    install_on_peers      = optional(list(string)) # Optional: specific peer keys within orgs
    approve_with_orgs     = list(string)           # List of org keys to approve
    commit_with_org       = string                 # Org key to use for commit
    environment_variables = optional(map(string))
  }))

  default = {
    basic = {
      name               = "basic"
      version            = "1.0"
      sequence           = 1
      docker_image       = "hyperledger/fabric-samples-basic:latest"
      chaincode_address  = "basic.example.com:7052"
      endorsement_policy = ""
      install_on_orgs    = ["org1", "org2"]
      approve_with_orgs  = ["org1", "org2"]
      commit_with_org    = "org1"
      environment_variables = {
        CORE_CHAINCODE_LOGGING_LEVEL = "info"
      }
    }
  }
}
