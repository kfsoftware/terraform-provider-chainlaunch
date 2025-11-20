# AWS KMS LocalStack Import Example

This example demonstrates a complete workflow:
1. **Create** an organization in Node 1 (source)
2. **Export** certificates from the source organization
3. **Import** the same organization into Node 2 using AWS KMS (via LocalStack) for key management

This is useful for testing multi-node setups, disaster recovery scenarios, and organization replication.

## Architecture

```
Node 1 (Create)                Node 2 (Import)
├── Source Organization        ├── Imported Organization
├── Database Provider          ├── LocalStack KMS Provider
└── Exports Certificates       └── Keys managed by LocalStack KMS
                                  ↓
                            Uses certificates from Node 1
```

## Prerequisites

1. **Chainlaunch**: Two running instances (Node 1 and Node 2)
   - Node 1: `http://localhost:8100` (creates organizations)
   - Node 2: `http://localhost:8101` (imports organizations)

2. **LocalStack**: Running KMS service (included in docker-compose.yml)
   ```bash
   docker run -d -p 4566:4566 localstack/localstack:latest
   ```
   Or use the provided docker-compose.yml:
   ```bash
   docker-compose up -d
   ```

3. **Terraform**: With AWS provider (Chainlaunch provider uses local build or registry)

4. **Chainlaunch Provider**:
   - For development: Build locally and configure `~/.terraformrc` with dev overrides
   - For production: Terraform will download from registry during `terraform init`

   Example `~/.terraformrc` for local development:
   ```hcl
   provider_installation {
     dev_overrides {
       "kfsoftware/chainlaunch" = "/path/to/chainlaunch-terraform"
     }
     direct {}
   }
   ```

## LocalStack Setup

LocalStack emulates AWS services locally without AWS credentials.

### Start LocalStack

```bash
# Using Docker Compose
cat > docker-compose.yml << 'EOF'
version: '3.8'
services:
  localstack:
    image: localstack/localstack:latest
    ports:
      - "4566:4566"
    environment:
      - SERVICES=kms
      - DEBUG=1
      - DATA_DIR=/tmp/localstack/data
    volumes:
      - "${TMPDIR:-/tmp}/localstack:/tmp/localstack"
    networks:
      - fabric-network

networks:
  fabric-network:
    driver: bridge
EOF

docker-compose up -d
```

### Verify LocalStack KMS

```bash
# Check KMS is running
curl http://localhost:4566/_localstack/health

# Or use AWS CLI with LocalStack endpoint
AWS_ACCESS_KEY_ID=test \
AWS_SECRET_ACCESS_KEY=test \
aws --endpoint-url=http://localhost:4566 \
  kms list-keys --region us-east-1
```

## Certificate Export from Node 1

The source organization certificates need to be exported from Node 1 and provided to this example.

### Option 1: Export from Node 1's Organization

```bash
# Create a temporary Terraform file in Node 1 to export certs
cat > export-certs.tf << 'EOF'
# Get the organization created in Node 1
data "chainlaunch_fabric_organization" "source" {
  id = "1"  # Replace with actual organization ID
}

# Export the certificates
output "sign_cert" {
  value     = data.chainlaunch_fabric_organization.source.sign_certificate
  sensitive = true
}

output "tls_cert" {
  value     = data.chainlaunch_fabric_organization.source.tls_certificate
  sensitive = true
}
EOF

# Run on Node 1
terraform init
terraform apply

# Save to files
terraform output -raw sign_cert > certs/sign-ca.pem
terraform output -raw tls_cert > certs/tls-ca.pem
```

### Option 2: Use fabric_identity to Extract Certificates

```bash
# Create an admin identity from the source organization
resource "chainlaunch_fabric_identity" "source_admin" {
  organization_id = chainlaunch_fabric_organization.source_org.id
  name            = "export-admin"
  role            = "admin"
}

# The certificate will be available in the identity
output "identity_certificate" {
  value     = chainlaunch_fabric_identity.source_admin.certificate
  sensitive = true
}
```

### Option 3: Manual Export

If you have direct access to Node 1's file system:

```bash
# Find the organization's crypto materials
find /path/to/node1/data -name "*sign-ca*" -o -name "*tls-ca*"

# Copy to this example's certs directory
cp /path/to/node1/data/orgs/source-org/sign-ca.pem certs/
cp /path/to/node1/data/orgs/source-org/tls-ca.pem certs/
```

## Development Setup

If you're developing the Chainlaunch provider locally and want to test this example:

### 1. Build the Provider

```bash
cd /path/to/chainlaunch-terraform
go build -o terraform-provider-chainlaunch
```

