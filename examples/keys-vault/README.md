# HashiCorp Vault Keys Example

This example demonstrates creating various types of cryptographic keys using HashiCorp Vault as the key provider in Chainlaunch.

## Important Limitation

**HashiCorp Vault does NOT support the secp256k1 curve.** If you need secp256k1 keys (for Bitcoin/Ethereum/blockchain applications), use:
- AWS KMS provider (supports secp256k1)
- Database provider (supports secp256k1)

## What This Example Does

- Creates a HashiCorp Vault key provider (connecting to existing Vault instance)
- Creates multiple keys with different algorithms:
  - RSA keys (2048-bit and 4096-bit)
  - EC keys with NIST curves (P-256, P-384, P-521)
  - Certificate Authority (CA) key
- Stores all keys in Vault for enhanced security
- **Does NOT create secp256k1 keys** (Vault limitation)

## Prerequisites

### HashiCorp Vault Setup

1. **Install Vault**:
   ```bash
   # macOS
   brew install vault

   # Linux
   wget https://releases.hashicorp.com/vault/1.15.0/vault_1.15.0_linux_amd64.zip
   unzip vault_1.15.0_linux_amd64.zip
   sudo mv vault /usr/local/bin/
   ```

2. **Start Vault in dev mode** (for testing):
   ```bash
   vault server -dev -dev-root-token-id="root"
   ```

   **Note**: Dev mode is NOT for production. The token is "root" and Vault runs in-memory.

3. **Set environment variables**:
   ```bash
   export VAULT_ADDR='http://127.0.0.1:8200'
   export VAULT_TOKEN='root'
   ```

4. **Verify Vault is running**:
   ```bash
   vault status
   ```

### Production Vault Setup

For production, follow HashiCorp's [production hardening guide](https://developer.hashicorp.com/vault/tutorials/operations/production-hardening):

1. Use persistent storage backend (Consul, Raft, etc.)
2. Enable TLS/HTTPS
3. Use proper authentication (AppRole, Kubernetes, etc.)
4. Configure audit logging
5. Set up seal/unseal procedures
6. Implement high availability

## Key Types Demonstrated

### RSA Keys
- **RSA 2048**: Standard RSA key with 2048-bit key size
- **RSA 4096**: Larger RSA key with 4096-bit key size

### Elliptic Curve (EC) Keys (NIST curves only)
- **P-256**: NIST P-256 curve (secp256r1)
- **P-384**: NIST P-384 curve
- **P-521**: NIST P-521 curve
- **secp256k1**: ❌ **NOT SUPPORTED** by Vault

### Certificate Authority
- **CA RSA**: RSA key marked as Certificate Authority

## Supported vs Unsupported Curves

| Curve       | Vault Support | Use Case                    | Alternative Provider |
|-------------|---------------|-----------------------------|--------------------|
| P-256       | ✅ Yes        | Standard EC operations      | -                  |
| P-384       | ✅ Yes        | High security EC operations | -                  |
| P-521       | ✅ Yes        | Maximum security EC         | -                  |
| secp256k1   | ❌ **NO**     | Bitcoin/Ethereum/Blockchain | AWS KMS, Database  |

## Configuration

The example uses the following default values:

- **Chainlaunch URL**: `http://localhost:8100`
- **Username**: `admin`
- **Password**: `admin123`
- **Vault Address**: `http://127.0.0.1:8200`
- **Vault Token**: `root` (dev mode only)
- **Vault Mount**: `secret`

## Usage

1. **Start Vault in dev mode**:
   ```bash
   vault server -dev -dev-root-token-id="root"
   ```

