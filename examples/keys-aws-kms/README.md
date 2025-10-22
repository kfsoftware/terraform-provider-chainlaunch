# AWS KMS Keys Example

This example demonstrates creating various types of cryptographic keys using AWS KMS (Key Management Service) as the key provider in Chainlaunch.

## What This Example Does

- Creates an AWS KMS key provider (using LocalStack for local testing)
- Creates multiple keys with different algorithms:
  - RSA keys (2048-bit and 4096-bit)
  - EC keys with NIST curves (P-256, P-384, P-521)
  - EC key with secp256k1 curve (Bitcoin/Ethereum)
- Stores all keys in AWS KMS for enhanced security
- Outputs details of the provider and all created keys

## Prerequisites

### LocalStack (for local testing)
This example uses LocalStack to emulate AWS KMS locally. To set it up:

1. **Install LocalStack**:
   ```bash
   pip install localstack
   ```

2. **Start LocalStack**:
   ```bash
   localstack start
   ```

3. **Verify KMS service is running**:
   ```bash
   aws --endpoint-url=http://localhost:4566 kms list-keys
   ```

### Production AWS KMS
For production use with real AWS KMS:

1. Update the `aws_kms_config` in main.tf:
   ```hcl
   aws_kms_config = {
     operation             = "IMPORT"
     aws_region            = "us-east-1"
     aws_access_key_id     = var.aws_access_key_id     # REQUIRED
     aws_secret_access_key = var.aws_secret_access_key # REQUIRED
     # Remove endpoint_url for production AWS
     kms_key_alias_prefix  = "chainlaunch/"
   }
   ```

2. **IMPORTANT**: AWS credentials are REQUIRED for IMPORT operation:
   - `aws_access_key_id` - AWS access key with KMS permissions
   - `aws_secret_access_key` - AWS secret access key

3. Alternative credential sources:
   - Via environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
   - Via ~/.aws/credentials file
   - Via IAM role (if running on EC2) - in this case, credentials in config are optional

## Key Types Demonstrated

### RSA Keys (Hardware Security Module backed)
- **RSA 2048**: Standard RSA key with 2048-bit key size
- **RSA 4096**: Larger RSA key with 4096-bit key size

### Elliptic Curve (EC) Keys
- **P-256**: NIST P-256 curve (secp256r1)
- **P-384**: NIST P-384 curve
- **P-521**: NIST P-521 curve
- **secp256k1**: Bitcoin/Ethereum curve (supported by AWS KMS)

**Note**: AWS KMS supports secp256k1, making it suitable for blockchain applications.

## Configuration

The example uses the following default values:

- **Chainlaunch URL**: `http://localhost:8100`
- **Username**: `admin`
- **Password**: `admin123`
- **AWS Region**: `us-east-1`
- **LocalStack Endpoint**: `http://localhost:4566`
- **KMS Key Alias Prefix**: `chainlaunch/`

## Usage

1. **Start LocalStack** (for local testing):
   ```bash
   localstack start
   ```

2. Navigate to this directory:
   ```bash
   cd examples/keys-aws-kms
   ```

3. Initialize Terraform:
   ```bash
   terraform init
   ```

4. Review the planned changes:
   ```bash
   terraform plan
   ```

   You should see:
   - 1 key provider to be created (AWS KMS)
   - 6 keys to be created

5. Create the provider and keys:
   ```bash
   terraform apply
   ```

6. View all outputs:
   ```bash
   terraform output
   ```

7. View summary:
   ```bash
   terraform output summary
   ```

## Expected Outputs

- `aws_kms_provider`: AWS KMS provider information
- `rsa_2048_key`: RSA 2048-bit key details
- `rsa_4096_key`: RSA 4096-bit key details
- `ec_p256_key`: EC P-256 key details
- `ec_p384_key`: EC P-384 key details
- `ec_p521_key`: EC P-521 key details
- `ec_secp256k1_key`: EC secp256k1 key details
- `summary`: Overview of all created resources

## Resources Created

- 1 AWS KMS key provider
- 6 cryptographic keys stored in AWS KMS

## AWS KMS Benefits

### Security Features
- **HSM-backed**: Keys are protected by FIPS 140-2 validated hardware
- **Never exported**: Private keys never leave the HSM
- **Audit trail**: All key operations are logged in CloudTrail
- **Access control**: Fine-grained IAM permissions

### Compliance
- FIPS 140-2 Level 2 (standard KMS)
- FIPS 140-2 Level 3 (CloudHSM)
- PCI-DSS, HIPAA, SOC compliant

### Availability
- Multi-AZ replication
- Automatic key rotation
- Disaster recovery capabilities

## secp256k1 Support

Unlike HashiCorp Vault, AWS KMS **does support** the secp256k1 curve:
- Used in Bitcoin and Ethereum
- Suitable for blockchain applications
- Can be used for cryptocurrency wallets
- Supported for signing operations

## Clean Up

To destroy all created resources:

```bash
terraform destroy
```

This will:
1. Delete all keys from AWS KMS
2. Remove the KMS key provider configuration

## Cost Considerations

### LocalStack (Free)
- No cost for local development
- Full KMS API emulation

### AWS KMS (Production)
- **Customer Managed Keys**: $1/month per key
- **API Requests**: $0.03 per 10,000 requests
- **Free Tier**: 20,000 free requests per month

Example monthly cost for this example:
- 6 keys × $1 = $6/month
- Plus API request costs (minimal for typical use)

## Troubleshooting

### Issue: "endpoint_url connection refused"
**Solution**: Ensure LocalStack is running on port 4566
```bash
localstack start
netstat -an | grep 4566
```

### Issue: "AccessDeniedException"
**Solution**: For production AWS, verify IAM permissions:
```json
{
  "Effect": "Allow",
  "Action": [
    "kms:CreateKey",
    "kms:CreateAlias",
    "kms:DescribeKey",
    "kms:GetPublicKey",
    "kms:Sign"
  ],
  "Resource": "*"
}
```

### Issue: "secp256k1 not supported"
**Solution**: This curve IS supported by AWS KMS. If you get this error, check:
- Chainlaunch version compatibility
- AWS region support (available in most regions)

### Issue: "KMS key not found after creation"
**Solution**:
- Keys may take a few seconds to propagate
- Check the key alias prefix matches configuration
- Verify KMS key state is "Enabled"

## Next Steps

After creating these keys:
1. Use them in Hyperledger Fabric organizations
2. Implement key rotation policies
3. Set up CloudTrail logging for audit
4. Configure key policies for access control
5. Enable automatic key rotation

## Production Checklist

Before using in production:
- [ ] Replace LocalStack endpoint with real AWS KMS
- [ ] Configure AWS credentials securely
- [ ] Set up CloudTrail for audit logging
- [ ] Define KMS key policies
- [ ] Enable key rotation
- [ ] Set up monitoring and alerts
- [ ] Document key usage and ownership
- [ ] Implement backup and recovery procedures
