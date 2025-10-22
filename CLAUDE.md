# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform Provider for Chainlaunch, a Hyperledger Fabric and Hyperledger Besu blockchain management platform. The provider enables Infrastructure-as-Code management of:
- Organizations (Fabric MSP organizations)
- Nodes (Fabric Peers, Orderers, CAs; Besu Nodes)
- Networks (Fabric Channels, Besu Networks)
- Key Providers (Database, AWS KMS, HashiCorp Vault)
- Cryptographic Keys (RSA, EC, ED25519, secp256k1)
- Chaincodes (Smart Contracts) - Full lifecycle: install, approve, commit, deploy
- Backups (S3-compatible storage targets and automated schedules)
- Notifications (Email alerts via SMTP for backups, node downtime, S3 issues)
- Plugins (Extend platform with custom Docker Compose-based services)

## Build & Development Commands

```bash
# Build the provider
go build -o terraform-provider-chainlaunch

# Build and install locally
make install

# Generate API client from swagger.yaml (when API changes)
make generate

# Format code (Go + Terraform examples)
make fmt

# Run unit tests only
make test-unit

# Run integration tests (requires TF_ACC=1 and Chainlaunch instance)
make test-integration

# Run all tests
make test-all

# Generate coverage report
make test-coverage
```

## Local Development Setup

1. **Build the provider**: `go build -o terraform-provider-chainlaunch`

2. **Configure dev overrides** in `~/.terraformrc`:
```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/kfsoftware/chainlaunch" = "/absolute/path/to/terraform-provider-chainlaunch"
  }
  direct {}
}
```

3. **Run Terraform commands** - Terraform will use your local binary instead of downloading from registry

## Architecture Overview

### Core Components

**Provider Entry Point** (`internal/provider/provider.go`):
- Defines provider configuration (URL, auth credentials)
- Registers all resources and data sources
- Creates HTTP client for API calls

**API Client** (`internal/provider/client.go`):
- Simple REST client using Go's `net/http`
- Handles authentication (API key or username/password)
- `DoRequest(method, path, body)` method for all API calls
- No generated client - manually constructed HTTP requests

**Resources** (`internal/provider/resource_*.go`):
- `resource_fabric_organization.go` - Fabric organizations with MSP
- `resource_node.go` - Generic node resource (legacy)
- `resource_network.go` - Generic network resource (legacy)
- `resource_key_provider.go` - Key storage backends (AWS KMS, Vault)
- `resource_key.go` - Cryptographic keys
- `resource_fabric_peer.go` - Fabric peer nodes
- `resource_fabric_orderer.go` - Fabric orderer nodes
- `resource_fabric_network.go` - Fabric channel/network creation
- `resource_fabric_join_node.go` - Join peers/orderers to channels
- `resource_fabric_anchor_peers.go` - Set anchor peers for organizations (delete is no-op, only updates state)
- `resource_fabric_identity.go` - Fabric admin/client identities with certificates
- `resource_besu_network.go` - Besu network creation with genesis configuration
- `resource_besu_node.go` - Besu validator/node deployment
- `resource_fabric_chaincode.go` - Chaincode (smart contract) records
- `resource_fabric_chaincode_definition.go` - Chaincode definitions (version, sequence, docker image, policy)
- `resource_fabric_chaincode_install.go` - Install chaincode (pull docker image to peers)
- `resource_fabric_chaincode_approve.go` - Approve chaincode for organizations
- `resource_fabric_chaincode_commit.go` - Commit chaincode to channels
- `resource_fabric_chaincode_deploy.go` - Deploy chaincode (start docker containers)
- `resource_backup_target.go` - S3-compatible backup storage configuration
- `resource_backup_schedule.go` - Automated backup schedules with cron expressions
- `resource_notification_provider.go` - Email notification providers (SMTP) for alerts
- `resource_plugin.go` - Plugin definitions from YAML files
- `resource_plugin_deployment.go` - Deploy plugins with parameters

