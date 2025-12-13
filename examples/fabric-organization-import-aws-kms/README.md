# AWS KMS Import Example

This example demonstrates importing Hyperledger Fabric organizations with certificates stored in Chainlaunch and private keys managed by AWS KMS.

## Use Cases

- Organizations with AWS KMS-managed keys
- High-security deployments requiring HSM-backed key management
- Organizations with existing AWS KMS infrastructure
- Integration with AWS security services

## Prerequisites

1. Chainlaunch API running (default: `http://localhost:8100`)
2. AWS account with KMS access
3. AWS credentials configured (via AWS CLI, IAM role, or environment variables)
4. AWS KMS key provider configured in Chainlaunch
5. PEM-encoded certificates available
6. AWS KMS keys created (or will be created by Terraform)

## AWS Setup

### AWS Credentials

Configure AWS credentials before running Terraform:

```bash
# Using AWS CLI
aws configure

# Or environment variables
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"

# Or IAM role (if running on EC2/ECS)
```

### Create KMS Keys (Optional)

If using existing KMS keys, note their IDs or ARNs:

```bash
# Create a new KMS key
aws kms create-key --description "Fabric Org1 Signing CA"

# Create an alias for easier reference
aws kms create-alias --alias-name "alias/fabric-org1-sign-ca" --target-key-id "key-id"

# List existing KMS keys
aws kms list-keys
aws kms describe-key --key-id "arn:aws:kms:region:account:key/key-id"
```

## Development Setup

If you're developing the Chainlaunch provider locally:

1. Build the provider:
   ```bash
   cd /path/to/chainlaunch-terraform
   go build -o terraform-provider-chainlaunch
   ```

2. Configure `~/.terraformrc` with dev overrides (see aws-kms-localstack example for details)

3. Initialize Terraform (one time):
   ```bash
   terraform init
   ```

4. Run terraform commands:
   ```bash
   terraform plan
   terraform apply
   ```

After `terraform init` creates the lock file, subsequent `terraform plan/apply` commands will reload your local provider on each run (fast feedback loop).

For production, use `terraform init` to download from the registry.

## Usage

### Create terraform.tfvars

Create a `terraform.tfvars` file with your AWS account ID and KMS key IDs:

```hcl
aws_account_id           = "123456789012"
org1_sign_kms_key_id     = "12345678-1234-1234-1234-123456789012"
org1_tls_kms_key_id      = "87654321-4321-4321-4321-210987654321"
org2_sign_kms_key_id     = "arn:aws:kms:us-east-1:123456789012:key/aaaaaaaa-bbbb-cccc-dddd-111111111111"
org2_tls_kms_key_id      = "arn:aws:kms:us-east-1:123456789012:key/eeeeeeee-ffff-gggg-hhhh-222222222222"
```

### View the Plan

```bash
terraform plan
```

### Apply the Configuration

```bash
terraform apply
```

### Verify Import

```bash
terraform state show chainlaunch_fabric_organization_import.org1_kms
terraform output org1_kms_id
```

## Configuration

### With Existing KMS Keys (ARN Format)

```hcl
resource "chainlaunch_fabric_organization_import" "org1_kms" {
  msp_id       = "Org1MSP"
  name         = "organization-1"
  provider_id  = 10  # AWS KMS provider ID
  source_type  = "aws_kms"

  aws_kms_import {
    sign_ca_cert   = file("${path.module}/certs/org1-sign-ca.pem")
    sign_ca_key_id = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"

    tls_ca_cert   = file("${path.module}/certs/org1-tls-ca.pem")
    tls_ca_key_id = "arn:aws:kms:us-east-1:123456789012:key/87654321-4321-4321-4321-210987654321"
  }
}
```

### With KMS Key IDs

```hcl
resource "chainlaunch_fabric_organization_import" "org2_kms_keyid" {
  msp_id       = "Org2MSP"
  name         = "organization-2"
  provider_id  = 10  # AWS KMS provider ID
  source_type  = "aws_kms"

  aws_kms_import {
    sign_ca_cert   = file("${path.module}/certs/org2-sign-ca.pem")
    sign_ca_key_id = "12345678-1234-1234-1234-123456789012"

    tls_ca_cert   = file("${path.module}/certs/org2-tls-ca.pem")
    tls_ca_key_id = "87654321-4321-4321-4321-210987654321"
  }
}
```

### Create KMS Keys with Terraform

```hcl
# Create KMS keys
resource "aws_kms_key" "org3_sign_key" {
  description             = "KMS key for Org3 signing CA"
  deletion_window_in_days = 10
  enable_key_rotation     = true
}

resource "aws_kms_key" "org3_tls_key" {
  description             = "KMS key for Org3 TLS CA"
  deletion_window_in_days = 10
  enable_key_rotation     = true
}

# Create aliases
resource "aws_kms_alias" "org3_sign_alias" {
  name          = "alias/org3-sign-ca"
  target_key_id = aws_kms_key.org3_sign_key.key_id
}

# Import organization with created keys
resource "chainlaunch_fabric_organization_import" "org3_kms_created" {
  msp_id       = "Org3MSP"
  provider_id  = 10
  source_type  = "aws_kms"

  aws_kms_import {
    sign_ca_cert   = file("${path.module}/certs/org3-sign-ca.pem")
    sign_ca_key_id = aws_kms_key.org3_sign_key.arn

    tls_ca_cert   = file("${path.module}/certs/org3-tls-ca.pem")
    tls_ca_key_id = aws_kms_key.org3_tls_key.arn
  }

  depends_on = [
    aws_kms_alias.org3_sign_alias,
    aws_kms_alias.org3_tls_alias
  ]
}
```