2. Navigate to this directory:
   ```bash
   cd examples/keys-vault
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
   - 1 key provider to be created (Vault)
   - 6 keys to be created (no secp256k1)

5. Create the provider and keys:
   ```bash
   terraform apply
   ```

6. View all outputs:
   ```bash
   terraform output
   ```

7. View summary (note secp256k1_supported = false):
   ```bash
   terraform output summary
   ```

## Expected Outputs

- `vault_provider`: Vault provider information
- `rsa_2048_key`: RSA 2048-bit key details
- `rsa_4096_key`: RSA 4096-bit key details
- `ec_p256_key`: EC P-256 key details
- `ec_p384_key`: EC P-384 key details
- `ec_p521_key`: EC P-521 key details
- `ca_key`: Certificate Authority key details
- `summary`: Overview with note about secp256k1 limitation

## Resources Created

- 1 HashiCorp Vault key provider
- 6 cryptographic keys stored in Vault (no secp256k1)

## Vault Benefits

### Security Features
- **Encryption at rest**: All data encrypted
- **Dynamic secrets**: Generate credentials on-demand
- **Lease management**: Automatic secret revocation
- **Audit logging**: Complete operation history
- **Access policies**: Fine-grained access control

### Key Management
- **Transit secrets engine**: Encryption as a Service
- **PKI secrets engine**: Certificate authority
- **Key rotation**: Automatic or manual rotation
- **Key versioning**: Multiple versions per key

### Integration
- **Multi-cloud**: Works across AWS, Azure, GCP
- **Kubernetes**: Native K8s integration
- **AppRole**: Machine authentication
- **OIDC/LDAP**: User authentication

## Vault Limitations

### secp256k1 Curve NOT Supported

If you attempt to create a secp256k1 key with Vault, you will get an error:

```bash
Error: curve secp256k1 is not supported by HashiCorp Vault
```

**Workaround**: Use AWS KMS or database provider for secp256k1 keys.

### Why Doesn't Vault Support secp256k1?

Vault focuses on NIST-standardized curves (P-256, P-384, P-521) for enterprise security compliance. The secp256k1 curve, while secure, is primarily used in cryptocurrency applications and is not part of NIST standards.

## When to Use Each Provider

| Provider  | RSA | P-256 | P-384 | P-521 | secp256k1 | Best For |
|-----------|-----|-------|-------|-------|-----------|----------|
| Database  | ✅  | ✅    | ✅    | ✅    | ✅        | Development, testing |
| Vault     | ✅  | ✅    | ✅    | ✅    | ❌        | Enterprise PKI, compliance |
| AWS KMS   | ✅  | ✅    | ✅    | ✅    | ✅        | AWS environments, blockchain |

## Clean Up

To destroy all created resources:

```bash
terraform destroy
```

This will:
1. Delete all keys from Vault
2. Remove the Vault key provider configuration

To stop Vault dev server:
```bash
# Press Ctrl+C in the terminal where Vault is running
```

## Production Considerations

### Don't Use Dev Mode
Dev mode is for testing only:
- ❌ Data is stored in-memory (lost on restart)
- ❌ Vault is unsealed automatically
- ❌ Root token is predictable
- ❌ No TLS encryption

### Production Setup
1. **Use persistent storage**:
   ```hcl
   storage "raft" {
     path = "/vault/data"
   }
   ```

2. **Enable TLS**:
   ```hcl
   listener "tcp" {
     address     = "0.0.0.0:8200"
     tls_cert_file = "/vault/tls/cert.pem"
     tls_key_file  = "/vault/tls/key.pem"
   }
   ```

3. **Use proper authentication**:
   - AppRole for machines
   - OIDC/LDAP for users
   - Kubernetes auth for K8s pods

4. **Enable audit logging**:
   ```bash
   vault audit enable file file_path=/vault/logs/audit.log
   ```

5. **Configure policies**:
   ```hcl
   path "secret/data/chainlaunch/*" {
     capabilities = ["create", "read", "update", "delete", "list"]
   }
   ```

## Troubleshooting

### Issue: "connection refused to 127.0.0.1:8200"
**Solution**: Ensure Vault is running
```bash
vault status
# If not running, start Vault:
vault server -dev -dev-root-token-id="root"
```

### Issue: "permission denied"
**Solution**: Check Vault token has proper permissions
```bash
vault token lookup
# Ensure token has access to secret mount
```

### Issue: "secp256k1 curve not supported"
**Solution**: This is expected. Use AWS KMS or database provider for secp256k1:
```hcl
# Switch to AWS KMS provider
resource "chainlaunch_key_provider" "kms" {
  name = "AWS-KMS"
  type = "AWS_KMS"
  aws_kms_config = { ... }
}
```

### Issue: "mount point not found"
**Solution**: Ensure the secret mount exists
```bash
vault secrets list
# If "secret/" is not listed, enable it:
vault secrets enable -path=secret kv-v2
```

### Issue: "Vault sealed"
**Solution**: Unseal Vault (production only, dev mode auto-unseals)
```bash
vault operator unseal
```

## Next Steps

After creating these keys:
1. Integrate with Hyperledger Fabric organizations
2. Set up key rotation policies
3. Configure audit logging
4. Implement access policies
5. For blockchain keys (secp256k1), switch to AWS KMS or database provider

## Additional Resources

- [Vault Documentation](https://developer.hashicorp.com/vault/docs)
- [Vault PKI Secrets Engine](https://developer.hashicorp.com/vault/docs/secrets/pki)
- [Vault Transit Secrets Engine](https://developer.hashicorp.com/vault/docs/secrets/transit)
- [Production Hardening](https://developer.hashicorp.com/vault/tutorials/operations/production-hardening)
