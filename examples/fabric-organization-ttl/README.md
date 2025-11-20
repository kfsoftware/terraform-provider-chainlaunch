# Fabric Organization Certificate TTL Example

This example demonstrates configuring certificate validity periods (TTL) for Hyperledger Fabric organizations using the Chainlaunch Terraform provider.

## Overview

The `chainlaunch_fabric_organization` resource supports two optional TTL (Time To Live) configuration fields:

- **`ca_cert_valid_for`**: Certificate validity configuration for CA certificates (e.g., "87600h" for 10 years)
- **`cert_valid_for`**: Certificate validity for admin/client/peer certificates (e.g., "8760h" for 1 year, default)

Both fields accept Go duration format strings (e.g., "8760h", "720h").

## Use Cases

### Development/Testing
Use short certificate lifetimes (e.g., 30 days) for development environments:
```hcl
resource "chainlaunch_fabric_organization" "dev_org" {
  msp_id              = "DevOrgMSP"
  cert_valid_for      = "720h"  # 30 days
}
```

### Production
Use longer certificate lifetimes for production stability:
```hcl
resource "chainlaunch_fabric_organization" "prod_org" {
  msp_id              = "ProdOrgMSP"
  ca_cert_valid_for   = "175200h"  # 20 years
  cert_valid_for      = "8760h"    # 1 year
}
```

### High-Security Environments
Use short-lived user certificates with long-lived CA certificates:
```hcl
resource "chainlaunch_fabric_organization" "secure_org" {
  msp_id              = "SecureOrgMSP"
  ca_cert_valid_for   = "87600h"   # 10 years (rarely changes)
  cert_valid_for      = "2160h"    # 90 days (frequently rotated)
}
```

## Duration Format Reference

Go duration format examples:
- `1h` - 1 hour
- `24h` - 1 day
- `720h` - 30 days
- `8760h` - 1 year (365 days × 24 hours)
- `17520h` - 2 years
- `87600h` - 10 years
- `175200h` - 20 years

## Prerequisites

1. Chainlaunch API running (default: `http://localhost:8100`)
2. Default credentials configured (admin/admin123)
3. At least one key provider available (Database provider is used by default)

## Development Setup

If you're developing the Chainlaunch provider locally:

1. Build the provider:
   ```bash
   cd /path/to/chainlaunch-terraform
   go build -o terraform-provider-chainlaunch
   ```

2. Configure `~/.terraformrc` with dev overrides:
   ```hcl
   provider_installation {
     dev_overrides {
       "kfsoftware/chainlaunch" = "/absolute/path/to/chainlaunch-terraform"
     }
     direct {}
   }
   ```

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

For production, use `terraform init` without dev overrides to download the provider from the registry.

## Usage

### View the Plan

```bash
terraform plan
```

### Apply the Configuration

```bash
terraform apply
```

### Verify Organizations

```bash
# View created organizations
terraform state show chainlaunch_fabric_organization.org1_default
terraform state show chainlaunch_fabric_organization.org4_custom_ttl

# View outputs
terraform output org4_ca_ttl
terraform output org4_cert_ttl
```

## Configuration Examples

### Basic Organization with Defaults

```hcl
resource "chainlaunch_fabric_organization" "basic" {
  msp_id      = "BasicOrgMSP"
  description = "Organization with default TTL settings"
  provider_id = data.chainlaunch_key_providers.database.default_provider_id
}
```

When no TTL fields are specified, the Chainlaunch API uses its default values.

### Organization with CA Certificate TTL

```hcl
resource "chainlaunch_fabric_organization" "long_ca" {
  msp_id            = "LongCAOrgMSP"
  description       = "Organization with long-lived CA certificates"
  provider_id       = data.chainlaunch_key_providers.database.default_provider_id
  ca_cert_valid_for = "87600h"  # 10 years
}
```

### Organization with Custom Certificate TTL

```hcl
resource "chainlaunch_fabric_organization" "custom_cert" {
  msp_id          = "CustomCertOrgMSP"
  description     = "Organization with custom certificate validity"
  provider_id     = data.chainlaunch_key_providers.database.default_provider_id
  cert_valid_for  = "4380h"     # 6 months
}
```

### Organization with Both CA and Certificate TTL

```hcl
resource "chainlaunch_fabric_organization" "full_config" {
  msp_id              = "FullConfigOrgMSP"
  description         = "Organization with full TTL configuration"
  provider_id         = data.chainlaunch_key_providers.database.default_provider_id
  ca_cert_valid_for   = "175200h"  # 20 years (root CA validity)
  cert_valid_for      = "8760h"    # 1 year (issued certificates)
}
```

