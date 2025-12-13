# ==============================================================================
# AWS INFRASTRUCTURE OUTPUTS
# ==============================================================================

output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.fabric_vpc.id
}

output "subnet_id" {
  description = "ID of the public subnet"
  value       = aws_subnet.fabric_public_subnet.id
}

# ==============================================================================
# CHAINLAUNCH MANAGEMENT SERVER OUTPUTS
# ==============================================================================

output "chainlaunch_server" {
  description = "Chainlaunch management server details (if installed)"
  value = var.install_chainlaunch_server ? {
    instance_id       = aws_instance.chainlaunch_management[0].id
    public_ip         = aws_instance.chainlaunch_management[0].public_ip
    private_ip        = aws_instance.chainlaunch_management[0].private_ip
    chainlaunch_url   = chainlaunch_install_ssh.management_server[0].chainlaunch_url
    service_status    = chainlaunch_install_ssh.management_server[0].service_status
    installed_version = chainlaunch_install_ssh.management_server[0].installed_version
    ssh_command       = "ssh -i ${var.ssh_private_key_path} ubuntu@${aws_instance.chainlaunch_management[0].public_ip}"
  } : null
}

# ==============================================================================
# PEER NODE OUTPUTS
# ==============================================================================

output "peer_instances" {
  description = "Details of peer EC2 instances"
  value = {
    for key, instance in aws_instance.fabric_peers : key => {
      instance_id = instance.id
      public_ip   = instance.public_ip
      private_ip  = instance.private_ip
      public_dns  = instance.public_dns
    }
  }
}

output "peer_endpoints" {
  description = "Peer node endpoints"
  value = {
    for key, peer in chainlaunch_fabric_peer.peers : key => {
      peer_id           = peer.id
      external_endpoint = peer.external_endpoint
      organization      = peer.msp_id
      status            = peer.status
    }
  }
}

output "peer_connection_strings" {
  description = "Connection strings for peer nodes (for client applications)"
  value = {
    for key, instance in aws_instance.fabric_peers :
    key => "grpcs://${instance.public_ip}:7051"
  }
}

# ==============================================================================
# ORDERER NODE OUTPUTS
# ==============================================================================

output "orderer_instances" {
  description = "Details of orderer EC2 instances"
  value = {
    for key, instance in aws_instance.fabric_orderers : key => {
      instance_id = instance.id
      public_ip   = instance.public_ip
      private_ip  = instance.private_ip
      public_dns  = instance.public_dns
    }
  }
}

output "orderer_endpoints" {
  description = "Orderer node endpoints"
  value = {
    for key, orderer in chainlaunch_fabric_orderer.orderers : key => {
      orderer_id        = orderer.id
      external_endpoint = orderer.external_endpoint
      organization      = orderer.msp_id
      status            = orderer.status
    }
  }
}

output "orderer_connection_strings" {
  description = "Connection strings for orderer nodes (for client applications)"
  value = {
    for key, instance in aws_instance.fabric_orderers :
    key => "grpcs://${instance.public_ip}:7050"
  }
}

# ==============================================================================
# ORGANIZATION OUTPUTS
# ==============================================================================

output "peer_organizations" {
  description = "Peer organization details"
  value = {
    for key, org in chainlaunch_fabric_organization.peer_orgs : key => {
      id         = org.id
      msp_id     = org.msp_id
      created_at = org.created_at
    }
  }
}

output "orderer_organizations" {
  description = "Orderer organization details"
  value = {
    for key, org in chainlaunch_fabric_organization.orderer_orgs : key => {
      id         = org.id
      msp_id     = org.msp_id
      created_at = org.created_at
    }
  }
}

# ==============================================================================
# QUICK REFERENCE OUTPUTS
# ==============================================================================

output "ssh_commands" {
  description = "SSH commands to connect to instances"
  value = merge(
    {
      for key, instance in aws_instance.fabric_peers :
      "ssh_${key}" => "ssh -i ~/.ssh/${var.key_pair_name}.pem ubuntu@${instance.public_ip}"
    },
    {
      for key, instance in aws_instance.fabric_orderers :
      "ssh_${key}" => "ssh -i ~/.ssh/${var.key_pair_name}.pem ubuntu@${instance.public_ip}"
    }
  )
}

output "network_summary" {
  description = "Summary of the deployed network"
  value = {
    peer_count    = length(aws_instance.fabric_peers)
    orderer_count = length(aws_instance.fabric_orderers)
    peer_orgs     = length(chainlaunch_fabric_organization.peer_orgs)
    orderer_orgs  = length(chainlaunch_fabric_organization.orderer_orgs)
    fabric_version = var.fabric_version
    aws_region    = var.aws_region
  }
}
