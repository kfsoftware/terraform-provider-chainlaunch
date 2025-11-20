variable "chainlaunch_url" {
  description = "Chainlaunch instance URL"
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

variable "orderer_node_id" {
  description = "ID of the orderer node for the Fabric network"
  type        = string
}

variable "orderer_tls_ca_cert" {
  description = "TLS CA certificate for the orderer"
  type        = string
}

variable "org1_id" {
  description = "ID of the first organization"
  type        = string
}

variable "peer_node_id_org2" {
  description = "Peer node ID (connection ID) of Org2's Chainlaunch instance to share the network with"
  type        = string
}

variable "peer_node_id_partner" {
  description = "Peer node ID of partner's Chainlaunch instance for Besu network sharing"
  type        = string
}

variable "validator_key_id" {
  description = "Validator key ID for Besu network"
  type        = string
}
