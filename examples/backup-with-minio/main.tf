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
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
    }
  }
}

# Configure the Chainlaunch Provider
provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# Configure Docker Provider
provider "docker" {
  host = "unix:///var/run/docker.sock"
}

# Variables
variable "chainlaunch_url" {
  description = "Chainlaunch API URL"
  type        = string
  default     = "http://localhost:8100"
}

variable "chainlaunch_username" {
  description = "Chainlaunch Username"
  type        = string
  default     = "admin"
}

variable "chainlaunch_password" {
  description = "Chainlaunch Password"
  type        = string
  sensitive   = true
  default     = "admin123"
}

variable "minio_root_user" {
  description = "MinIO root username"
  type        = string
  default     = "minioadmin"
}

variable "minio_root_password" {
  description = "MinIO root password"
  type        = string
  sensitive   = true
  default     = "minioadmin"
}

variable "minio_access_key" {
  description = "MinIO access key for backups"
  type        = string
  default     = "chainlaunch-backup"
}

variable "minio_secret_key" {
  description = "MinIO secret key for backups"
  type        = string
  sensitive   = true
  default     = "chainlaunch-backup-secret-key-123"
}

variable "backup_bucket" {
  description = "S3 bucket name for backups"
  type        = string
  default     = "chainlaunch-backups"
}

variable "restic_password" {
  description = "Password for encrypting backups with Restic"
  type        = string
  sensitive   = true
  default     = "super-secret-restic-password"
}

# Create a Docker network for MinIO
resource "docker_network" "minio_network" {
  name = "minio-network"
}

# Deploy MinIO as a Docker container
resource "docker_image" "minio" {
  name         = "minio/minio:latest"
  keep_locally = true
}

resource "docker_container" "minio" {
  name  = "minio-backup-storage"
  image = docker_image.minio.image_id

  networks_advanced {
    name = docker_network.minio_network.name
  }

  ports {
    internal = 9000
    external = 9100 # Non-standard port to avoid conflicts
  }

  ports {
    internal = 9001
    external = 9101 # Non-standard port for console
  }

  env = [
    "MINIO_ROOT_USER=${var.minio_root_user}",
    "MINIO_ROOT_PASSWORD=${var.minio_root_password}",
  ]

  command = [
    "server",
    "/data",
    "--console-address",
    ":9001"
  ]

  volumes {
    volume_name    = docker_volume.minio_data.name
    container_path = "/data"
  }

  restart = "unless-stopped"
}

resource "docker_volume" "minio_data" {
  name = "minio-backup-data"
}

# Wait for MinIO to be ready
resource "time_sleep" "wait_for_minio" {
  depends_on = [docker_container.minio]

  create_duration = "10s"
}

# Setup MinIO bucket and user using local-exec (one-time operations)
resource "null_resource" "minio_setup" {
  provisioner "local-exec" {
    command = <<-EOT
      docker run --rm --network minio-network minio/mc:latest alias set myminio http://minio-backup-storage:9000 ${var.minio_root_user} ${var.minio_root_password} && \
      docker run --rm --network minio-network minio/mc:latest mb --ignore-existing myminio/${var.backup_bucket} && \
      docker run --rm --network minio-network minio/mc:latest admin user add myminio ${var.minio_access_key} ${var.minio_secret_key} || true && \
      docker run --rm --network minio-network minio/mc:latest admin policy attach myminio readwrite --user ${var.minio_access_key} || true
    EOT
  }

  depends_on = [time_sleep.wait_for_minio]

  triggers = {
    minio_container_id = docker_container.minio.id
  }
}

# Configure Chainlaunch backup target pointing to MinIO
resource "chainlaunch_backup_target" "minio_target" {
  name              = "MinIO Local Backup"
  type              = "S3"
  endpoint          = "http://localhost:9100" # Updated to use non-standard port
  region            = "us-east-1"
  access_key_id     = var.minio_access_key
  secret_access_key = var.minio_secret_key
  bucket_name       = var.backup_bucket
  bucket_path       = "fabric-backups"
  force_path_style  = true # Required for MinIO
  restic_password   = var.restic_password

  depends_on = [null_resource.minio_setup]
}

# Create a daily backup schedule
resource "chainlaunch_backup_schedule" "daily_backup" {
  name            = "Daily Fabric Backup"
  description     = "Automated daily backup at 2 AM"
  target_id       = chainlaunch_backup_target.minio_target.id
  cron_expression = "0 2 * * *" # Every day at 2:00 AM
  enabled         = true
  retention_days  = 30
}

# Create a weekly backup schedule (for longer retention)
resource "chainlaunch_backup_schedule" "weekly_backup" {
  name            = "Weekly Fabric Backup"
  description     = "Automated weekly backup on Sundays at 3 AM"
  target_id       = chainlaunch_backup_target.minio_target.id
  cron_expression = "0 3 * * 0" # Every Sunday at 3:00 AM
  enabled         = true
  retention_days  = 90
}

# Outputs
output "minio_console_url" {
  description = "MinIO Console URL"
  value       = "http://localhost:9101"
}

output "minio_api_url" {
  description = "MinIO API URL"
  value       = "http://localhost:9100"
}

output "minio_credentials" {
  description = "MinIO login credentials"
  value = {
    username = var.minio_root_user
    password = var.minio_root_password
  }
  sensitive = true
}

output "backup_target_id" {
  description = "ID of the backup target"
  value       = chainlaunch_backup_target.minio_target.id
}

output "daily_schedule_id" {
  description = "ID of the daily backup schedule"
  value       = chainlaunch_backup_schedule.daily_backup.id
}

output "daily_schedule_next_run" {
  description = "Next run time for daily backup"
  value       = chainlaunch_backup_schedule.daily_backup.next_run_at
}

output "weekly_schedule_id" {
  description = "ID of the weekly backup schedule"
  value       = chainlaunch_backup_schedule.weekly_backup.id
}

output "weekly_schedule_next_run" {
  description = "Next run time for weekly backup"
  value       = chainlaunch_backup_schedule.weekly_backup.next_run_at
}
