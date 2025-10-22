# Database Keys Example

This example demonstrates creating various types of cryptographic keys using the default database key provider in Chainlaunch.

## What This Example Does

- Fetches the default key provider (Database)
- Creates multiple keys with different algorithms and configurations:
  - RSA keys (2048-bit and 4096-bit)
  - EC keys with NIST curves (P-256, P-384, P-521)
  - EC key with secp256k1 curve (Bitcoin/Ethereum)
  - ED25519 key
  - Certificate Authority (CA) key
- Outputs details of all created keys

## Key Types Demonstrated

### RSA Keys
- **RSA 2048**: Standard RSA key with 2048-bit key size (common for general use)
- **RSA 4096**: Larger RSA key with 4096-bit key size (higher security, slower operations)

### Elliptic Curve (EC) Keys
- **P-256**: NIST P-256 curve (also known as secp256r1 or prime256v1)
- **P-384**: NIST P-384 curve (higher security than P-256)
- **P-521**: NIST P-521 curve (highest security NIST curve)
- **secp256k1**: Bitcoin/Ethereum curve (used in blockchain applications)

### ED25519 Keys
- **ED25519**: Modern EdDSA signature algorithm (fast, secure, small keys)

### Certificate Authority
- **CA RSA**: RSA key marked as Certificate Authority (can sign other certificates)

## Configuration

The example uses the following default values:

- **Chainlaunch URL**: `http://localhost:8100`
- **Username**: `admin`
- **Password**: `admin123`

## Usage

1. Navigate to this directory:
   ```bash
   cd examples/keys-database
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Review the planned changes:
   ```bash
   terraform plan
   ```

   You should see:
   - 1 data source to be read (key providers)
   - 8 keys to be created

4. Create the keys:
   ```bash
   terraform apply
   ```

5. View all outputs:
   ```bash
   terraform output
   ```

6. View specific key details:
   ```bash
   terraform output rsa_2048_key
   terraform output ec_p256_key
   terraform output ed25519_key
   ```

## Expected Outputs

- `default_provider_details`: Default provider information (ID, name, type)
- `rsa_2048_key`: RSA 2048-bit key details
- `rsa_4096_key`: RSA 4096-bit key details
- `ec_p256_key`: EC P-256 key details
- `ec_p384_key`: EC P-384 key details
- `ec_p521_key`: EC P-521 key details
- `ec_secp256k1_key`: EC secp256k1 key details
- `ed25519_key`: ED25519 key details
- `ca_key`: Certificate Authority key details

## Resources Created

- 8 cryptographic keys stored in the database provider

## Key Algorithm Selection Guide

### When to use RSA:
- Traditional PKI infrastructure
- Wide compatibility required
- Certificate signing operations
- Key sizes: 2048 (standard), 4096 (high security)

### When to use EC:
- Modern applications requiring efficiency
- Mobile/embedded devices (smaller key sizes)
- Blockchain applications (secp256k1)
- Curves: P-256 (standard), P-384/P-521 (high security), secp256k1 (blockchain)

### When to use ED25519:
- Modern applications
- Performance-critical operations
- Small signature/key size requirements
- Resistance to side-channel attacks

### When to use CA keys:
- Creating certificate hierarchies
- Signing other certificates
- Building PKI infrastructure

## Clean Up

To destroy all created keys:

```bash
terraform destroy
```

## Notes

- All keys are stored in the database provider (default)
- The `is_ca` flag marks keys as Certificate Authority keys
- Key IDs are returned after creation and can be used in other resources
- Public keys are available in the state file but not shown in outputs for brevity
- Created timestamps are in ISO 8601 format

## Next Steps

After creating these keys, you can:
1. Use them to create certificates
2. Assign them to organizations
3. Use them for signing operations
4. Export public keys for distribution

## Troubleshooting

### Issue: "algorithm not supported"
**Solution**: Check that you're using a valid algorithm: RSA, EC, or ED25519

### Issue: "curve not supported for algorithm"
**Solution**: Curves only apply to EC algorithm. Remove curve for RSA and ED25519

### Issue: "key_size required for RSA"
**Solution**: RSA keys require key_size parameter (2048 or 4096)

### Issue: "provider not found"
**Solution**: Ensure the Chainlaunch instance is running and accessible