**Data Sources** (`internal/provider/data_source_*.go`):
- Read-only access to existing resources
- `data_source_key_providers.go` - Filtering and default provider logic
- `data_source_fabric_organization.go` - Query Fabric organizations by ID or MSP ID
- `data_source_node.go` - Query nodes by ID
- `data_source_network.go` - Generic network data source (requires ID + type)
- `data_source_fabric_peer.go` - Query Fabric peers by ID
- `data_source_fabric_orderer.go` - Query Fabric orderers by ID
- `data_source_fabric_network.go` - Query Fabric networks by name (returns id, platform, status, etc.)
- `data_source_besu_network.go` - Query Besu networks by name
- `data_source_besu_node.go` - Query Besu nodes by ID
- `data_source_fabric_chaincode.go` - Query chaincode by name and network
- `data_source_plugin.go` - Query plugin information and deployment status

### Key Implementation Patterns

#### Fabric Organization - Using msp_id as name

The Fabric organization resource (`chainlaunch_fabric_organization`) uses `msp_id` as both the MSP identifier and the organization name. This simplifies the resource configuration and ensures consistency.

```go
// In Create and Update methods
createReq := CreateOrganizationRequest{
    Name:        data.MSPID.ValueString(),  // Use msp_id as name
    MSPID:       data.MSPID.ValueString(),
    Description: data.Description.ValueString(),
}
```

**Why**:
- Eliminates the need for a separate `name` field
- Ensures MSP ID and name are always in sync
- Follows Fabric best practices where organization name typically matches MSP ID

#### Update Method Pattern - Preserving Computed Fields

**CRITICAL**: The Update method MUST preserve computed fields from state, especially `created_at`. When Terraform calls Update, `req.Plan` only contains the planned values (user input + any changes), but does NOT include computed fields like `created_at`. If you only use `req.Plan.Get(ctx, &data)`, these fields will be null/unknown, causing Terraform errors like:

```
Error: Provider returned invalid result object after apply
After the apply operation, the provider still indicated an unknown value
for <resource>.<name>.created_at.
```

**ALWAYS use this pattern in Update methods:**

```go
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var data ResourceModel
    var state ResourceModel

    // Get current state to preserve computed fields
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Get plan
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Preserve created_at from state (it's a computed field that never changes)
    data.CreatedAt = state.CreatedAt

    // ... continue with update API call and set state
}
```

**Why**:
- `req.Plan` contains only user-specified values and unknown/null for computed fields
- `req.State` contains the complete current state including all computed fields
- We must preserve computed fields (especially `created_at`) from state
- This pattern prevents Terraform from detecting "unknown" values after apply

**Applied to all resources**:
- `resource_fabric_organization.go`
- `resource_node.go`
- `resource_network.go`
- `resource_key_provider.go`
- `resource_key.go`
- `resource_fabric_peer.go`
- `resource_fabric_orderer.go`

#### Key Resource - is_ca Default Value

The `is_ca` field in the key resource (`chainlaunch_key`) defaults to `false` when not specified. This is implemented in the schema and throughout the resource lifecycle:

**Schema Definition**:
```go
"is_ca": schema.BoolAttribute{
    MarkdownDescription: "Whether this key is a Certificate Authority key (defaults to false)",
    Optional:            true,
    Computed:            true,  // Allows default value
},
```

**Create Method**:
```go
// Default is_ca to false if not specified
if !data.IsCA.IsNull() {
    if data.IsCA.ValueBool() {
        payload["isCA"] = 1
    } else {
        payload["isCA"] = 0
    }
} else {
    payload["isCA"] = 0 // Default to false
    data.IsCA = types.BoolValue(false)
}
```

**Read Method**:
```go
// Set is_ca from API, defaulting to false if not present
data.IsCA = types.BoolValue(key.IsCA)
```

**Why**:
- Most keys are node keys, not CA keys, so `false` is the sensible default
- Reduces boilerplate in Terraform configurations
- Prevents state drift when field is not specified
- Users can explicitly set `is_ca = true` for CA keys

#### Provider Status Checking

When creating key providers (Vault, AWS KMS), the provider waits for readiness:

