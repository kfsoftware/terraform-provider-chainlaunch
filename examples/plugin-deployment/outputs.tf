output "plugin_name" {
  description = "Name of the deployed plugin"
  value       = data.chainlaunch_plugin.hlf_api.name
}

output "plugin_version" {
  description = "Version of the deployed plugin"
  value       = data.chainlaunch_plugin.hlf_api.metadata_version
}

output "plugin_description" {
  description = "Description of the deployed plugin"
  value       = data.chainlaunch_plugin.hlf_api.description
}

output "deployment_status" {
  description = "Status of the plugin deployment"
  value       = chainlaunch_plugin_deployment.hlf_api.status
}

output "deployment_project_name" {
  description = "Docker Compose project name for the deployment"
  value       = chainlaunch_plugin_deployment.hlf_api.project_name
}

output "deployment_started_at" {
  description = "Timestamp when deployment started"
  value       = chainlaunch_plugin_deployment.hlf_api.started_at
}

output "api_endpoint" {
  description = "API endpoint URL"
  value       = "http://localhost:${var.api_port}"
}

output "peer0_endpoint" {
  description = "External endpoint of the first peer"
  value       = data.chainlaunch_fabric_peer.peer0.external_endpoint
}

output "peer1_endpoint" {
  description = "External endpoint of the second peer (if configured)"
  value       = var.peer1_name != "" ? data.chainlaunch_fabric_peer.peer1[0].external_endpoint : null
}

output "identity_id" {
  description = "ID of the created admin identity"
  value       = chainlaunch_fabric_identity.api_admin.id
}

output "identity_name" {
  description = "Name of the created admin identity"
  value       = chainlaunch_fabric_identity.api_admin.name
}

output "identity_certificate" {
  description = "Certificate of the admin identity"
  value       = chainlaunch_fabric_identity.api_admin.certificate
  sensitive   = true
}

output "setup_summary" {
  description = "Setup summary"
  value       = <<-EOT

    ╔══════════════════════════════════════════════════════════════╗
    ║        Hyperledger Fabric API Plugin - Deployed             ║
    ╚══════════════════════════════════════════════════════════════╝

    Plugin Information:
      Name:        ${data.chainlaunch_plugin.hlf_api.name}
      Version:     ${data.chainlaunch_plugin.hlf_api.metadata_version}
      Description: ${data.chainlaunch_plugin.hlf_api.description}
      Author:      ${data.chainlaunch_plugin.hlf_api.author}
      Repository:  ${data.chainlaunch_plugin.hlf_api.repository}
      License:     ${data.chainlaunch_plugin.hlf_api.license}

    Deployment Status:
      Status:       ${chainlaunch_plugin_deployment.hlf_api.status}
      Project Name: ${chainlaunch_plugin_deployment.hlf_api.project_name}
      Started At:   ${chainlaunch_plugin_deployment.hlf_api.started_at}

    API Configuration:
      Endpoint:     http://localhost:${var.api_port}
      Organization: ${var.organization_msp_id}
      Channel:      ${var.channel_name}
      Peer(s):      ${data.chainlaunch_fabric_peer.peer0.name}${var.peer1_name != "" ? format(", %s", var.peer1_name) : ""}

    Identity Information:
      ID:           ${chainlaunch_fabric_identity.api_admin.id}
      Name:         ${chainlaunch_fabric_identity.api_admin.name}
      Role:         ${chainlaunch_fabric_identity.api_admin.role}
      Algorithm:    ${chainlaunch_fabric_identity.api_admin.algorithm}

    Next Steps:
    1. Verify deployment: docker ps | grep hlf-plugin-api
    2. Check logs: docker logs $(docker ps -q -f name=hlf-plugin-api)
    3. Test API health: curl http://localhost:${var.api_port}/api/v1/health
    4. View metrics: curl http://localhost:${var.api_port}/metrics

  EOT
}
