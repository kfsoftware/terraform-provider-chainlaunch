output "plugin_name" {
  description = "Name of the registered plugin"
  value       = chainlaunch_plugin.hlf_api.name
}

output "plugin_version" {
  description = "Version of the plugin"
  value       = chainlaunch_plugin.hlf_api.metadata_version
}

output "plugin_description" {
  description = "Description of the plugin"
  value       = chainlaunch_plugin.hlf_api.description
}

output "plugin_author" {
  description = "Author of the plugin"
  value       = chainlaunch_plugin.hlf_api.author
}

output "plugin_repository" {
  description = "Repository URL of the plugin"
  value       = chainlaunch_plugin.hlf_api.repository
}

output "plugin_api_version" {
  description = "API version of the plugin"
  value       = chainlaunch_plugin.hlf_api.api_version
}

output "summary" {
  description = "Plugin registration summary"
  value       = <<-EOT

    ╔══════════════════════════════════════════════════════════════╗
    ║              Plugin Registered Successfully                  ║
    ╚══════════════════════════════════════════════════════════════╝

    Plugin Details:
      Name:        ${chainlaunch_plugin.hlf_api.name}
      Version:     ${chainlaunch_plugin.hlf_api.metadata_version}
      API Version: ${chainlaunch_plugin.hlf_api.api_version}
      Description: ${chainlaunch_plugin.hlf_api.description}
      Author:      ${chainlaunch_plugin.hlf_api.author}
      Repository:  ${chainlaunch_plugin.hlf_api.repository}
      License:     ${chainlaunch_plugin.hlf_api.license}

    Next Steps:
    1. Query the plugin:
       terraform output plugin_name

    2. Deploy the plugin using the plugin-deployment example:
       cd ../plugin-deployment
       terraform apply -var="plugin_name=${chainlaunch_plugin.hlf_api.name}"

    3. Or query it via API:
       curl -s -u admin:admin123 \\
         http://localhost:8100/api/v1/plugins/${chainlaunch_plugin.hlf_api.name}

  EOT
}
