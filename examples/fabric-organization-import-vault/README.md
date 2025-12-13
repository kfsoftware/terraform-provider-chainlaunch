# Vault Certificate Import Example

This example demonstrates importing Hyperledger Fabric organizations using certificates stored in HashiCorp Vault.

## Use Cases

- Organizations with Vault-managed certificates
- External PKI integration with Vault
- Centralized certificate management
- Secure certificate retrieval via Vault paths

## Prerequisites

1. Chainlaunch API running (default: `http://localhost:8100`)
2. HashiCorp Vault configured and accessible to Chainlaunch
3. Vault key provider configured in Chainlaunch
4. Certificates already stored in Vault at specified paths

## Vault Setup

### Store Certificates in Vault

Before importing, ensure your certificates are stored in Vault:

```bash
# Using Vault KV v2 (default)
vault kv put secret/fabric/org1-sign-ca certificate=@org1-sign-ca.pem key=@org1-sign-ca-key.pem
vault kv put secret/fabric/org1-tls-ca certificate=@org1-tls-ca.pem key=@org1-tls-ca-key.pem

# Or with curl
curl -X POST http://localhost:8200/v1/secret/data/fabric/org1-sign-ca \
  -H "X-Vault-Token: $(vault print token)" \
  -d @request.json
```

### Vault Path Format

For Vault KV v2 (default), paths must include `/data/`:

```
secret/data/path/to/secret  ✅ Correct
secret/path/to/secret       ❌ Incorrect (missing /data/)
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
terraform state show chainlaunch_fabric_organization_import.org1_vault
terraform output org1_vault_id
```

## Configuration

### Basic Vault Import

```hcl
resource "chainlaunch_fabric_organization_import" "org1_vault" {
  msp_id       = "Org1MSP"
  name         = "organization-1"
  provider_id  = 5  # Vault provider ID
  source_type  = "vault"

  vault_import {
    sign_ca_path = "secret/data/fabric/org1-sign-ca"
    tls_ca_path  = "secret/data/fabric/org1-tls-ca"
  }
}
```

### Custom Vault Paths

```hcl
resource "chainlaunch_fabric_organization_import" "org2_custom" {
  msp_id       = "Org2MSP"
  name         = "organization-2"
  provider_id  = 5  # Vault provider ID
  source_type  = "vault"

  vault_import {
    sign_ca_path = "secret/data/organizations/org2/signing-ca"
    tls_ca_path  = "secret/data/organizations/org2/tls-ca"
  }
}
```

### Parameterized Vault Paths

```hcl
variable "vault_base_path" {
  default = "secret/data/fabric"
}

resource "chainlaunch_fabric_organization_import" "org3" {
  msp_id       = "Org3MSP"
  name         = "organization-3"
  provider_id  = 5
  source_type  = "vault"

  vault_import {
    sign_ca_path = "${var.vault_base_path}/org3-sign-ca"
    tls_ca_path  = "${var.vault_base_path}/org3-tls-ca"
  }
}
```

### Bulk Import with for_each

```hcl
locals {
  vault_organizations = {
    org4 = {
      msp_id       = "Org4MSP"
      sign_ca_path = "secret/data/fabric/org4-sign-ca"
      tls_ca_path  = "secret/data/fabric/org4-tls-ca"
    }
    org5 = {
      msp_id       = "Org5MSP"
      sign_ca_path = "secret/data/fabric/org5-sign-ca"
      tls_ca_path  = "secret/data/fabric/org5-tls-ca"
    }
  }
}

resource "chainlaunch_fabric_organization_import" "bulk" {
  for_each = local.vault_organizations

  msp_id       = each.value.msp_id
  provider_id  = 5
  source_type  = "vault"

  vault_import {
    sign_ca_path = each.value.sign_ca_path
    tls_ca_path  = each.value.tls_ca_path
  }
}
```

## Finding the Vault Provider ID

To find your Vault provider ID:

```bash
# Apply to retrieve provider data
terraform apply -target=data.chainlaunch_key_providers.vault

# View the provider ID
terraform state show data.chainlaunch_key_providers.vault
```

## Troubleshooting

### Error: "Vault path not found"

Ensure the Vault path exists and is accessible:

```bash
# Verify path exists in Vault
vault kv get secret/fabric/org1-sign-ca

# Check path format (must include /data/)
✅ secret/data/fabric/org1-sign-ca
❌ secret/fabric/org1-sign-ca
```

### Error: "sign_ca_path is required for vault imports"

The Vault path is missing or empty. Ensure:
- Path is set in the configuration
- Path includes the `/data/` prefix for KV v2
- Path is within quotes

```hcl
vault_import {
  sign_ca_path = "secret/data/fabric/org1-sign-ca"  # ✅ Correct
  tls_ca_path  = "secret/data/fabric/org1-tls-ca"
}
```

### Error: "provider_id points to a Vault key provider"

The provider_id doesn't point to a Vault provider. Verify:
- The ID is from a Vault key provider (not Database or AWS KMS)
- The Vault provider is configured in Chainlaunch
- The ID is correct

### Vault Connection Issues

If Chainlaunch cannot reach Vault:

```bash
# Check Vault is running
curl http://localhost:8200/v1/sys/health

# Check Vault token/auth in Chainlaunch logs
docker logs chainlaunch

# Verify network connectivity
ping vault-host
```

## Post-Import Operations

Once imported, use the organization in other resources:

```hcl
# Create identities
resource "chainlaunch_fabric_identity" "org1_admin" {
  organization_id = chainlaunch_fabric_organization_import.org1_vault.id
  name            = "admin"
  role            = "admin"
}

# Add to Fabric network
resource "chainlaunch_fabric_network" "mychannel" {
  name = "mychannel"
  peer_organizations = [
    {
      id = chainlaunch_fabric_organization_import.org1_vault.id
    }
  ]
}
```

## Vault Secret Structure

Vault secrets should contain certificate and key data:

```bash
# Example: Create Vault secret
vault kv put secret/fabric/org1-sign-ca \
  certificate=@org1-sign-ca.pem \
  key=@org1-sign-ca-key.pem

# View stored secret
vault kv get secret/fabric/org1-sign-ca
```

## Clean Up

To remove all imported organizations:

```bash
terraform destroy
```

Note: This only removes Terraform state and the organizations from Chainlaunch. Vault secrets are not deleted.

## References

- [Fabric Organization Import Resource Documentation](https://registry.terraform.io/providers/kfsoftware/chainlaunch/latest/docs/resources/fabric_organization_import)
- [HashiCorp Vault Documentation](https://www.vaultproject.io/docs)
- [Hyperledger Fabric MSP Concepts](https://hyperledger-fabric.readthedocs.io/en/latest/msp.html)
