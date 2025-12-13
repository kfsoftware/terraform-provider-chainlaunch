# Raw Certificate Import Example

This example demonstrates importing Hyperledger Fabric organizations using PEM-encoded certificates directly.

## Use Cases

- Importing existing organizations from external Fabric networks
- Using certificates managed outside of Chainlaunch
- Organizations where private keys may or may not be available
- Testing and development scenarios

## Features

- Import with full certificate chain (including private keys)
- Import with certificates only (no private keys)
- Create identities after import
- Simple and straightforward approach

## Prerequisites

1. Chainlaunch API running (default: `http://localhost:8100`)
2. Default credentials configured (admin/admin123)
3. PEM-encoded certificates available

## Certificate Structure

Place your PEM-encoded certificates in a `certs/` directory:

```bash
mkdir -p certs/
# Copy or create your certificates
# - org1-sign-ca.pem (signing CA certificate)
# - org1-sign-ca-key.pem (signing CA private key - optional)
# - org1-tls-ca.pem (TLS CA certificate)
# - org1-tls-ca-key.pem (TLS CA private key - optional)
# - org2-sign-ca.pem
# - org2-tls-ca.pem
```

### Certificate Format

Certificates must be in PEM format:

```
-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIRAKZw...
-----END CERTIFICATE-----
```

Private keys must be in PEM format:

```
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQ...
-----END PRIVATE KEY-----
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
terraform state show chainlaunch_fabric_organization_import.org1_with_keys
terraform output org1_id
```

## Configuration

### With Private Keys (Allows Identity Generation)

```hcl
resource "chainlaunch_fabric_organization_import" "org1_with_keys" {
  msp_id       = "Org1MSP"
  name         = "organization-1"
  provider_id  = 1  # Database provider
  source_type  = "raw"

  raw_import {
    sign_ca_cert        = file("${path.module}/certs/org1-sign-ca.pem")
    sign_ca_private_key = file("${path.module}/certs/org1-sign-ca-key.pem")
    tls_ca_cert         = file("${path.module}/certs/org1-tls-ca.pem")
    tls_ca_private_key  = file("${path.module}/certs/org1-tls-ca-key.pem")
  }
}
```

### Without Private Keys

```hcl
resource "chainlaunch_fabric_organization_import" "org2_certs_only" {
  msp_id       = "Org2MSP"
  name         = "organization-2"
  provider_id  = 1  # Database provider
  source_type  = "raw"

  raw_import {
    sign_ca_cert = file("${path.module}/certs/org2-sign-ca.pem")
    tls_ca_cert  = file("${path.module}/certs/org2-tls-ca.pem")
  }
}
```

## Creating Identities After Import

After importing an organization, you can create admin or client identities:

```hcl
resource "chainlaunch_fabric_identity" "org1_admin" {
  organization_id = chainlaunch_fabric_organization_import.org1_with_keys.id
  name            = "admin"
  role            = "admin"
  description     = "Admin identity for Organization 1"
}
```

## Provider ID

The `provider_id` must point to a key provider. For raw imports, use:

- **Database Provider** (default): Most common choice for raw imports
- **Vault Provider**: Also works, but typically used for Vault-managed certificates
- **AWS KMS Provider**: Also works, but typically used for KMS-managed keys

To find your database provider ID:

```bash
terraform apply -target=data.chainlaunch_key_providers.database
terraform state show data.chainlaunch_key_providers.database
```

## Troubleshooting

### Error: "sign_ca_cert is required for raw imports"

The signing CA certificate is missing. Ensure:
- The file path is correct
- The file contains a valid PEM certificate
- The file is readable

```bash
ls -la certs/org1-sign-ca.pem
cat certs/org1-sign-ca.pem  # Should show BEGIN CERTIFICATE
```

### Error: "provider_id is required and must match the certificate source"

The `provider_id` is missing or invalid. Get it from:

```bash
terraform apply -target=data.chainlaunch_key_providers.database
```

### Error: "source_type must be 'raw', 'vault', or 'aws_kms'"

The `source_type` must be explicitly set to `"raw"` for raw certificate imports.

## Post-Import Operations

Once imported, use the organization in other resources:

```hcl
# Add to a fabric network
resource "chainlaunch_fabric_network" "mychannel" {
  name       = "mychannel"

  peer_organizations = [
    {
      id = chainlaunch_fabric_organization_import.org1_with_keys.id
      node_ids = [...]
    }
  ]
}

# Create additional identities
resource "chainlaunch_fabric_identity" "org1_client" {
  organization_id = chainlaunch_fabric_organization_import.org1_with_keys.id
  name            = "client"
  role            = "client"
}
```

## Clean Up

To remove all imported organizations and identities:

```bash
terraform destroy
```

## References

- [Fabric Organization Import Resource Documentation](https://registry.terraform.io/providers/kfsoftware/chainlaunch/latest/docs/resources/fabric_organization_import)
- [Hyperledger Fabric MSP Concepts](https://hyperledger-fabric.readthedocs.io/en/latest/msp.html)
- [Certificate Generation Guide](https://hyperledger-fabric.readthedocs.io/en/latest/commands/cryptogen.html)
