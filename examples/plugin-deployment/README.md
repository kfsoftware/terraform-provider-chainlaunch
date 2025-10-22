# Plugin Deployment Example

This example demonstrates how to **deploy a registered plugin** in Chainlaunch using Terraform. This is **Part 2** of the plugin workflow - deploying the plugin with specific runtime parameters.

## What This Example Does

1. **Queries Plugin**: References an already-registered plugin definition
2. **Gathers Resources**: Queries existing Fabric resources (organizations, peers)
3. **Deploys Services**: Starts Docker Compose services with provided parameters
4. **Monitors Status**: Tracks deployment status and health

## Two-Step Plugin Workflow

```
┌─────────────────────┐       ┌─────────────────────┐
│  1. Plugin          │       │  2. Plugin          │
│     Definition      │  -->  │     Deployment      │
│  (Separate Example) │       │  (This Example)     │
└─────────────────────┘       └─────────────────────┘
       Register                     Deploy with
       YAML spec                    parameters
```

**Previous example**: Register the plugin (define WHAT it is)
**This example**: Deploy the plugin (define HOW to run it)

## Prerequisites

Before deploying, you must have:

1. **Plugin Registered**: Run the [plugin-definition example](../plugin-definition/) first
   ```bash
   cd ../plugin-definition
   terraform apply
   ```

2. **Fabric Network Ready**:
   - At least one organization created
   - At least one peer node running
   - A channel created
   - Valid MSP certificates

3. **Resource Information**:
   - Organization MSP ID
   - Peer name(s)
   - Channel name

## Quick Start

### 1. Verify Plugin Exists

```bash
# Check if plugin is registered
curl -s -u admin:admin123 http://localhost:8100/api/v1/plugins/hlf-plugin-api

# Or list all plugins
curl -s -u admin:admin123 http://localhost:8100/api/v1/plugins | jq '.[].metadata.name'
```

### 2. Get Resource Names

```bash
# List peers
curl -s -u admin:admin123 http://localhost:8100/api/v1/nodes \
  | jq '.[] | select(.type == "PEER") | {name, mspId, externalEndpoint}'
```

### 3. Configure

Create `terraform.tfvars`:

```hcl
plugin_name         = "hlf-plugin-api"
peer0_name          = "peer0.org1.example.com"
organization_msp_id = "Org1MSP"
identity_name       = "hlf-api-admin"
channel_name        = "mychannel"
api_port            = 8080
```

### 4. Deploy

```bash
# Initialize
terraform init

# Deploy the plugin
terraform apply

# View deployment status
terraform output setup_summary
```

## Configuration

### Required Variables

```hcl
plugin_name = "hlf-plugin-api"  # Must match registered plugin
peer0_name  = "peer0.org1.example.com"  # At least one peer required
```

### Optional Variables

See [variables.tf](variables.tf) for all available options:

- `plugin_name`: Name of the registered plugin (default: "hlf-plugin-api")
- `organization_msp_id`: Organization MSP ID (default: "Org1MSP")
- `channel_name`: Fabric channel name (default: "mychannel")
- `peer0_name`: First peer name (required)
- `peer1_name`: Second peer name (optional, for redundancy)
- `identity_name`: Admin identity name (default: "hlf-api-admin")
- `api_port`: API server port (default: 8080)

## How It Works

### 1. Plugin Lookup

```hcl
data "chainlaunch_plugin" "hlf_api" {
  name = var.plugin_name
}
```

Queries the registered plugin to get its definition, parameter schema, and Docker Compose template.

### 2. Resource Discovery

```hcl
data "chainlaunch_fabric_peer" "peer0" {
  name = var.peer0_name
}
```

Automatically discovers peer endpoints - no need to manually specify!

### 3. Identity Creation

```hcl
resource "chainlaunch_fabric_identity" "api_admin" {
  organization_id = data.chainlaunch_fabric_organization.org1.id
  name            = var.identity_name
  role            = "admin"
  description     = "Admin identity for HLF API plugin"
}
```

Creates an admin identity with certificate and private key automatically.

### 4. Parameter Assembly

