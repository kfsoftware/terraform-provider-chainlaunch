# AWS KMS Key Provider Example

This example demonstrates how to create an AWS KMS key provider in Chainlaunch and use it for organization key management.

## What This Example Does

- Creates an AWS KMS key provider configured to use LocalStack (or real AWS KMS)
- Creates a Fabric organization that uses the AWS KMS provider for cryptographic operations
- Demonstrates the full lifecycle of provider and organization creation

## Prerequisites

For this example to work fully, you need either:

1. **LocalStack** running locally:
   ```bash
   docker run -d -p 4566:4566 localstack/localstack
   ```

2. **Real AWS KMS** with proper credentials configured

## Configuration

### AWS KMS Provider Configuration

The example creates an AWS KMS provider with the following settings:

```hcl
aws_kms_config = {
  operation             = "IMPORT"        # Use existing KMS keys
  aws_region            = "us-east-1"     # AWS region
  aws_access_key_id     = "test"          # REQUIRED for IMPORT
  aws_secret_access_key = "test"          # REQUIRED for IMPORT
  endpoint_url          = "http://localhost:4566"  # LocalStack endpoint
  kms_key_alias_prefix  = "chainlaunch/"  # Key alias prefix
}
```

### Available AWS KMS Options

- **operation**: `IMPORT` (use existing keys) or `CREATE` (create new keys)
- **aws_region** (required): AWS region (e.g., "us-east-1", "eu-west-1")
- **aws_access_key_id** (required for IMPORT): AWS access key
- **aws_secret_access_key** (required for IMPORT): AWS secret key
- **endpoint_url** (optional): Custom endpoint for LocalStack or private AWS endpoints
- **aws_session_token** (optional): Temporary session token
- **assume_role_arn** (optional): IAM role ARN for cross-account access
- **external_id** (optional): External ID for role assumption
- **kms_key_alias_prefix** (optional): Prefix for KMS key aliases (default: "chainlaunch/")

## Usage

1. **Start LocalStack** (optional, for testing):
   ```bash
   docker run -d -p 4566:4566 localstack/localstack
   ```

2. **Navigate to this directory**:
   ```bash
   cd examples/aws-kms-provider
   ```

3. **Review the configuration**:
   ```bash
   terraform plan
   ```

4. **Apply the configuration**:
   ```bash
   terraform apply
   ```

   Expected output:
   - Key provider created with ID
   - Organization creation may fail if LocalStack isn't running (this is expected)

5. **View outputs**:
   ```bash
   terraform output
   ```

## Example with Real AWS KMS

To use real AWS KMS instead of LocalStack:

```hcl
resource "chainlaunch_key_provider" "aws_kms_prod" {
  name       = "ProductionAWSKMS"
  type       = "AWS_KMS"
  is_default = false

  aws_kms_config = {
    operation    = "IMPORT"
    aws_region   = "us-east-1"

    # Option 1: Use IAM roles (recommended for EC2/ECS)
    # No credentials needed - use instance profile

    # Option 2: Use static credentials
    # aws_access_key_id     = var.aws_access_key_id
    # aws_secret_access_key = var.aws_secret_access_key

    # Option 3: Cross-account access
    # assume_role_arn = "arn:aws:iam::123456789012:role/KMSRole"
    # external_id     = "unique-external-id"

    kms_key_alias_prefix = "chainlaunch/prod/"
  }
}
```

## Expected Outputs

- `key_provider_id`: The ID of the created AWS KMS provider (e.g., "4")
- `key_provider_type`: The type ("AWS_KMS")
- `organization_id`: The ID of the organization (if successfully created)
- `organization_provider_id`: The provider ID used by the organization

## Resources Created

1. **AWS KMS Key Provider**: Configured to communicate with AWS KMS (or LocalStack)
2. **Fabric Organization**: Using the AWS KMS provider for key management

## Troubleshooting

### Error: "no EC2 IMDS role found"

This means the Chainlaunch instance cannot find AWS credentials. Solutions:

1. **For LocalStack testing**: Ensure LocalStack is running on port 4566
2. **For real AWS**: Configure AWS credentials in one of these ways:
   - Add `aws_access_key_id` and `aws_secret_access_key` to the config
   - Use IAM instance profile (if running on EC2)
   - Use `assume_role_arn` for cross-account access

### Error: "request canceled, context deadline exceeded"

This usually means:
- LocalStack is not running (if using endpoint_url)
- Network connectivity issues to AWS KMS
- Invalid endpoint URL

### Provider Created But Organization Failed

This is expected if KMS isn't properly configured. The key provider resource was successfully created! To verify:

```bash
export CHAINLAUNCH_USER=""
export CHAINLAUNCH_PASSWORD=""
curl -s http://localhost:8100/api/v1/key-providers/[ID] -u $CHAINLAUNCH_USER:$CHAINLAUNCH_PASSWORD
```

## Security Best Practices

1. **Never commit AWS credentials** to version control
2. **Use IAM roles** instead of static credentials when possible
3. **Store credentials in variables**:
   ```hcl
   variable "aws_access_key_id" {
     sensitive = true
   }
   variable "aws_secret_access_key" {
     sensitive = true
   }
   ```
4. **Use assume_role_arn** for cross-account access with external_id
5. **Restrict KMS key permissions** to only what's needed

## Clean Up

To destroy the resources:

```bash
terraform destroy
```

Note: If organization creation failed, you may only need to delete the key provider.

## Additional Resources

- [AWS KMS Documentation](https://docs.aws.amazon.com/kms/)
- [LocalStack KMS](https://docs.localstack.cloud/user-guide/aws/kms/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Chainlaunch API Documentation](http://localhost:8100/api-docs)