### 2. Configure Development Overrides

Create or update `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "kfsoftware/chainlaunch" = "/path/to/chainlaunch-terraform"
  }
  direct {}
}
```

Replace `/path/to/chainlaunch-terraform` with the actual path to your provider repository.

### 3. Test the Example

```bash
cd examples/fabric-organization-import-aws-kms-localstack

# Initialize Terraform (required once for lock file)
terraform init

# Now terraform reloads local provider on each command
terraform plan
terraform apply
```

**After the first `terraform init`**:
- The lock file (`.terraform.lock.hcl`) is created with AWS provider version locked
- Chainlaunch provider is loaded from your local binary via dev_overrides
- On each subsequent `terraform plan/apply`, the local provider is reloaded (fast feedback loop)
- You do NOT need to run `terraform init` again unless you change provider versions

## Usage

### 1. Set Up Certificates

Create the `certs/` directory and place certificates from Node 1:

```bash
mkdir -p certs/

# Option A: If you have the certificates exported
cp /path/to/sign-ca.pem certs/
cp /path/to/tls-ca.pem certs/

# Option B: Create test certificates for demonstration
openssl req -new -x509 -days 365 -keyout certs/sign-ca-key.pem \
  -out certs/sign-ca.pem -subj "/CN=test-sign-ca"

openssl req -new -x509 -days 365 -keyout certs/tls-ca-key.pem \
  -out certs/tls-ca.pem -subj "/CN=test-tls-ca"
```

### 2. Configure Node 2

Update variables to point to Node 2 (if different from default):

```bash
# Create terraform.tfvars
cat > terraform.tfvars << 'EOF'
chainlaunch_url      = "http://localhost:8101"  # Node 2
chainlaunch_username = "admin"
chainlaunch_password = "admin123"
aws_region          = "us-east-1"
localstack_endpoint = "http://localhost:4566"
aws_access_key      = "test"
aws_secret_key      = "test"
EOF
```

### 3. Initialize Terraform

#### Option A: Using Local Development Provider (Recommended for Development)

If you have the Chainlaunch provider built locally with dev overrides configured in `~/.terraformrc`:

```bash
# Initialize once to create lock file
# This downloads AWS provider from registry, Chainlaunch uses dev_overrides
terraform init

# Subsequent terraform commands don't need init
terraform plan
terraform apply
```

The dev overrides will automatically use your local provider build instead of downloading from the registry. The lock file ensures provider versions are consistent.

**Important**:
- `terraform init` is still required ONCE to create the lock file for AWS provider
- After that, Terraform reloads the local Chainlaunch provider on each run (fast feedback)
- The `.terraform.lock.hcl` file should NOT be deleted between runs

#### Option B: Standard Registry Provider

For production or CI/CD environments:

```bash
terraform init
terraform plan
terraform apply
```

This downloads both the AWS provider and Chainlaunch provider from the Terraform Registry.

### 4. View the Plan

```bash
terraform plan
```

This will show:
- Source organization creation on Node 1
- KMS keys creation in LocalStack
- AWS KMS provider registration in Chainlaunch
- Organization import on Node 2

### 5. Apply Configuration

```bash
terraform apply
```

### 6. Verify Both Nodes

```bash
# Check Node 1 source organization
terraform output source_organization_id
terraform state show chainlaunch_fabric_organization.source_org

# Check Node 2 imported organization
terraform output imported_organization_id
terraform state show chainlaunch_fabric_organization_import.imported_org

# Check LocalStack KMS keys
AWS_ACCESS_KEY_ID=test \
AWS_SECRET_ACCESS_KEY=test \
aws --endpoint-url=http://localhost:4566 \
  kms list-keys --region us-east-1
```

## Configuration Details

### LocalStack AWS Provider

```hcl
provider "aws" {
  region      = "us-east-1"
  access_key  = "test"
  secret_key  = "test"

  endpoints {
    kms = "http://localhost:4566"
  }

  skip_credentials_validation = true
  skip_requesting_account_id  = true
}
```

Key points:
- **access_key/secret_key**: LocalStack doesn't validate these (use "test")
- **endpoints.kms**: Points to LocalStack KMS service
- **skip_credentials_validation**: Required for LocalStack (no real AWS account)

### AWS KMS Provider Configuration

```hcl
resource "chainlaunch_key_provider" "localstack_kms" {
  name = "localstack-kms-provider"
  type = "AWS_KMS"

  awskms_config = {
    operation   = "IMPORT"
    region      = "us-east-1"
    access_key  = "test"
    secret_key  = "test"
    endpoint    = "http://localhost:4566"
  }
}
```

