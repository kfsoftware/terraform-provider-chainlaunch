# ==============================================================================
# Notification Provider with Mailpit - Example
# ==============================================================================
# This example demonstrates setting up email notifications using Mailpit,
# a lightweight SMTP testing server with a web UI
# ==============================================================================

terraform {
  required_providers {
    chainlaunch = {
      source  = "registry.terraform.io/kfsoftware/chainlaunch"
      version = "0.1.0"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0"
    }
  }
}

provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

provider "docker" {
  host = "unix:///var/run/docker.sock"
}

# ==============================================================================
# STEP 1: Deploy Mailpit Container
# ==============================================================================

resource "docker_image" "mailpit" {
  name = "axllent/mailpit:${var.mailpit_version}"
}

resource "docker_container" "mailpit" {
  name  = var.mailpit_container_name
  image = docker_image.mailpit.image_id

  ports {
    internal = 1025 # SMTP port
    external = var.mailpit_smtp_port
    protocol = "tcp"
  }

  ports {
    internal = 8025 # Web UI port
    external = var.mailpit_web_port
    protocol = "tcp"
  }

  env = [
    "MP_SMTP_AUTH_ACCEPT_ANY=1",
    "MP_SMTP_AUTH_ALLOW_INSECURE=1",
  ]

  restart = "unless-stopped"
}

# ==============================================================================
# STEP 2: Configure Chainlaunch Notification Provider
# ==============================================================================

resource "chainlaunch_notification_provider" "mailpit" {
  name = "Mailpit Email Notifications"
  type = "SMTP"

  # Notification triggers
  notify_backup_failure = true
  notify_backup_success = var.notify_backup_success
  notify_node_downtime  = true
  notify_s3_conn_issue  = true

  # Make this the default provider
  is_default = true

  smtp_config = {
    host       = var.mailpit_host
    port       = var.mailpit_smtp_port
    from_email = var.from_email
    from_name  = var.from_name
    to_email   = var.to_email

    # Mailpit doesn't require authentication, but we can provide dummy credentials
    username = var.smtp_username
    password = var.smtp_password

    # Mailpit typically doesn't use TLS in local development
    use_tls         = var.use_tls
    skip_tls_verify = var.skip_tls_verify
  }

  depends_on = [docker_container.mailpit]
}

# ==============================================================================
# EXAMPLE: Add a Backup Target and Schedule to Test Notifications
# ==============================================================================
# Uncomment these resources to test backup notifications

# resource "chainlaunch_backup_target" "test" {
#   name               = "Test Backup Target"
#   type               = "S3"
#   region             = "us-east-1"
#   access_key_id      = "test-access-key"
#   secret_access_key  = "test-secret-key"
#   bucket_name        = "test-bucket"
#   restic_password    = "test-password"
#   endpoint           = "http://localhost:9000"
#   force_path_style   = true
# }

# resource "chainlaunch_backup_schedule" "test" {
#   name            = "Test Backup Schedule"
#   target_id       = chainlaunch_backup_target.test.id
#   cron_expression = "0 2 * * *" # Daily at 2 AM
#   retention_days  = 7
# }
