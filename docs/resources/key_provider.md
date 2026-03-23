---
page_title: "chainlaunch_key_provider Resource - chainlaunch"
subcategory: "Key Management"
description: |-
  Manages a key provider in Chainlaunch for managing cryptographic keys. Supports Database, AWS KMS, and HashiCorp Vault backends.
---

# chainlaunch_key_provider (Resource)

Manages a key provider in Chainlaunch for managing cryptographic keys. Key providers are the backends that store and operate on cryptographic keys used by blockchain nodes.

Supported provider types:

| Type | Algorithms | Notes |
|------|-----------|-------|
| `DATABASE` | RSA, EC, secp256k1, ED25519 | Built-in, no external dependencies |
| `AWS_KMS` | RSA, EC, secp256k1 | Requires AWS credentials or IAM role |
| `VAULT` | RSA, EC | HashiCorp Vault, IMPORT or CREATE mode |

## Example Usage

### Database Provider (simplest)

```terraform
resource "chainlaunch_key_provider" "db" {
  name       = "default-db"
  type       = "DATABASE"
  is_default = true
}
```

### AWS KMS with IAM Instance Role / EKS IRSA (recommended)

When ChainLaunch runs on EC2 or EKS, omit credentials entirely. The AWS SDK discovers them automatically from the instance metadata service or IRSA-projected token.

```terraform
resource "chainlaunch_key_provider" "aws_kms" {
  name       = "aws-kms-production"
  type       = "AWS_KMS"
  is_default = false

  aws_kms_config = {
    operation            = "IMPORT"
    aws_region           = "us-east-1"
    kms_key_alias_prefix = "chainlaunch/"
  }
}
```

Required IAM policy for the instance role or IRSA service account:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kms:CreateKey",
        "kms:CreateAlias",
        "kms:DeleteAlias",
        "kms:DescribeKey",
        "kms:GetPublicKey",
        "kms:ListAliases",
        "kms:ListKeys",
        "kms:Sign",
        "kms:Verify",
        "kms:TagResource",
        "kms:ScheduleKeyDeletion"
      ],
      "Resource": "*"
    }
  ]
}
```

### AWS KMS with STS AssumeRole (cross-account)

ChainLaunch assumes a dedicated IAM role via STS. Use this for cross-account access or to scope permissions to a specific role.

```terraform
resource "chainlaunch_key_provider" "aws_kms_cross_account" {
  name       = "aws-kms-cross-account"
  type       = "AWS_KMS"
  is_default = false

  aws_kms_config = {
    operation            = "IMPORT"
    aws_region           = "us-east-1"
    assume_role_arn      = "arn:aws:iam::123456789012:role/ChainLaunchKMSRole"
    external_id          = "chainlaunch-unique-id"
    kms_key_alias_prefix = "chainlaunch/"
  }
}
```

The target role's trust policy must allow the ChainLaunch principal:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::111111111111:role/ChainLaunchInstanceRole"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "chainlaunch-unique-id"
        }
      }
    }
  ]
}
```

### AWS KMS with Static Credentials (development only)

For local development with LocalStack or when ChainLaunch runs outside AWS.

```terraform
resource "chainlaunch_key_provider" "aws_kms_localstack" {
  name       = "aws-kms-localstack"
  type       = "AWS_KMS"
  is_default = false

  aws_kms_config = {
    operation             = "IMPORT"
    aws_region            = "us-east-1"
    aws_access_key_id     = "test"
    aws_secret_access_key = "test"
    endpoint_url          = "http://localhost:4566"
    kms_key_alias_prefix  = "chainlaunch/"
  }
}
```

### AWS KMS with Static Credentials + AssumeRole

When ChainLaunch runs outside AWS (on-prem, other cloud), provide base credentials that are allowed to assume the target role.

```terraform
resource "chainlaunch_key_provider" "aws_kms_external" {
  name       = "aws-kms-external"
  type       = "AWS_KMS"
  is_default = false

  aws_kms_config = {
    operation             = "IMPORT"
    aws_region            = "us-east-1"
    aws_access_key_id     = var.aws_access_key_id
    aws_secret_access_key = var.aws_secret_access_key
    assume_role_arn       = "arn:aws:iam::123456789012:role/ChainLaunchKMSRole"
    external_id           = "chainlaunch-unique-id"
    kms_key_alias_prefix  = "chainlaunch/"
  }
}
```

### HashiCorp Vault (IMPORT existing)

```terraform
resource "chainlaunch_key_provider" "vault" {
  name = "vault-existing"
  type = "VAULT"

  vault_config = {
    operation = "IMPORT"
    address   = "https://vault.example.com:8200"
    token     = var.vault_token
    namespace = "admin"
  }
}
```