```go
// In resource_key_provider.go Create method
if data.Type.ValueString() == "VAULT" && vaultConfig.Operation.ValueString() == "CREATE" {
    if err := r.waitForVaultReady(ctx, providerResp.ID); err != nil {
        resp.Diagnostics.AddWarning("Provider Status Check", ...)
    }
}
```

**Vault**: Checks `/key-providers/{id}/vault/status` for:
- `vault_reachable: true`
- `vault_initialized: true`
- `sealed: false`
- `container_running: true`
- Timeout: 60 seconds (30 × 2s)

**AWS KMS**: Checks `/key-providers/{id}/awskms/status` for:
- `kms_reachable: true`
- `has_credentials: true`
- `kms_status: "available"`
- Timeout: 20 seconds (10 × 2s)

#### Nested Configuration Objects

Key providers use nested objects for type-specific config:

```go
// Schema definition
"vault_config": schema.SingleNestedAttribute{
    Optional: true,
    Attributes: map[string]schema.Attribute{
        "operation": schema.StringAttribute{Required: true},
        "address": schema.StringAttribute{Optional: true},
        // ... more fields
    },
}

// Model
type VaultConfigModel struct {
    Operation types.String `tfsdk:"operation"`
    Address   types.String `tfsdk:"address"`
    // ... more fields
}

// Usage in Create
var vaultConfig VaultConfigModel
data.VaultConfig.As(ctx, &vaultConfig, basetypes.ObjectAsOptions{})
```

#### Type Conversion Int64 vs String

IDs are returned as `int64` from API but stored as `types.String` in Terraform:

```go
// Setting ID from API response
data.ID = types.StringValue(fmt.Sprintf("%d", apiResponse.ID))

// Using ID in API call
providerID := data.ProviderID.ValueInt64() // for types.Int64
providerID, _ := strconv.ParseInt(data.ID.ValueString(), 10, 64) // for types.String
```

### Key Provider & Algorithm Support Matrix

| Provider  | RSA | EC (NIST) | secp256k1 | ED25519 | Notes |
|-----------|-----|-----------|-----------|---------|-------|
| Database  | ✅  | ✅        | ✅        | ✅      | Default, all algorithms |
| AWS KMS   | ✅  | ✅        | ✅        | ❌      | Supports blockchain curves |
| Vault     | ✅  | ✅        | ❌        | ❌      | NIST curves only, requires `version` field for CREATE |

**Important**: HashiCorp Vault does NOT support secp256k1 curve. This must be documented in examples and validation.

### Vault CREATE Operation Requirements

When creating a Vault provider with `operation = "CREATE"`, the `version` field is REQUIRED:

```hcl
vault_config = {
  operation = "CREATE"
  mode      = "docker"
  network   = "bridge"
  port      = 8200
  version   = "1.15.6"  # REQUIRED for CREATE
}
```

### Fabric Identity Management

The provider supports creating and managing Fabric admin and client identities with automatic certificate generation:

**Resource** (`chainlaunch_fabric_identity`):
- Creates admin or client identities for Fabric organizations
- Automatically generates keypair and X.509 certificate
- Returns certificate, public key, private key (sensitive)
- Required fields: `organization_id`, `name`, `role` ("admin" or "client")
- Optional fields: `description`, `dns_names` (list), `ip_addresses` (list), `algorithm`
- Computed fields: `certificate`, `public_key`, `sha1_fingerprint`, `ethereum_address`, `expires_at`
- API: `POST /organizations/{id}/keys`, `GET /organizations/{id}/keys`, `DELETE /organizations/{id}/keys/{keyId}`

**Common Use Cases**:
1. **Plugin Deployment**: Create admin identity for HLF API plugin authentication
2. **Application Access**: Generate client identities for application access to Fabric network
3. **Testing**: Create temporary identities for integration tests

**Example**:
```hcl
resource "chainlaunch_fabric_identity" "api_admin" {
  organization_id = chainlaunch_fabric_organization.org1.id
  name            = "hlf-api-admin"
  role            = "admin"
  description     = "Admin identity for HLF API plugin"
  dns_names       = ["api.example.com"]
  algorithm       = "ECDSA"  # Optional, default from org settings
}

# Use in plugin deployment
resource "chainlaunch_plugin_deployment" "hlf_api" {
  plugin_name = "hlf-plugin-api"
  parameters = jsonencode({
    key = {
      keyId = tonumber(chainlaunch_fabric_identity.api_admin.id)
      orgId = tonumber(chainlaunch_fabric_organization.org1.id)
    }
    # ... other parameters
  })
}
```

