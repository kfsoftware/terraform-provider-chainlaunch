# HashiCorp Vault Key Provider Example

This example demonstrates how to create HashiCorp Vault key providers in Chainlaunch for managing cryptographic keys and certificates in Hyperledger Fabric organizations.

## What This Example Does

- Creates a Vault provider in IMPORT mode (connecting to an existing Vault instance)
- Creates a Vault provider in CREATE mode (Chainlaunch manages the Vault instance)
- Creates a Fabric organization that uses Vault for key management
- Demonstrates comprehensive Vault configuration options

## Vault Operation Modes

### IMPORT Mode
Connect to an existing Vault instance that you manage:
- **Use case**: You already have Vault running and want to use it
- **Required fields**: `address`, `token`
- **Optional fields**: `mount`, `ca_cert`, `namespace`

### CREATE Mode
Chainlaunch deploys and manages a Vault instance for you:
- **Use case**: You want Chainlaunch to handle Vault deployment
- **Required fields**: `mode` (docker or service)
- **Optional fields**: `network`, `port`, PKI configuration

## Configuration Examples

### Example 1: IMPORT Mode (Existing Vault)

```hcl
resource "chainlaunch_key_provider" "vault_existing" {
  name       = "ExistingVaultProvider"
  type       = "VAULT"
  is_default = false

  vault_config = {
    operation = "IMPORT"
    address   = "http://127.0.0.1:8200"  # Vault server URL
    token     = "root"                    # Vault token
    mount     = "secret"                  # KV mount path (default: "secret")
  }
}
```

### Example 2: CREATE Mode (Managed Vault)

```hcl
resource "chainlaunch_key_provider" "vault_managed" {
  name       = "ManagedVaultProvider"
  type       = "VAULT"
  is_default = false

  vault_config = {
    operation = "CREATE"
    mode      = "docker"   # "docker" or "service"
    network   = "bridge"   # "host" or "bridge"
    port      = 8200       # Vault server port
  }
}
```

### Example 3: Advanced Configuration with PKI Settings

```hcl
resource "chainlaunch_key_provider" "vault_advanced" {
  name       = "AdvancedVaultProvider"
  type       = "VAULT"
  is_default = false

  vault_config = {
    operation = "IMPORT"
    address   = "https://vault.example.com:8200"
    token     = var.vault_token

    # Mount paths
    kv_mount  = "secret"
    pki_mount = "pki"

    # PKI certificate TTLs
    default_cert_ttl    = "8760h"   # 1 year
    max_cert_ttl        = "87600h"  # 10 years
    default_ca_cert_ttl = "87600h"  # 10 years
    max_ca_cert_ttl     = "175200h" # 20 years

    # Optional: TLS configuration
    ca_cert   = file("path/to/ca.crt")
    namespace = "admin"
  }
}
```

## Available Configuration Options

### Common Fields (Both Modes)
- **operation** (required): "IMPORT" or "CREATE"

### IMPORT Mode Fields
- **address** (required): Vault server URL (e.g., "http://127.0.0.1:8200")
- **token** (required, sensitive): Vault authentication token
- **mount** (optional): KV secrets mount path (default: "secret")
- **ca_cert** (optional): CA certificate for TLS verification
- **namespace** (optional): Vault namespace (Vault Enterprise feature)

### CREATE Mode Fields
- **mode** (required): Deployment mode - "docker" or "service"
- **network** (optional): Network mode - "host" or "bridge" (default: "bridge")
- **port** (optional): Vault server port (default: 8200)

### PKI Configuration (Both Modes)
- **pki_mount** (optional): PKI engine mount path (default: "pki")
- **kv_mount** (optional): KV secrets mount path (default: "secret")
- **default_cert_ttl** (optional): Default certificate TTL (e.g., "8760h")
- **max_cert_ttl** (optional): Maximum certificate TTL (e.g., "87600h")
- **default_ca_cert_ttl** (optional): Default CA certificate TTL
- **max_ca_cert_ttl** (optional): Maximum CA certificate TTL

## Prerequisites

### For IMPORT Mode
1. **Vault Server Running**: Have Vault accessible at the specified address
   ```bash
   # Start Vault in dev mode for testing
   vault server -dev -dev-root-token-id="root"
   ```

2. **Vault Token**: Have a valid Vault token with appropriate permissions