### Dynamic TTL Based on Environment

```hcl
variable "environment" {
  type = string
  default = "development"
}

variable "cert_ttl" {
  type = map(string)
  default = {
    development = "720h"    # 30 days
    staging     = "2160h"   # 90 days
    production  = "8760h"   # 1 year
  }
}

resource "chainlaunch_fabric_organization" "env_org" {
  msp_id         = "EnvOrgMSP"
  description    = "Organization for ${var.environment} environment"
  provider_id    = data.chainlaunch_key_providers.database.default_provider_id
  cert_valid_for = var.cert_ttl[var.environment]
}
```

## Key Provider Selection

The `provider_id` field specifies which key provider to use:

### Database Provider (Default)
Most common choice for testing and development:
```hcl
data "chainlaunch_key_providers" "database" {
  type_filter = "DATABASE"
}

resource "chainlaunch_fabric_organization" "org" {
  provider_id = data.chainlaunch_key_providers.database.default_provider_id
}
```

### AWS KMS Provider
For production deployments with AWS KMS:
```hcl
data "chainlaunch_key_providers" "kms" {
  type_filter = "AWS_KMS"
}

resource "chainlaunch_fabric_organization" "org" {
  provider_id = data.chainlaunch_key_providers.kms.default_provider_id
}
```

### HashiCorp Vault Provider
For deployments using HashiCorp Vault:
```hcl
data "chainlaunch_key_providers" "vault" {
  type_filter = "VAULT"
}

resource "chainlaunch_fabric_organization" "org" {
  provider_id = data.chainlaunch_key_providers.vault.default_provider_id
}
```

## Best Practices

1. **CA Certificates**: Use longer TTLs (5-20 years) to minimize CA rotation
2. **User Certificates**: Use shorter TTLs (90 days to 1 year) for security
3. **Development**: Use very short TTLs (30 days) for rapid testing
4. **Production**: Use conservative TTLs that match your certificate rotation policy
5. **Consistency**: Apply the same TTL configuration across organizations in the same environment

## Troubleshooting

### Invalid Duration Format

If you get an error like "invalid duration format", ensure:
- Format follows Go duration syntax (e.g., "8760h", not "1y")
- Use only hours (h), minutes (m), seconds (s) units
- No spaces in the duration string

```bash
# Valid formats
"1h"      # 1 hour
"24h"     # 1 day
"8760h"   # 1 year
"720h"    # 30 days

# Invalid formats
"1 year"      # ❌ Use "8760h" instead
"365d"        # ❌ Use "8760h" instead
"1y"          # ❌ Use "8760h" instead
```

### TTL Not Applied

If TTL settings don't appear in the organization:
1. Verify the API supports TTL parameters (requires recent Chainlaunch version)
2. Check that the key provider supports custom TTL configuration
3. Review the Chainlaunch server logs for validation errors

```bash
terraform apply -target=chainlaunch_fabric_organization.org1_default
```

## Post-Organization Operations

Once organizations are created with TTL settings, you can:

### Create Identities
```hcl
resource "chainlaunch_fabric_identity" "admin" {
  organization_id = chainlaunch_fabric_organization.org1_default.id
  name            = "admin"
  role            = "admin"
}
```

### Add to Networks
```hcl
resource "chainlaunch_fabric_network" "channel" {
  name = "mychannel"

  peer_organizations = [
    {
      id = chainlaunch_fabric_organization.org1_default.id
    }
  ]
}
```

### Import into Other Nodes
```hcl
resource "chainlaunch_fabric_organization_import" "imported" {
  msp_id      = "ImportedOrgMSP"
  provider_id = 2
  source_type = "raw"

  raw_import {
    sign_ca_cert = data.chainlaunch_fabric_organization.org1_default.sign_certificate
    tls_ca_cert  = data.chainlaunch_fabric_organization.org1_default.tls_certificate
  }
}
```

## Clean Up

To remove all organizations:

```bash
terraform destroy
```

## References

- [Fabric Organization Resource Documentation](https://registry.terraform.io/providers/kfsoftware/chainlaunch/latest/docs/resources/fabric_organization)
- [Go Duration Format](https://golang.org/pkg/time/#ParseDuration)
- [Hyperledger Fabric MSP Concepts](https://hyperledger-fabric.readthedocs.io/en/latest/msp.html)
- [Certificate Management Best Practices](https://hyperledger-fabric.readthedocs.io/en/latest/getting_started.html)