**Important Notes**:
- Identities are stored in Chainlaunch's database
- Private keys are automatically generated and returned only on creation (sensitive)
- Certificates are valid for organization-defined duration
- Deletion immediately revokes the identity
- Use `role = "admin"` for administrative operations
- Use `role = "client"` for application/user access
- The Read method handles the API's map-based response: `{"keys": {"1": {...}, "2": {...}}}`
- Optional fields (`ethereum_address`, `expires_at`, `last_rotated_at`) are explicitly set to null if not provided by API

### Fabric Chaincode Lifecycle

The provider supports the complete Hyperledger Fabric chaincode lifecycle with **six resources** using a **Docker image-based deployment** approach:

**1. Chaincode Resource** (`chainlaunch_fabric_chaincode`):
- Creates a logical chaincode entity associated with a network
- Can have multiple definitions (versions)
- Required fields: `name`, `network_id`

**2. Chaincode Definition** (`chainlaunch_fabric_chaincode_definition`):
- Defines version, sequence, **docker image**, chaincode address, and endorsement policy
- **The docker image contains the chaincode code**
- **Immutable**: All fields require replacement - any change creates a new definition and destroys the old one
- Sequence must be incremented for each new definition
- Required fields: `chaincode_id`, `version`, `sequence`, `docker_image`, `chaincode_address`
- `chaincode_address`: The address where chaincode service will be accessible (e.g., `mycc.example.com:7052`)
- Optional: `endorsement_policy`
- API: `POST /sc/fabric/chaincodes/{id}/definitions`, `DELETE /sc/fabric/definitions/{definitionId}`

**3. Install** (`chainlaunch_fabric_chaincode_install`):
- **Pulls the Docker image** to specified peers
- Can install on multiple peers in a single operation
- **Idempotent** - "already installed" is treated as success, not an error
- Required fields: `definition_id`, `peer_ids` (list)
- API: `POST /sc/fabric/definitions/{definitionId}/install`
- **No package files needed** - uses the docker image from the definition

**4. Approve** (`chainlaunch_fabric_chaincode_approve`):
- Each organization approves the definition using one of their peers
- References the definition (which contains version, sequence, policy)
- **Idempotent** - "attempted to redefine sequence" is treated as success, not an error
- Required fields: `definition_id`, `peer_id`
- API: `POST /sc/fabric/definitions/{definitionId}/approve`

**5. Commit** (`chainlaunch_fabric_chaincode_commit`):
- Commits the definition to the channel after sufficient approvals
- **Idempotent** - "attempted to redefine sequence" is treated as success, not an error
- Required fields: `definition_id`, `peer_id`
- API: `POST /sc/fabric/definitions/{definitionId}/commit`

**6. Deploy** (`chainlaunch_fabric_chaincode_deploy`):
- **Starts the Docker containers** for the chaincode
- Makes chaincode ready to accept transactions
- Required fields: `definition_id`
- Optional: `environment_variables` (map)
- API: `POST /sc/fabric/definitions/{definitionId}/deploy`
- Deletion calls: `POST /sc/fabric/definitions/{definitionId}/undeploy`

**Workflow Order**:
```
Chaincode → Definition (with docker image) → Install (pulls image to peers) →
Approve (parallel per org) → Commit → Deploy (start containers)
```

**Important Notes**:
- **Docker image required**: Chaincode must be built and pushed as a Docker image before deployment
- **No .tar.gz packages**: Unlike traditional Fabric, no package files are needed
- **Definition-based operations**: Install/Approve/Commit reference the definition ID, not individual parameters
- **Deploy is separate**: The chaincode containers don't start until the deploy step
- **Deletion behavior**:
  - Deploy resource: Actually undeploys (stops containers)
  - Install/Approve/Commit: Remove from state only (operations cannot be reversed)
  - Anchor Peers: Remove from state only (cannot be destroyed, only updated)