```hcl
# Expected format: {"channelName":"mychannel","key":{"keyId":508,"orgId":49},"peers":[125],"port":9501}
parameters = jsonencode({
  channelName = var.channel_name
  key = {
    keyId = tonumber(chainlaunch_fabric_identity.api_admin.id)
    orgId = tonumber(data.chainlaunch_fabric_organization.org1.id)
  }
  peers = [tonumber(data.chainlaunch_fabric_peer.peer0.id)]
  port = var.api_port
})
```

Parameters use identity and resource IDs - no manual certificate management needed!

### 5. Deployment

```hcl
resource "chainlaunch_plugin_deployment" "hlf_api" {
  plugin_name = data.chainlaunch_plugin.hlf_api.name
  parameters  = jsonencode({...})
}
```

Chainlaunch:
- Renders the Docker Compose template with your parameters
- Mounts required certificates/keys automatically
- Starts the Docker Compose services
- Waits for services to be healthy

## What Gets Created

After deployment:

- ✅ **Admin Identity**: Certificate and private key for API authentication
- ✅ **Docker Containers**: Services from the plugin's Docker Compose
- ✅ **Volume Mounts**: Auto-mounted certificates and keys
- ✅ **Network Configuration**: Container networking setup
- ✅ **Deployment Record**: Status tracking in Chainlaunch

## Verify Deployment

```bash
# Check container status
docker ps | grep hlf-plugin-api

# View logs
docker logs $(docker ps -q -f name=hlf-plugin-api)

# Test API
curl http://localhost:8080/api/v1/health

# View metrics
curl http://localhost:8080/metrics
```

## Updating Deployment

To change parameters:

1. **Edit terraform.tfvars**:
   ```hcl
   api_port = 9090  # Change port
   peer1_name = "peer1.org1.example.com"  # Add second peer
   ```

2. **Apply changes**:
   ```bash
   terraform apply
   ```

Terraform will:
- Stop the current deployment
- Redeploy with new parameters
- Wait for the new deployment to be ready

## Multiple Deployments

You can deploy the same plugin multiple times with different parameters:

```hcl
# Identity for Org1
resource "chainlaunch_fabric_identity" "org1_admin" {
  organization_id = data.chainlaunch_fabric_organization.org1.id
  name            = "org1-api-admin"
  role            = "admin"
}

# Deployment 1: Organization 1
resource "chainlaunch_plugin_deployment" "org1_api" {
  plugin_name = "hlf-plugin-api"
  parameters = jsonencode({
    channelName = "channel1"
    key = {
      keyId = tonumber(chainlaunch_fabric_identity.org1_admin.id)
      orgId = tonumber(data.chainlaunch_fabric_organization.org1.id)
    }
    peers = [tonumber(data.chainlaunch_fabric_peer.org1_peer0.id)]
    port  = 8080
  })
}

# Identity for Org2
resource "chainlaunch_fabric_identity" "org2_admin" {
  organization_id = data.chainlaunch_fabric_organization.org2.id
  name            = "org2-api-admin"
  role            = "admin"
}

# Deployment 2: Organization 2
resource "chainlaunch_plugin_deployment" "org2_api" {
  plugin_name = "hlf-plugin-api"
  parameters = jsonencode({
    channelName = "channel2"
    key = {
      keyId = tonumber(chainlaunch_fabric_identity.org2_admin.id)
      orgId = tonumber(data.chainlaunch_fabric_organization.org2.id)
    }
    peers = [tonumber(data.chainlaunch_fabric_peer.org2_peer0.id)]
    port  = 8081  # Different port
  })
}
```

## Deployment Status

The deployment resource tracks:

- `status`: Current state (running, stopped, error)
- `project_name`: Docker Compose project name
- `started_at`: When deployment started
- `stopped_at`: When deployment stopped (if applicable)
- `error`: Error message (if deployment failed)

```bash
# Query status
terraform output deployment_status

# Or via API
curl -s -u admin:admin123 \
  http://localhost:8100/api/v1/plugins/hlf-plugin-api/deployment-status
```

## Troubleshooting

### Plugin Not Found

```
Error: Plugin "hlf-plugin-api" not found
```

