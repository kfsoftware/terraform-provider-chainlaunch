# Email Notifications with Mailpit

This example demonstrates how to set up email notifications for Chainlaunch using Mailpit, a lightweight SMTP testing server with a web UI.

## Overview

This configuration:
1. Deploys a Mailpit Docker container for SMTP email capture and viewing
2. Configures a Chainlaunch notification provider to send alerts via SMTP
3. Sets up notification triggers for various events (backup failures, node downtime, etc.)

## What is Mailpit?

[Mailpit](https://github.com/axllent/mailpit) is a multi-platform email testing tool that:
- Captures all SMTP emails sent to it
- Provides a web UI to view and manage emails
- Requires no authentication (perfect for local development)
- Supports attachments, HTML emails, and more

## Prerequisites

- Docker installed and running
- Chainlaunch instance running (default: http://localhost:8100)
- Terraform installed

## Quick Start

```bash
cd examples/notifications-mailpit

# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Deploy Mailpit and configure notifications
terraform apply -auto-approve

# View the setup summary
terraform output setup_summary
```

## Accessing Mailpit

After running `terraform apply`, access the Mailpit web UI:

```bash
# Open in your browser
open http://localhost:8025
```

The web UI shows:
- All emails captured by Mailpit
- Email content (HTML and plain text)
- Attachments
- Email headers

## Testing Notifications

### Option 1: Trigger a Test Email from Chainlaunch UI

1. Log in to the Chainlaunch UI
2. Navigate to Settings → Notifications
3. Find your "Mailpit Email Notifications" provider
4. Click "Test" to send a test email
5. Check the Mailpit web UI to see the received email

### Option 2: Trigger Automatic Notifications

Enable automatic notifications by creating backup schedules or nodes:

```hcl
# Uncomment in main.tf to test backup notifications
resource "chainlaunch_backup_target" "test" {
  name               = "Test Backup Target"
  type               = "S3"
  region             = "us-east-1"
  access_key_id      = "test-access-key"
  secret_access_key  = "test-secret-key"
  bucket_name        = "test-bucket"
  restic_password    = "test-password"
}

resource "chainlaunch_backup_schedule" "test" {
  name            = "Test Backup Schedule"
  target_id       = chainlaunch_backup_target.test.id
  cron_expression = "*/5 * * * *" # Every 5 minutes (for testing)
  retention_days  = 7
}
```

When the backup runs (and likely fails due to invalid credentials), you'll receive a notification email in Mailpit.

## Configuration Options

### Notification Triggers

Control which events trigger email notifications:

```hcl
resource "chainlaunch_notification_provider" "mailpit" {
  # ...
  notify_backup_failure = true  # Alert on backup failures
  notify_backup_success = false # Alert on backup success (can be noisy)
  notify_node_downtime  = true  # Alert when nodes go down
  notify_s3_conn_issue  = true  # Alert on S3 connectivity problems
}
```

### Email Configuration

Customize sender and recipient:

```hcl
smtp_config = {
  from_email = "chainlaunch@example.com"
  from_name  = "Chainlaunch Alerts"
  to_email   = "admin@example.com"
  # ...
}
```

### Mailpit Ports

Change default ports if needed:

```hcl
variable "mailpit_smtp_port" {
  default = 1025  # SMTP port
}

variable "mailpit_web_port" {
  default = 8025  # Web UI port
}
```

## Docker-in-Docker Considerations

If Chainlaunch is running in a Docker container, you may need to adjust the SMTP host:

```hcl
# For Chainlaunch in Docker on macOS/Windows
variable "mailpit_host" {
  default = "host.docker.internal"
}

# For Chainlaunch in Docker on Linux
variable "mailpit_host" {
  default = "172.17.0.1"  # Docker bridge IP
}
```

## Using with Production SMTP Servers

While this example uses Mailpit for local testing, you can easily adapt it for production SMTP servers (Gmail, SendGrid, AWS SES, etc.):

### Gmail Example

```hcl
resource "chainlaunch_notification_provider" "gmail" {
  name = "Gmail Notifications"
  type = "SMTP"

  notify_backup_failure = true
  notify_node_downtime  = true
  is_default            = true

  smtp_config = {
    host       = "smtp.gmail.com"
    port       = 587
    username   = "your-email@gmail.com"
    password   = "your-app-password"  # Use App Password, not account password
    from_email = "your-email@gmail.com"
    from_name  = "Chainlaunch Alerts"
    to_email   = "admin@example.com"
    use_tls    = true
  }
}
```

### SendGrid Example

```hcl
resource "chainlaunch_notification_provider" "sendgrid" {
  name = "SendGrid Notifications"
  type = "SMTP"

  notify_backup_failure = true
  notify_node_downtime  = true
  is_default            = true

  smtp_config = {
    host       = "smtp.sendgrid.net"
    port       = 587
    username   = "apikey"
    password   = var.sendgrid_api_key
    from_email = "alerts@yourdomain.com"
    from_name  = "Chainlaunch Alerts"
    to_email   = "ops@yourdomain.com"
    use_tls    = true
  }
}
```

### AWS SES Example

```hcl
resource "chainlaunch_notification_provider" "ses" {
  name = "AWS SES Notifications"
  type = "SMTP"

  notify_backup_failure = true
  notify_node_downtime  = true
  is_default            = true

  smtp_config = {
    host       = "email-smtp.us-east-1.amazonaws.com"
    port       = 587
    username   = var.ses_smtp_username
    password   = var.ses_smtp_password
    from_email = "alerts@yourdomain.com"
    from_name  = "Chainlaunch Alerts"
    to_email   = "ops@yourdomain.com"
    use_tls    = true
  }
}
```

## Outputs

After deployment, view useful information:

```bash
# View Mailpit web UI URL
terraform output mailpit_web_ui

# View SMTP endpoint
terraform output mailpit_smtp_endpoint

# View notification provider details
terraform output notification_provider_id
terraform output notification_triggers

# View complete setup summary
terraform output setup_summary
```

## Cleanup

To remove all resources:

```bash
terraform destroy -auto-approve
```

This will:
1. Delete the notification provider from Chainlaunch
2. Stop and remove the Mailpit Docker container
3. Remove the Mailpit Docker image (if no other containers use it)

## Troubleshooting

### Mailpit container won't start

**Error:** Port already in use

**Solution:** Change the ports in `variables.tf`:
```hcl
variable "mailpit_smtp_port" {
  default = 1026  # Changed from 1025
}

variable "mailpit_web_port" {
  default = 8026  # Changed from 8025
}
```

### Emails not appearing in Mailpit

1. **Check Mailpit is running:**
   ```bash
   docker ps | grep mailpit
   ```

2. **Check Mailpit logs:**
   ```bash
   docker logs mailpit
   ```

3. **Verify SMTP endpoint:**
   ```bash
   terraform output mailpit_smtp_endpoint
   ```

4. **Test SMTP connection:**
   ```bash
   telnet localhost 1025
   ```

### Chainlaunch can't connect to Mailpit

If Chainlaunch is in Docker, use `host.docker.internal` (macOS/Windows) or the Docker bridge IP (Linux) instead of `localhost`.

## Advanced Configuration

### Multiple Email Recipients

To send notifications to multiple addresses, configure your SMTP server to forward emails or use email aliases.

### Different Notification Channels

For different types of notifications, create multiple providers:

```hcl
resource "chainlaunch_notification_provider" "critical" {
  name       = "Critical Alerts"
  type       = "SMTP"
  is_default = false

  notify_backup_failure = true
  notify_node_downtime  = true
  notify_s3_conn_issue  = false
  notify_backup_success = false

  smtp_config = {
    to_email = "oncall@example.com"
    # ...
  }
}

resource "chainlaunch_notification_provider" "info" {
  name       = "Info Notifications"
  type       = "SMTP"
  is_default = false

  notify_backup_failure = false
  notify_node_downtime  = false
  notify_s3_conn_issue  = false
  notify_backup_success = true

  smtp_config = {
    to_email = "ops@example.com"
    # ...
  }
}
```

## Related Examples

- [Backup with MinIO](../backup-with-minio/) - Test backup failure notifications
- [Metrics Monitoring](../metrics-monitoring/) - Monitor metrics that trigger alerts

## Resources

- [Mailpit GitHub](https://github.com/axllent/mailpit)
- [Mailpit Docker Hub](https://hub.docker.com/r/axllent/mailpit)
- [SMTP Protocol](https://en.wikipedia.org/wiki/Simple_Mail_Transfer_Protocol)