- **Upgrading**: Create new definition with incremented sequence, then install/approve/commit/deploy

### Hyperledger Besu Networks

The provider supports creating and managing Hyperledger Besu networks with the **dedicated Besu network resource** (`chainlaunch_besu_network`):

**Besu Network Resource** (`chainlaunch_besu_network`):
- Creates a Besu network with genesis configuration
- Required fields: `name`, `chain_id`, `consensus`, `block_period`, `epoch_length`, `request_timeout`, `initial_validator_key_ids`
- Optional fields: `description`, `gas_limit`, `difficulty`, `mix_hash`, `nonce`, `timestamp`, `coinbase`
- Computed fields: `id`, `status`, `platform`, `created_at`, `updated_at`
- Consensus mechanisms: `qbft` (recommended), `ibft2`
- All initial validator keys must use `secp256k1` curve (Ethereum curve)
- API: `POST /networks/besu`, `GET /networks/besu/{id}`, `DELETE /networks/besu/{id}`

**Key Requirements**:
- **Cryptographic Keys**: All validator keys must be created with `algorithm = "EC"` and `curve = "secp256k1"`
- **Key Provider**: Use `DATABASE` or `AWS_KMS` providers (Vault does NOT support secp256k1)
- **Minimum Validators**: At least 1 validator key required in `initial_validator_key_ids`
- **Immutable Fields**: Most fields require replacement on change (network recreation)

**Workflow**:
```
1. Create secp256k1 keys for validators
2. Create Besu network with validator key IDs
3. Deploy Besu nodes (using chainlaunch_besu_node resource)
```

**Example**:
```hcl
# Create validator keys with secp256k1 curve
resource "chainlaunch_key" "validator_keys" {
  count       = 4
  name        = "besu-validator-${count.index}"
  algorithm   = "EC"
  curve       = "secp256k1"  # Ethereum curve - required for Besu
  provider_id = tonumber(data.chainlaunch_key_providers.db.providers[0].id)
}

# Create Besu network
resource "chainlaunch_besu_network" "main" {
  name            = "my-besu-network"
  description     = "Production Besu network with QBFT consensus"
  chain_id        = 1337
  consensus       = "qbft"
  block_period    = 5
  epoch_length    = 30000
  request_timeout = 10

  initial_validator_key_ids = [
    for key in chainlaunch_key.validator_keys : tonumber(key.id)
  ]
}
```

**Besu Nodes** (`chainlaunch_besu_node`):
- Deploys Besu validator/transaction nodes
- Required fields: `name`, `network_id`, `key_id`, `mode`, `external_ip`, `internal_ip`, `p2p_host`, `p2p_port`, `rpc_host`, `rpc_port`
- Optional: `version`, `boot_nodes`, `min_gas_price`, `metrics_enabled`, `jwt_enabled`, `environment`, etc.
- Uses `/nodes` endpoint with `blockchainPlatform: "BESU"` and `besuNode` configuration object
- Computed fields: `id`, `status`, `created_at`, `updated_at`

**Important Notes**:
- Use `chainlaunch_besu_network` instead of the generic `chainlaunch_network` resource
- The generic resource requires JSON-encoded config which is error-prone
- The dedicated resource provides type-safe fields and better validation
- Network recreation is required for most configuration changes (immutable genesis)
- The API expects `blockchainPlatform: "BESU"` (uppercase) at the top level of node creation requests

### Backup and Restore

The provider supports automated backups to S3-compatible storage (AWS S3, MinIO, etc.) with **two resources**:

**1. Backup Target** (`chainlaunch_backup_target`):
- Configures S3-compatible storage for backups
- Supports AWS S3 and S3-compatible services (MinIO, Wasabi, DigitalOcean Spaces, etc.)
- Required fields: `name`, `type` (currently only "S3"), `region`, `access_key_id`, `secret_access_key`, `bucket_name`, `restic_password`
- Optional: `endpoint` (for non-AWS S3), `bucket_path`, `force_path_style`
- **Important**: `force_path_style` must be `true` for MinIO and most S3-compatible services
- Backups are encrypted using Restic with the provided password
- API: `POST /backups/targets`, `PUT /backups/targets/{id}`, `DELETE /backups/targets/{id}`