**Solution**: Register the plugin first:
```bash
cd ../plugin-definition
terraform apply
```

### Peer Not Found

```
Error: Peer "peer0.org1.example.com" not found
```

**Solution**: Verify peer name:
```bash
curl -s -u admin:admin123 http://localhost:8100/api/v1/nodes \
  | jq '.[] | select(.type == "PEER") | .name'
```

### Parameter Validation Failed

```
Error: Invalid parameters for plugin deployment
```

**Solution**: Check the plugin's parameter schema:
```bash
curl -s -u admin:admin123 http://localhost:8100/api/v1/plugins/hlf-plugin-api \
  | jq '.spec.parameters'
```

### Deployment Stuck

If deployment status shows "deploying" for more than 2 minutes:

```bash
# Check container logs
docker logs $(docker ps -a -q -f name=hlf-plugin-api)

# Check Chainlaunch logs
docker logs chainlaunch  # If running in Docker

# Force stop and retry
terraform destroy
terraform apply
```

### Port Already in Use

```
Error: Port 8080 already in use
```

**Solution**: Change the port:
```hcl
api_port = 8081
```

### Certificate Errors

```
Error: TLS handshake failed
```

**Solution**: Verify certificate paths:
```bash
# Check what Chainlaunch mounted
docker exec $(docker ps -q -f name=hlf-plugin-api) ls -la /crypto/
```

## Cleanup

```bash
# Stop deployment
terraform destroy
```

This:
- ✅ Stops Docker Compose services
- ✅ Removes containers
- ✅ Cleans up deployment record
- ❌ Does NOT delete the plugin definition

To also remove the plugin definition:
```bash
cd ../plugin-definition
terraform destroy
```

## API Operations

This example uses:

- **GET /plugins/{name}**: Query plugin definition
- **POST /plugins/{name}/deploy**: Deploy plugin with parameters
- **GET /plugins/{name}/deployment-status**: Check deployment status
- **POST /plugins/{name}/stop**: Stop deployment
- **GET /plugins/{name}/services**: List running services

## Advanced Usage

### Using Remote State

Reference plugin from another Terraform workspace:

```hcl
data "terraform_remote_state" "plugin" {
  backend = "s3"
  config = {
    bucket = "my-terraform-state"
    key    = "plugin-definition/terraform.tfstate"
  }
}

resource "chainlaunch_plugin_deployment" "api" {
  plugin_name = data.terraform_remote_state.plugin.outputs.plugin_name
  parameters  = jsonencode({...})
}
```

### Dynamic Parameters

Use locals for complex parameter logic:

```hcl
locals {
  # Build list of peer IDs
  peer_ids = concat(
    [tonumber(data.chainlaunch_fabric_peer.peer0.id)],
    var.peer1_name != "" ? [tonumber(data.chainlaunch_fabric_peer.peer1[0].id)] : []
  )
}

resource "chainlaunch_plugin_deployment" "api" {
  parameters = jsonencode({
    channelName = var.channel_name
    key = {
      keyId = tonumber(chainlaunch_fabric_identity.api_admin.id)
      orgId = tonumber(data.chainlaunch_fabric_organization.org1.id)
    }
    peers = local.peer_ids
    port  = var.api_port
  })
}
```

### Conditional Deployment

Deploy only if certain conditions are met:

```hcl
resource "chainlaunch_plugin_deployment" "api" {
  count = var.enable_api ? 1 : 0

  plugin_name = "hlf-plugin-api"
  parameters  = jsonencode({...})
}
```

## Related Examples

- [Plugin Definition](../plugin-definition/) - Register a plugin first (required)
- [Plugin HLF API](../plugin-hlf-api/) - Complete end-to-end example (both parts combined)
- [Fabric Network](../fabric-network/) - Set up a Fabric network
- [Fabric Peer](../fabric-peer/) - Deploy peer nodes

## Additional Resources

- [Plugin YAML Reference](../../docs/plugin-yaml-reference.md)
- [Chainlaunch Plugin API](../../swagger.yaml) - See `/plugins/{name}/deploy` endpoint
- [Docker Compose Docs](https://docs.docker.com/compose/) - Understanding the plugin runtime