### For CREATE Mode
1. **Docker** (if mode = "docker"): Docker must be available on the Chainlaunch host
2. **System Service Support** (if mode = "service"): systemd or equivalent service manager

## Usage

1. **Navigate to this directory**:
   ```bash
   cd examples/vault-provider
   ```

2. **Review and customize the configuration**:
   - Update Vault address, token, and other settings in `main.tf`
   - Consider using variables for sensitive values

3. **Plan the deployment**:
   ```bash
   terraform plan
   ```

4. **Apply the configuration**:
   ```bash
   terraform apply
   ```

5. **View the outputs**:
   ```bash
   terraform output
   ```

## Security Best Practices

### 1. Never Hardcode Sensitive Values

**Bad**:
```hcl
token = "s.1234567890abcdef"  # Never do this!
```

**Good**:
```hcl
variable "vault_token" {
  type      = string
  sensitive = true
}

vault_config = {
  token = var.vault_token
}
```

### 2. Use Environment Variables

```bash
export TF_VAR_vault_token="s.1234567890abcdef"
terraform apply
```

### 3. Use Vault for Terraform State

Store Terraform state in Vault or use encrypted backend:
```hcl
terraform {
  backend "consul" {
    address = "consul.example.com:8500"
    scheme  = "https"
    path    = "terraform/chainlaunch"
  }
}
```

### 4. Rotate Tokens Regularly

For IMPORT mode, use short-lived tokens and rotate them regularly using Vault's token renewal mechanism.

### 5. Use TLS in Production

Always use HTTPS for Vault in production:
```hcl
address = "https://vault.example.com:8200"
ca_cert = file("${path.module}/ca.crt")
```

## Expected Outputs

After successful apply:

```
vault_existing_provider_id   = "6"
vault_existing_provider_type = "VAULT"
vault_managed_provider_id    = "7"
organization_id              = "16"
organization_provider_id     = 6
```

## Troubleshooting

### Error: "dial tcp 127.0.0.1:8200: connect: connection refused"

**Cause**: Vault server is not running or not accessible

**Solutions**:
1. Start Vault server:
   ```bash
   vault server -dev -dev-root-token-id="root"
   ```
2. Check Vault address in configuration
3. Ensure firewall allows connection

### Error: "mode must be either 'docker' or 'service'"

**Cause**: Invalid mode value in CREATE operation

**Solution**: Use "docker" or "service" as the mode value:
```hcl
mode = "docker"  # Not "dev" or "prod"
```

### Error: "permission denied"

**Cause**: Vault token doesn't have required permissions

**Solution**: Ensure token has permissions for:
- PKI engine operations
- KV secrets operations
- Mount management

### Provider Created But Organization Failed

This is expected if Vault isn't properly configured or accessible. The key provider resource creation itself succeeded! Verify:

```bash
export CHAINLAUNCH_USER=""
export CHAINLAUNCH_PASSWORD=""
curl -s http://localhost:8100/api/v1/key-providers/[ID] -u $CHAINLAUNCH_USER:$CHAINLAUNCH_PASSWORD
```

## TTL Format Reference

Time-To-Live (TTL) values use Go duration format:
- `h` = hours (e.g., "24h" = 24 hours)
- `m` = minutes (e.g., "30m" = 30 minutes)
- `s` = seconds (e.g., "60s" = 60 seconds)

Common examples:
- 1 hour: "1h"
- 1 day: "24h"
- 1 month: "720h"
- 1 year: "8760h"
- 10 years: "87600h"
- 20 years: "175200h"

## Clean Up

To destroy the resources:

```bash
terraform destroy
```

**Note**: If Chainlaunch created a Vault instance (CREATE mode), it will be stopped and removed.

## Additional Resources

- [HashiCorp Vault Documentation](https://www.vaultproject.io/docs)
- [Vault PKI Secrets Engine](https://www.vaultproject.io/docs/secrets/pki)
- [Vault KV Secrets Engine](https://www.vaultproject.io/docs/secrets/kv)
- [Hyperledger Fabric MSP](https://hyperledger-fabric.readthedocs.io/en/latest/msp.html)
- [Chainlaunch API Documentation](http://localhost:8100/api-docs)

## Next Steps

After creating the Vault provider:
1. Create organizations using the Vault provider
2. Manage certificates through Vault's PKI engine
3. Store organization secrets in Vault's KV engine
4. Monitor Vault audit logs for key operations
5. Implement Vault policies for fine-grained access control