**2. Backup Schedule** (`chainlaunch_backup_schedule`):
- Defines automated backup schedules using cron expressions
- Required fields: `name`, `target_id`, `cron_expression`
- Optional: `description`, `enabled` (default: true), `retention_days` (default: 30)
- Computed fields: `last_run_at`, `next_run_at`
- API: `POST /backups/schedules`, `PUT /backups/schedules/{id}`, `DELETE /backups/schedules/{id}`

**Workflow**:
```
Backup Target (S3 storage) → Backup Schedule (cron-based automation)
```

**Cron Expression Examples**:
- `0 0 * * *` - Daily at midnight
- `0 2 * * *` - Daily at 2:00 AM
- `0 */6 * * *` - Every 6 hours
- `0 0 * * 0` - Weekly on Sunday at midnight
- `0 0 1 * *` - Monthly on the 1st

**Example with MinIO**:
```hcl
# Backup target pointing to MinIO
resource "chainlaunch_backup_target" "minio_target" {
  name               = "MinIO Local Backup"
  type               = "S3"
  endpoint           = "http://localhost:9000"
  region             = "us-east-1"
  access_key_id      = "backup-user"
  secret_access_key  = "backup-password"
  bucket_name        = "chainlaunch-backups"
  bucket_path        = "fabric-backups"
  force_path_style   = true  # Required for MinIO
  restic_password    = "encryption-password"
}

# Daily backup schedule
resource "chainlaunch_backup_schedule" "daily" {
  name            = "Daily Backup"
  target_id       = chainlaunch_backup_target.minio_target.id
  cron_expression = "0 2 * * *"
  retention_days  = 30
}
```

### Plugins

The provider supports extending the Chainlaunch platform with custom functionality using **plugins**. Plugins are Docker Compose-based services defined using a Kubernetes-like YAML format.

**Architecture**:
- **Plugin Definition** (`chainlaunch_plugin`): Creates the plugin from a YAML file
- **Plugin Deployment** (`chainlaunch_plugin_deployment`): Deploys the plugin with specific parameters

**Plugin YAML Structure**:
```yaml
apiVersion: dev.chainlaunch/v1
kind: Plugin
metadata:
  name: plugin-name
  version: '1.0'
  description: 'Plugin description'
  author: 'Author Name'
  tags: [tag1, tag2]
  repository: 'https://github.com/...'
  license: 'Apache-2.0'

spec:
  dockerCompose:
    contents: |
      # Docker Compose configuration with Go templates
      # Use {{ .parameters.fieldName }} to access deployment parameters
      # Use {{ range .volumeMounts }}...{{ end }} for auto-mounted volumes

  parameters:
    # JSON Schema defining required/optional parameters
    $schema: http://json-schema.org/draft-07/schema#
    type: object
    properties:
      fieldName:
        type: string
        title: Field Title
        description: Field description

  metrics:
    endpoints:
      - service: service-name
        port: '{{ .parameters.port }}'
        path: /metrics

  documentation:
    readme: |
      # Plugin documentation
    examples:
      - name: 'Example name'
        parameters: {...}
```

**Key Features**:
1. **YAML-based Definition**: Kubernetes-like structure with metadata and spec
2. **Docker Compose**: Full Docker Compose support with Go template variables
3. **Parameter Schema**: JSON Schema validation for deployment parameters
4. **Auto Volume Mounts**: Chainlaunch automatically mounts certificates/keys
5. **Metrics Integration**: Prometheus metrics endpoints configuration
6. **Documentation**: Inline README, examples, and troubleshooting

**Resource Workflow**:
```
Plugin YAML File → chainlaunch_plugin (register) →
chainlaunch_plugin_deployment (deploy with params) → Running Services
```

