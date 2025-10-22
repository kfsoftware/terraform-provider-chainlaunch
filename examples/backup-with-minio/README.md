# Chainlaunch Backups with MinIO Example

This example demonstrates how to set up automated backups for Chainlaunch using MinIO as an S3-compatible storage backend. It includes:

- **MinIO Server**: S3-compatible object storage deployed via Docker
- **Backup Target**: S3 storage configuration for Chainlaunch backups
- **Backup Schedules**: Automated daily and weekly backup schedules

## What This Example Does

1. Deploys MinIO server in a Docker container with persistent storage
2. Creates an S3 bucket for backups using MinIO Client (mc)
3. Sets up MinIO user credentials for backup access
4. Configures Chainlaunch backup target pointing to MinIO
5. Creates two backup schedules:
   - **Daily backup** at 2:00 AM with 30-day retention
   - **Weekly backup** on Sundays at 3:00 AM with 90-day retention

## Prerequisites

- **Docker**: Running Docker daemon
- **Chainlaunch**: Running Chainlaunch instance (http://localhost:8100)
- **Terraform**: Terraform CLI installed
- **Docker Provider**: The example uses the `kreuzwerker/docker` provider

## Quick Start

```bash
# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Apply the configuration
terraform apply

# Access MinIO Console
# URL: http://localhost:9001
# Username: minioadmin
# Password: minioadmin
```

## Configuration

### MinIO Ports

- **9000**: MinIO S3 API (for backup operations)
- **9001**: MinIO Console (web UI for management)

### Backup Schedules

The example creates two backup schedules using cron expressions:

| Schedule | Cron Expression | Description | Retention |
|----------|----------------|-------------|-----------|
| Daily    | `0 2 * * *`    | Every day at 2:00 AM | 30 days |
| Weekly   | `0 3 * * 0`    | Every Sunday at 3:00 AM | 90 days |

### Cron Expression Format

```
 ┌───────────── minute (0 - 59)
 │ ┌───────────── hour (0 - 23)
 │ │ ┌───────────── day of the month (1 - 31)
 │ │ │ ┌───────────── month (1 - 12)
 │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
 │ │ │ │ │
 * * * * *
```

**Common Cron Examples:**
- `0 0 * * *` - Daily at midnight
- `0 2 * * *` - Daily at 2:00 AM
- `0 */6 * * *` - Every 6 hours
- `0 0 * * 0` - Weekly on Sunday at midnight
- `0 0 1 * *` - Monthly on the 1st at midnight

## Variables

You can customize the deployment by setting these variables:

```hcl
# Chainlaunch connection
chainlaunch_url      = "http://localhost:8100"
chainlaunch_username = "admin"
chainlaunch_password = "admin123"

# MinIO root credentials (for admin console)
minio_root_user     = "minioadmin"
minio_root_password = "minioadmin"

# MinIO backup user (used by Chainlaunch)
minio_access_key = "chainlaunch-backup"
minio_secret_key = "chainlaunch-backup-secret-key-123"

# Backup configuration
backup_bucket    = "chainlaunch-backups"
restic_password  = "super-secret-restic-password"
```

## Resource Configuration

### Backup Target

The backup target configures S3-compatible storage:

```hcl
resource "chainlaunch_backup_target" "minio_target" {
  name               = "MinIO Local Backup"
  type               = "S3"
  endpoint           = "http://localhost:9000"
  region             = "us-east-1"
  access_key_id      = var.minio_access_key
  secret_access_key  = var.minio_secret_key
  bucket_name        = var.backup_bucket
  bucket_path        = "fabric-backups"
  force_path_style   = true  # Required for MinIO
  restic_password    = var.restic_password
}
```

**Important Fields:**
- `endpoint`: Custom S3 endpoint (empty for AWS S3)
- `force_path_style`: **Must be `true` for MinIO** and most S3-compatible services
- `restic_password`: Used to encrypt backups with Restic
- `bucket_path`: Optional path within the bucket to organize backups

### Backup Schedule

Backup schedules define when backups run automatically:

```hcl
resource "chainlaunch_backup_schedule" "daily_backup" {
  name            = "Daily Fabric Backup"
  description     = "Automated daily backup at 2 AM"
  target_id       = chainlaunch_backup_target.minio_target.id
  cron_expression = "0 2 * * *"
  enabled         = true
  retention_days  = 30
}
```

**Important Fields:**
- `target_id`: References the backup target where backups are stored
- `cron_expression`: Defines the schedule in cron format
- `enabled`: Can be set to `false` to temporarily disable the schedule
- `retention_days`: Backups older than this will be automatically deleted

## Accessing MinIO Console

After applying the configuration, you can access the MinIO web console:

1. **Open browser**: http://localhost:9001
2. **Login** with credentials:
   - Username: `minioadmin`
   - Password: `minioadmin`

In the console, you can:
- Browse backup files in the `chainlaunch-backups` bucket
- View bucket statistics and usage
- Manage access policies
- Monitor backup operations

## Using AWS S3 Instead of MinIO

To use AWS S3 instead of MinIO, simply create the backup target without the MinIO infrastructure:

```hcl
resource "chainlaunch_backup_target" "aws_s3_target" {
  name               = "AWS S3 Production Backups"
  type               = "S3"
  # No endpoint needed for AWS S3
  region             = "us-east-1"
  access_key_id      = var.aws_access_key_id
  secret_access_key  = var.aws_secret_access_key
  bucket_name        = "my-production-backups"
  bucket_path        = "chainlaunch/fabric"
  force_path_style   = false  # AWS S3 uses virtual-hosted style
  restic_password    = var.restic_password
}
```

**AWS S3 Differences:**
- No `endpoint` field (uses default AWS S3 endpoint)
- `force_path_style` should be `false`
- `region` must be a valid AWS region (e.g., `us-east-1`, `eu-west-1`)
- Requires valid AWS credentials with S3 permissions

## Backup Encryption

All backups are encrypted using [Restic](https://restic.net/), a secure backup program. The `restic_password` is used to encrypt and decrypt backups.

**Important**:
- Store the Restic password securely (e.g., in a secrets manager)
- Without this password, backups cannot be restored
- Use a strong, random password

## Monitoring Backups

After creating schedules, you can monitor them:

```bash
# View next scheduled run times
terraform output daily_schedule_next_run
terraform output weekly_schedule_next_run

# Check backup target ID
terraform output backup_target_id
```

You can also use the Chainlaunch API to:
- List all backups
- View backup status
- Trigger manual backups
- Restore from backups

## Cleanup

To remove all resources:

```bash
terraform destroy
```

This will:
1. Delete backup schedules in Chainlaunch
2. Delete the backup target in Chainlaunch
3. Stop and remove the MinIO container
4. Remove the Docker volume (backup data will be lost)

## Troubleshooting

### MinIO Container Won't Start

Check Docker logs:
```bash
docker logs minio-backup-storage
```

### Backup Schedule Not Running

1. Check if the schedule is enabled:
   ```bash
   terraform state show chainlaunch_backup_schedule.daily_backup
   ```

2. Verify the cron expression using an online validator

3. Check Chainlaunch logs for backup errors

### Connection to MinIO Fails

1. Verify MinIO is running:
   ```bash
   docker ps | grep minio
   ```

2. Test MinIO connectivity:
   ```bash
   curl http://localhost:9000/minio/health/live
   ```

3. Check `force_path_style` is set to `true` for MinIO

### Bucket Not Found

Ensure the bucket creation container succeeded:
```bash
docker ps -a | grep mc-create-bucket
```

Re-run bucket creation if needed:
```bash
terraform taint docker_container.mc_create_bucket
terraform apply
```

## Security Considerations

### Production Deployments

For production use, consider:

1. **Secure Credentials**: Use Terraform secrets management
   ```hcl
   # Use environment variables
   variable "minio_secret_key" {
     type      = string
     sensitive = true
     # Set via TF_VAR_minio_secret_key environment variable
   }
   ```

2. **Network Security**:
   - Use HTTPS endpoints
   - Restrict network access to MinIO
   - Use VPC/private networking

3. **Access Control**:
   - Use minimal IAM permissions
   - Rotate access keys regularly
   - Enable MFA for admin access

4. **Backup Validation**:
   - Test restores regularly
   - Monitor backup sizes and success rates
   - Set up alerts for failed backups

## Additional Resources

- [MinIO Documentation](https://min.io/docs/minio/linux/index.html)
- [Restic Documentation](https://restic.readthedocs.io/)
- [Cron Expression Guide](https://crontab.guru/)
- [Terraform Docker Provider](https://registry.terraform.io/providers/kreuzwerker/docker/latest/docs)

## Example Outputs

After applying, you'll see:

```
minio_console_url = "http://localhost:9001"
minio_api_url = "http://localhost:9000"
backup_target_id = "1"
daily_schedule_id = "1"
daily_schedule_next_run = "2025-10-20T02:00:00Z"
weekly_schedule_id = "2"
weekly_schedule_next_run = "2025-10-24T03:00:00Z"
```
