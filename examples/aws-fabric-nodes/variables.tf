variable "aws_region" {
  description = "AWS region where resources will be created"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "fabric-network"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "instance_type" {
  description = "EC2 instance type for Fabric nodes"
  type        = string
  default     = "t3.medium"
}

variable "key_pair_name" {
  description = "Name of the AWS EC2 key pair for SSH access"
  type        = string
}

variable "fabric_version" {
  description = "Hyperledger Fabric version to deploy"
  type        = string
  default     = "2.5.9"
}

variable "install_chainlaunch_server" {
  description = "Whether to install Chainlaunch on a new management server (true) or use an existing instance (false)"
  type        = bool
  default     = false
}

variable "management_instance_type" {
  description = "EC2 instance type for Chainlaunch management server"
  type        = string
  default     = "t3.medium"
}

variable "chainlaunch_version" {
  description = "Chainlaunch version to install (only used if install_chainlaunch_server is true)"
  type        = string
  default     = "latest"
}

variable "ssh_private_key_path" {
  description = "Path to SSH private key file (only used if install_chainlaunch_server is true)"
  type        = string
  default     = "~/.ssh/id_rsa"
}

variable "chainlaunch_url" {
  description = "Chainlaunch platform URL (only used if install_chainlaunch_server is false)"
  type        = string
  default     = ""
}

variable "chainlaunch_username" {
  description = "Chainlaunch username for authentication"
  type        = string
  default     = "admin"
}

variable "chainlaunch_password" {
  description = "Chainlaunch password for authentication"
  type        = string
  sensitive   = true
  default     = "admin123"
}

variable "peer_nodes" {
  description = "Map of peer nodes to create"
  type = map(object({
    organization = string
  }))
  default = {
    "peer0-org1" = {
      organization = "Org1"
    }
    "peer1-org1" = {
      organization = "Org1"
    }
    "peer0-org2" = {
      organization = "Org2"
    }
    "peer1-org2" = {
      organization = "Org2"
    }
  }
}

variable "orderer_nodes" {
  description = "Map of orderer nodes to create"
  type = map(object({
    organization = string
  }))
  default = {
    "orderer0" = {
      organization = "OrdererOrg"
    }
    "orderer1" = {
      organization = "OrdererOrg"
    }
    "orderer2" = {
      organization = "OrdererOrg"
    }
  }
}