**Example - HLF API Plugin**:
```hcl
# Register plugin from YAML file
resource "chainlaunch_plugin" "hlf_api" {
  yaml_file_path = "${path.module}/plugin.yaml"
}

# Deploy plugin with parameters
resource "chainlaunch_plugin_deployment" "hlf_api" {
  plugin_name = chainlaunch_plugin.hlf_api.name

  parameters = jsonencode({
    key = {
      MspID    = "Org1MSP"
      CertPath = "/crypto/peer/cert.pem"
      KeyPath  = "/crypto/peer/key.pem"
    }
    channelName = "mychannel"
    peers = [
      {
        ExternalEndpoint = "peer0.org1.example.com:7051"
        TLSCACertPath    = "/crypto/peer0/tls/ca.crt"
      }
    ]
    port = 8080
  })
}
```

**Deployment Lifecycle**:
- **Create**: Deploys Docker Compose services, waits for ready status
- **Read**: Queries deployment status (running, stopped, error)
- **Update**: Stops current deployment, redeploys with new parameters
- **Delete**: Stops and removes Docker Compose services

**Status Monitoring**:
The deployment resource tracks:
- `status`: Current deployment status (running, stopped, error)
- `project_name`: Docker Compose project name
- `started_at`: Deployment start timestamp
- `stopped_at`: Stop timestamp (if applicable)
- `error`: Error message if deployment failed

**Important Notes**:
- Plugin YAML can be provided via `yaml_file_path` or `yaml_content`
- Parameters are JSON-encoded and validated against the plugin's schema
- Certificate paths refer to paths **inside containers**, not host paths
- Chainlaunch auto-mounts required volumes based on parameter types
- Deployment waits up to 60 seconds for services to reach ready state
- API: `POST /plugins`, `PUT /plugins/{name}`, `DELETE /plugins/{name}`, `POST /plugins/{name}/deploy`, `POST /plugins/{name}/stop`

## Testing

### Test Organization

```
test/
├── manual/              # Manual testing configurations
│   ├── test-key-providers.tf
│   ├── test-vault-status.tf
│   └── test-awskms-status.tf
└── e2e/                # End-to-end test scenarios
```

### Running Tests

```bash
# Quick manual test with default provider
cd test/manual
terraform plan
terraform apply -auto-approve

# Test specific provider
terraform apply -target=chainlaunch_key_provider.vault_managed

# Integration tests (requires Chainlaunch instance)
export TF_ACC=1
export CHAINLAUNCH_URL="http://localhost:8100"
export CHAINLAUNCH_USERNAME="admin"
export CHAINLAUNCH_PASSWORD="admin123"
make test-integration
```

## Working with Examples

Examples are located in `examples/` and demonstrate:
- `keys-database/` - All key types with database provider
- `keys-aws-kms/` - RSA/EC keys including secp256k1
- `keys-vault/` - RSA/EC keys (NIST only, CREATE mode)
- `vault-provider/` - Vault provider IMPORT and CREATE modes
- `aws-kms-provider/` - AWS KMS with LocalStack
- `fabric-peer/` - Creating Fabric peer nodes
- `fabric-orderer/` - Creating Fabric orderer nodes
- `fabric-network/` - Creating Fabric channels with anchor peers
- `fabric-chaincode/` - Complete chaincode lifecycle (install/approve/commit)
- `besu-network-complete/` - Complete Besu network with QBFT/IBFT2 consensus and multiple nodes
- `backup-with-minio/` - Automated backups with MinIO S3-compatible storage
- `notifications-mailpit/` - Email notifications using Mailpit SMTP server
- `plugin-definition/` - Register a plugin from YAML specification (Part 1 of plugin workflow)
- `plugin-deployment/` - Deploy a registered plugin with parameters (Part 2 of plugin workflow)
- `plugin-hlf-api/` - Complete end-to-end: register + deploy Hyperledger Fabric REST API plugin

Each example has a comprehensive README with configuration details and troubleshooting.

## API Client Generation

The provider does NOT use generated clients. Instead:
- Manual HTTP client in `client.go`
- `swagger.yaml` is reference documentation only
- Direct JSON marshaling/unmarshaling for API requests