### HashiCorp Vault (CREATE managed instance)

```terraform
resource "chainlaunch_key_provider" "vault_managed" {
  name = "vault-managed"
  type = "VAULT"

  vault_config = {
    operation = "CREATE"
    mode      = "docker"
    network   = "bridge"
    port      = 8200
    version   = "1.15.6"
  }
}
```

## Schema

### Required

- `name` (String) The name of the key provider.
- `type` (String) The type of key provider: `AWS_KMS`, `VAULT`, `DATABASE`, or `HSM`.

### Optional

- `aws_kms_config` (Attributes) AWS KMS configuration. Required when type is `AWS_KMS`. Supports three authentication modes: (1) IAM Instance Role / EKS IRSA — omit credentials; (2) STS AssumeRole — set `assume_role_arn`; (3) Static credentials — set `aws_access_key_id` and `aws_secret_access_key`. See [below for nested schema](#nestedatt--aws_kms_config).
- `is_default` (Boolean) Whether this is the default key provider.
- `vault_config` (Attributes) Vault configuration. Required when type is `VAULT`. See [below for nested schema](#nestedatt--vault_config).

### Read-Only

- `created_at` (String) The timestamp when the key provider was created.
- `id` (String) The unique identifier of the key provider.

<a id="nestedatt--aws_kms_config"></a>
### Nested Schema for `aws_kms_config`

Required:

- `aws_region` (String) AWS region where KMS keys are located (e.g., `us-east-1`).
- `operation` (String) Operation mode: `IMPORT` (use existing KMS keys) or `CREATE` (create new KMS keys).

Optional:

- `assume_role_arn` (String) IAM role ARN to assume via STS AssumeRole. Use for cross-account access or to scope permissions to a dedicated KMS role. Works with both instance role and static credential base authentication.
- `aws_access_key_id` (String, Sensitive) AWS access key ID for static credential authentication. Omit when running on EC2 (instance role) or EKS (IRSA) — the AWS SDK picks up credentials automatically.
- `aws_secret_access_key` (String, Sensitive) AWS secret access key for static credential authentication. Required when `aws_access_key_id` is set. Omit for IAM role-based auth.
- `aws_session_token` (String, Sensitive) AWS session token for temporary credentials (e.g., from STS GetSessionToken). Usually not needed.
- `endpoint_url` (String) Custom KMS endpoint URL. Use for LocalStack (`http://localhost:4566`) or VPC endpoints.
- `external_id` (String) External ID for STS AssumeRole, providing an additional layer of security. Only used when `assume_role_arn` is set.
- `kms_key_alias_prefix` (String) Prefix for KMS key aliases (default: `chainlaunch/`).

#### AWS KMS Authentication Modes

| Mode | Fields to Set | Use When |
|------|--------------|----------|
| Instance Role / IRSA | `aws_region` only | ChainLaunch on EC2 or EKS |
| STS AssumeRole | `aws_region` + `assume_role_arn` | Cross-account or dedicated role |
| Static Credentials | `aws_region` + `aws_access_key_id` + `aws_secret_access_key` | Development or outside AWS |
| Static + AssumeRole | All of the above | Cross-account from outside AWS |

<a id="nestedatt--vault_config"></a>
### Nested Schema for `vault_config`

Required:

- `operation` (String) Operation mode: `IMPORT` (use existing Vault) or `CREATE` (create new Vault instance).

Optional:

- `address` (String) Vault server address (required for IMPORT operation).
- `ca_cert` (String) CA certificate for Vault TLS verification (optional for IMPORT).
- `default_ca_cert_ttl` (String) Default TTL for CA certificates (e.g., `87600h` for 10 years).
- `default_cert_ttl` (String) Default TTL for certificates (e.g., `8760h` for 1 year).
- `kv_mount` (String) KV secrets mount path (default: `secret`).
- `max_ca_cert_ttl` (String) Maximum TTL for CA certificates (e.g., `175200h` for 20 years).
- `max_cert_ttl` (String) Maximum TTL for certificates (e.g., `87600h` for 10 years).
- `mode` (String) Deployment mode for CREATE operation (e.g., `dev`, `prod`).
- `mount` (String) Vault mount path for secrets (KV mount).
- `namespace` (String) Vault namespace (optional).
- `network` (String) Network mode for CREATE operation: `host` or `bridge`.
- `pki_mount` (String) PKI mount path (default: `pki`).
- `port` (Number) Port number for Vault server (used in CREATE mode).
- `token` (String, Sensitive) Vault authentication token (required for IMPORT operation).
- `version` (String) Vault version for CREATE operation (e.g., `1.15.6`). Required for CREATE mode.

## Import

Import is supported using the resource ID:

```shell
terraform import chainlaunch_key_provider.example 123
```