## KMS Key Identifier Formats

Both key IDs and ARNs are supported:

### Key ID (Short Form)
```
12345678-1234-1234-1234-123456789012
```

### Key ARN (Full Form)
```
arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012
```

To find your KMS key ID or ARN:

```bash
# List all KMS keys
aws kms list-keys

# Describe a specific key
aws kms describe-key --key-id "12345678-1234-1234-1234-123456789012"

# Get the ARN
aws kms describe-key --key-id "alias/fabric-org1-sign-ca" --query 'KeyMetadata.Arn'
```

## Finding the AWS KMS Provider ID

To find your AWS KMS provider ID in Chainlaunch:

```bash
# Apply to retrieve provider data
terraform apply -target=data.chainlaunch_key_providers.awskms

# View the provider ID
terraform state show data.chainlaunch_key_providers.awskms
```

## AWS KMS Key Permissions

Ensure the AWS KMS keys have proper permissions. The KMS key policy should allow the Chainlaunch service to use the keys:

```bash
# View key policy
aws kms get-key-policy --key-id "12345678-1234-1234-1234-123456789012" --policy-name default

# Update key policy if needed (grant permissions to a principal)
aws kms create-grant \
  --key-id "12345678-1234-1234-1234-123456789012" \
  --grantee-principal "arn:aws:iam::123456789012:role/chainlaunch-role" \
  --operations "Decrypt" "GenerateDataKey"
```

## Troubleshooting

### Error: "InvalidKeyId.NotFound"

The KMS key doesn't exist or is not accessible:

```bash
# Verify key exists
aws kms describe-key --key-id "12345678-1234-1234-1234-123456789012"

# Check if key is enabled
aws kms describe-key --key-id "..." --query 'KeyMetadata.Enabled'

# Check IAM permissions
aws iam get-user  # Verify you have KMS access
```

### Error: "User is not authorized to use the key"

The AWS credentials don't have permission to use the KMS key:

```bash
# Check key policy
aws kms get-key-policy --key-id "..." --policy-name default

# Grant permissions (as key administrator)
aws kms create-grant \
  --key-id "..." \
  --grantee-principal "arn:aws:iam::123456789012:user/username" \
  --operations "Decrypt" "GenerateDataKey"
```

### Error: "sign_ca_key_id is required for AWS KMS imports"

The KMS key ID is missing or empty:

```hcl
aws_kms_import {
  sign_ca_cert   = file("certs/org1-sign-ca.pem")
  sign_ca_key_id = "12345678-1234-1234-1234-123456789012"  # ✅ Required
  tls_ca_cert    = file("certs/org1-tls-ca.pem")
  tls_ca_key_id  = "87654321-4321-4321-4321-210987654321"  # ✅ Required
}
```

### Terraform Plan Shows Key Recreation

If Terraform wants to recreate KMS keys, check:
- Key naming conventions
- Key rotation settings
- Tags and metadata

```bash
# Force refresh to avoid unnecessary recreation
terraform refresh

# Or apply with target to skip KMS keys
terraform apply -target=chainlaunch_fabric_organization_import.org1_kms
```

## Post-Import Operations

Once imported, use the organization with Chainlaunch identities and networks:

```hcl
# Create identities
resource "chainlaunch_fabric_identity" "org1_admin" {
  organization_id = chainlaunch_fabric_organization_import.org1_kms.id
  name            = "admin"
  role            = "admin"
}

# Add to Fabric network
resource "chainlaunch_fabric_network" "mychannel" {
  name = "mychannel"
  peer_organizations = [
    {
      id = chainlaunch_fabric_organization_import.org1_kms.id
    }
  ]
}
```

## KMS Cost Considerations

AWS KMS charges for:
- **Key storage**: $1/month per key
- **Requests**: $0.03 per 10,000 requests
- **Key rotation**: Free if automatic rotation enabled

For a production deployment:
- 2 keys (sign + TLS) per organization
- Estimated: $2/month per organization + request costs

## Security Best Practices

1. **Enable Key Rotation**: `enable_key_rotation = true`
2. **Use Key Policies**: Restrict to minimum required principals
3. **Monitor Key Usage**: Enable CloudTrail logging for KMS operations
4. **Backup Certificates**: Store certificate backups securely
5. **Audit Access**: Review KMS key grants regularly

```bash
# Enable CloudTrail logging
aws kms create-grant \
  --key-id "..." \
  --grantee-principal "arn:aws:cloudtrail:region:account:service-role/..." \
  --operations "Decrypt" "GenerateDataKey"

# List key grants
aws kms list-grants --key-id "..."

# Retire a grant when no longer needed
aws kms retire-grant --grant-token "..."
```

## Clean Up

To remove all imported organizations and created KMS keys:

```bash
terraform destroy
```

Note: By default, KMS keys have a 30-day deletion window. To immediately delete:

```bash
# Schedule KMS key deletion (7-30 day window)
aws kms schedule-key-deletion --key-id "..." --pending-window-in-days 7
```

## References

- [Fabric Organization Import Resource Documentation](https://registry.terraform.io/providers/kfsoftware/chainlaunch/latest/docs/resources/fabric_organization_import)
- [AWS KMS Documentation](https://docs.aws.amazon.com/kms/)
- [AWS Terraform KMS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key)
- [Hyperledger Fabric MSP Concepts](https://hyperledger-fabric.readthedocs.io/en/latest/msp.html)