If API changes are needed:
1. Update `swagger.yaml` (for documentation)
2. Update `client.go` or resource files manually
3. Test with integration tests

## Common Issues & Solutions

### Issue: Organization recreates on every apply
**Cause**: API doesn't return `name` field in GET response
**Solution**: Preserve name from state in Read method (already implemented)

### Issue: Vault provider creation hangs
**Cause**: Status check waiting for Vault to initialize
**Solution**: Increase timeout or check Vault logs. Status check has 60s timeout.

### Issue: Keys fail with secp256k1 on Vault
**Cause**: Vault doesn't support this curve
**Solution**: Use AWS KMS or database provider for secp256k1 keys

### Issue: Provider development overrides not working
**Cause**: Terraform using registry version
**Solution**: Check `~/.terraformrc` has correct absolute path to binary

## Code Patterns to Follow

1. **Always preserve fields not returned by API** - Check Read methods
2. **CRITICAL: Preserve computed fields in Update methods** - Get state first, then plan, then preserve `created_at` from state. See "Update Method Pattern" section above.
3. **Use status checking for managed resources** - Vault CREATE, future AWS resources
4. **Validate constraints in schema** - Mark required fields, document limitations
5. **Use consistent error messages** - Include API response in diagnostics
6. **Test both IMPORT and CREATE modes** - For key providers
7. **Document provider-specific limitations** - secp256k1, version requirements
8. **Use warnings for non-fatal issues** - Status checks, deprecated configs

## File Organization

```
internal/provider/
├── provider.go              # Main provider registration
├── client.go                # HTTP client
├── resource_*.go            # Resource implementations
├── data_source_*.go         # Data source implementations
└── *_test.go                # Integration tests

docs/
├── index.md                 # Provider homepage for Terraform Registry
├── README.md                # Documentation guidelines
├── resources/               # Resource documentation
│   ├── fabric_organization.md
│   ├── fabric_identity.md
│   └── ...
├── data-sources/            # Data source documentation
│   ├── fabric_organization.md
│   └── ...
└── guides/                  # User guides and tutorials
    ├── getting-started.md
    ├── fabric-network-setup.md
    └── ...

examples/
├── keys-*/                  # Key creation examples
├── *-provider/              # Provider configuration examples
└── README.md                # Examples overview

test/
├── manual/                  # Quick testing configs
└── e2e/                     # Full scenario tests
```

## Provider Documentation

Documentation in the `docs/` directory is automatically published to the Terraform Registry. Follow these guidelines when adding or updating documentation:

### Documentation Structure

- **`docs/index.md`**: Provider homepage with overview, authentication methods, and links to all resources
- **`docs/resources/`**: One file per resource with usage examples, schema, and troubleshooting
- **`docs/data-sources/`**: One file per data source with query examples and filters
- **`docs/guides/`**: Step-by-step tutorials for common workflows

### Writing Documentation

1. **Include Frontmatter**: Every doc file needs YAML frontmatter:
   ```yaml
   ---
   page_title: "chainlaunch_fabric_identity Resource - Chainlaunch"
   subcategory: "Fabric"
   description: |-
     Brief description for SEO and search results.
   ---
   ```

2. **Follow the Structure**:
   - Title (H1)
   - Example Usage (H2) - Multiple examples from basic to advanced
   - Schema (H2) - Required, Optional, and Computed fields
   - Import (H2) - For resources only
   - Usage Notes (H2) - Important details and best practices
   - Error Handling (H2) - Common errors and solutions
   - See Also (H2) - Cross-references to related docs

3. **Use Practical Examples**: All code examples should be complete and runnable
   ```terraform
   resource "chainlaunch_fabric_organization" "org1" {
     msp_id = "Org1MSP"
   }
   ```

4. **Test Documentation Locally**:
   ```bash
   # Install doc preview tool
   go install github.com/hashicorp/terraform-registry-docs@latest

   # Preview docs
   terraform-registry-docs preview .
   ```

5. **Keep Cross-References**: Link related resources using relative paths:
   - Same directory: `[text](filename.md)`
   - From guides: `[text](../resources/filename.md)`

See `docs/README.md` for complete documentation standards and style guide.