### Organization Import

```hcl
resource "chainlaunch_fabric_organization_import" "imported_org" {
  msp_id       = "ImportedOrgMSP"
  name         = "imported-org"
  provider_id  = chainlaunch_key_provider.localstack_kms.id
  source_type  = "aws_kms"

  aws_kms_import {
    sign_ca_cert   = file("certs/sign-ca.pem")
    sign_ca_key_id = aws_kms_key.sign_ca_key.arn

    tls_ca_cert   = file("certs/tls-ca.pem")
    tls_ca_key_id = aws_kms_key.tls_ca_key.arn
  }
}
```

## Troubleshooting

### Error: "Connection refused" to LocalStack

LocalStack is not running:

```bash
# Start LocalStack
docker run -d -p 4566:4566 localstack/localstack:latest

# Verify it's running
curl http://localhost:4566/_localstack/health
```

### Error: "KMS key not found"

The key ID/ARN doesn't exist:

```bash
# List available keys in LocalStack
AWS_ACCESS_KEY_ID=test \
AWS_SECRET_ACCESS_KEY=test \
aws --endpoint-url=http://localhost:4566 \
  kms list-keys --region us-east-1
```

### Error: "Certificate file not found"

Certificates aren't in the expected location:

```bash
# Verify certificate files exist
ls -la certs/
cat certs/sign-ca.pem  # Should show -----BEGIN CERTIFICATE-----
```

### Node 2 Cannot Access Certificates from Node 1

You need to export them explicitly:

```bash
# Option 1: Run on Node 1 to get certificate data
cd /path/to/node1/example
terraform output -raw sign_cert > /shared/certs/sign-ca.pem

# Option 2: SSH/copy from Node 1
scp node1:/path/to/certs/sign-ca.pem certs/
scp node1:/path/to/certs/tls-ca.pem certs/
```

## Multi-Node Setup

For a complete two-node setup:

```
Host Machine
├── Node 1 (Chainlaunch) - Port 8100
│   └── Database Provider
├── Node 2 (Chainlaunch) - Port 8101
│   └── LocalStack KMS Provider
└── LocalStack - Port 4566
    └── KMS Service
```

### Docker Compose for Complete Setup

```yaml
version: '3.8'
services:
  # Node 1 - Create organization
  chainlaunch-node1:
    image: chainlaunch:latest
    ports:
      - "8100:8100"
    environment:
      - DATABASE_URL=...
      - PORT=8100
    networks:
      - fabric-network

  # Node 2 - Import organization
  chainlaunch-node2:
    image: chainlaunch:latest
    ports:
      - "8101:8100"
    environment:
      - DATABASE_URL=...
      - PORT=8100
    networks:
      - fabric-network

  # LocalStack - KMS Service
  localstack:
    image: localstack/localstack:latest
    ports:
      - "4566:4566"
    environment:
      - SERVICES=kms
    volumes:
      - "${TMPDIR:-/tmp}/localstack:/tmp/localstack"
    networks:
      - fabric-network

networks:
  fabric-network:
    driver: bridge
```

## Workflow Summary

```
1. Node 1: Create organization
   ↓
2. Node 1: Export certificates
   ↓
3. LocalStack: Create KMS keys
   ↓
4. Node 2: Register AWS KMS provider (LocalStack)
   ↓
5. Node 2: Import organization with KMS-managed keys
```

## Clean Up

To remove all created resources:

```bash
terraform destroy
```

This will:
- Delete the imported organization from Node 2
- Delete the KMS provider from Node 2
- Delete KMS keys from LocalStack
- **Keep** the source organization in Node 1 (managed separately)

To also remove the source organization:

```bash
terraform destroy -target=chainlaunch_fabric_organization.source_org
```

## Real-World Usage

### Disaster Recovery

```bash
# Node 1 fails, need to import to Node 2
# Certificates already backed up
terraform apply -target=chainlaunch_fabric_organization_import.imported_org
```

### Testing

```bash
# Test organization changes before production
# Create in test node, import to another test node
terraform apply -target=chainlaunch_fabric_organization_import.imported_org
```

### Development

```bash
# Develop against imported copy of production org
# Without risking original
terraform apply
```

## References

- [LocalStack Documentation](https://docs.localstack.cloud/)
- [AWS KMS with Terraform](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key)
- [Chainlaunch Organization Import](https://chainlaunch.dev/docs/resources/fabric_organization_import)
- [Fabric Organization Concepts](https://hyperledger-fabric.readthedocs.io/en/latest/msp.html)
