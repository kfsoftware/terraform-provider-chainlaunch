# ==============================================================================
# Mailpit Outputs
# ==============================================================================

output "mailpit_web_ui" {
  description = "Mailpit web UI URL"
  value       = "http://localhost:${var.mailpit_web_port}"
}

output "mailpit_smtp_endpoint" {
  description = "Mailpit SMTP endpoint"
  value       = "${var.mailpit_host}:${var.mailpit_smtp_port}"
}

output "mailpit_container_id" {
  description = "Mailpit Docker container ID"
  value       = docker_container.mailpit.id
}

output "mailpit_container_name" {
  description = "Mailpit Docker container name"
  value       = docker_container.mailpit.name
}

# ==============================================================================
# Notification Provider Outputs
# ==============================================================================

output "notification_provider_id" {
  description = "Notification provider ID"
  value       = chainlaunch_notification_provider.mailpit.id
}

output "notification_provider_name" {
  description = "Notification provider name"
  value       = chainlaunch_notification_provider.mailpit.name
}

output "notification_provider_created_at" {
  description = "When the notification provider was created"
  value       = chainlaunch_notification_provider.mailpit.created_at
}

output "notification_triggers" {
  description = "Enabled notification triggers"
  value = {
    backup_success = chainlaunch_notification_provider.mailpit.notify_backup_success
    backup_failure = chainlaunch_notification_provider.mailpit.notify_backup_failure
    node_downtime  = chainlaunch_notification_provider.mailpit.notify_node_downtime
    s3_conn_issue  = chainlaunch_notification_provider.mailpit.notify_s3_conn_issue
  }
}

output "setup_summary" {
  description = "Setup summary"
  value       = <<-EOT

    ╔══════════════════════════════════════════════════════════════╗
    ║        Mailpit Email Notifications - Setup Complete         ║
    ╚══════════════════════════════════════════════════════════════╝

    Mailpit Web UI: http://localhost:${var.mailpit_web_port}
    SMTP Endpoint:  ${var.mailpit_host}:${var.mailpit_smtp_port}

    Notification Provider:
      Name:        ${chainlaunch_notification_provider.mailpit.name}
      Type:        ${chainlaunch_notification_provider.mailpit.type}
      Default:     ${chainlaunch_notification_provider.mailpit.is_default ? "Yes" : "No"}

    Email Configuration:
      From: ${var.from_name} <${var.from_email}>
      To:   ${var.to_email}

    Enabled Notifications:
      ✓ Backup Failures:   ${chainlaunch_notification_provider.mailpit.notify_backup_failure ? "✅" : "❌"}
      ✓ Backup Success:    ${chainlaunch_notification_provider.mailpit.notify_backup_success ? "✅" : "❌"}
      ✓ Node Downtime:     ${chainlaunch_notification_provider.mailpit.notify_node_downtime ? "✅" : "❌"}
      ✓ S3 Conn Issues:    ${chainlaunch_notification_provider.mailpit.notify_s3_conn_issue ? "✅" : "❌"}

    Next Steps:
    1. Open http://localhost:${var.mailpit_web_port} to view the Mailpit web UI
    2. Trigger a test notification from the Chainlaunch UI
    3. View received emails in the Mailpit inbox

  EOT
}
