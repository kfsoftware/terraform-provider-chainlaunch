# Hyperledger Fabric API Plugin - Complete Example

This example demonstrates the **complete end-to-end workflow** for deploying the Hyperledger Fabric API plugin using Terraform. It combines both plugin registration and deployment in a single configuration.

> **Note**: This is a combined example showing both steps together. For learning purposes, see the separate examples:
> - [plugin-definition](../plugin-definition/) - Plugin registration only
> - [plugin-deployment](../plugin-deployment/) - Plugin deployment only

## What This Example Does

1. **Registers Plugin**: Creates the HLF API plugin definition from a YAML specification
2. **Queries Resources**: Discovers existing Fabric organizations and peers
3. **Deploys Plugin**: Deploys the plugin with your Fabric network configuration
4. **Configures API**: Sets up the REST API to connect to your peers and channel
5. **Exposes Metrics**: Enables Prometheus metrics collection

## Two-Step Workflow (Combined Here)

```
┌─────────────────────┐       ┌─────────────────────┐
│  1. Plugin          │       │  2. Plugin          │
│     Definition      │  -->  │     Deployment      │
│  (Step 1 in main.tf)│       │  (Step 2 in main.tf)│
└─────────────────────┘       └─────────────────────┘
       Register                     Deploy with
       from YAML                    parameters
```

## Prerequisites

Before running this example, you need:

1. **Chainlaunch Instance**: A running Chainlaunch server
   ```bash
   # Verify connection
   curl -u admin:admin123 http://localhost:8100/api/v1/health
   ```

2. **Existing Fabric Network**: You must have:
   - A Fabric organization created
   - At least one peer node running
   - A channel created
   - Valid MSP certificates and keys

3. **Network Information**: Gather the following:
   - Organization MSP ID
   - Peer ID(s)
   - Peer external endpoints
   - Channel name
   - TLS CA certificate paths

## Plugin YAML Structure

The `plugin.yaml` file defines the plugin using a Kubernetes-like format:

```yaml
apiVersion: dev.chainlaunch/v1
kind: Plugin
metadata:
  name: hlf-plugin-api        # Plugin identifier
  version: '1.0'              # Plugin version
  description: '...'          # Description
  tags: [fabric, api, rest]   # Tags for discovery

spec:
  dockerCompose:
    contents: |                # Docker Compose configuration with Go templates
      version: '2.2'
      services:
        app:
          image: ghcr.io/kfsoftware/plugin-hlf-api:5ef1ea1
          command: [...]       # Plugin-specific commands

  parameters:                  # JSON Schema for deployment parameters
    type: object
    properties:
      key: {...}              # Fabric identity
      channelName: {...}      # Channel to connect to
      peers: {...}            # List of peers
      port: {...}             # API port
```

### Key Components

- **metadata**: Plugin information (name, version, tags, author)
- **spec.dockerCompose**: Docker Compose template with Go template variables
- **spec.parameters**: JSON Schema defining required/optional parameters
- **spec.metrics**: Prometheus metrics configuration
- **spec.documentation**: README, examples, and troubleshooting

## Configuration

### Required Variables

You must provide these variables (see `terraform.tfvars.example`):

```hcl
peer0_name = "peer0.org1.example.com"  # Name/slug of your first peer
```

### Optional Variables

```hcl
# Connection settings
chainlaunch_url      = "http://localhost:8100"
chainlaunch_username = "admin"
chainlaunch_password = "admin123"

# Organization
organization_msp_id = "Org1MSP"

# Channel
channel_name = "mychannel"

# Second peer (optional)
peer1_name = "peer1.org1.example.com"

# API port
api_port = 8080

# Certificate paths (inside container)
cert_path = "/crypto/peer/cert.pem"
key_path  = "/crypto/peer/key.pem"

# TLS certificate paths (inside container)
peer0_tls_ca_cert_path = "/crypto/peer0/tls/ca.crt"
peer1_tls_ca_cert_path = "/crypto/peer1/tls/ca.crt"  # if using peer1
```

## Usage

### 1. Get Peer Information

First, find your peer names:

```bash
# List all peers
curl -s -u admin:admin123 http://localhost:8100/api/v1/nodes \
  | jq '.[] | select(.type == "PEER") | {name, mspId, externalEndpoint}'

# Example output:
# {
#   "name": "peer0.org1.example.com",
#   "mspId": "Org1MSP",
#   "externalEndpoint": "peer0.org1.example.com:7051"
# }
```

### 2. Create terraform.tfvars

```hcl
peer0_name          = "peer0.org1.example.com"
organization_msp_id = "Org1MSP"
channel_name        = "mychannel"
api_port            = 8080
```

### 3. Deploy

```bash
# Initialize Terraform
terraform init

# Review plan
terraform plan

# Deploy plugin
terraform apply
```

### 4. Verify Deployment

```bash
# Check deployment status
terraform output setup_summary

# Verify containers
docker ps | grep hlf-plugin-api

# Check logs
docker logs $(docker ps -q -f name=hlf-plugin-api)

# Test API
curl http://localhost:8080/api/v1/health
```

## Plugin Parameters Explained

The `parameters` field in the deployment resource is JSON-encoded and follows the plugin's JSON Schema:

```hcl
parameters = jsonencode({
  # Fabric identity (MSP ID, cert, key)
  key = {
    MspID    = "Org1MSP"
    CertPath = "/crypto/peer/cert.pem"    # Path inside container
    KeyPath  = "/crypto/peer/key.pem"     # Path inside container
  }

  # Channel to connect to
  channelName = "mychannel"

  # List of peers to connect to
  peers = [
    {
      ExternalEndpoint = "peer0.org1.example.com:7051"
      TLSCACertPath    = "/crypto/peer0/tls/ca.crt"
    }
  ]

  # API server port
  port = 8080
})
```

### Certificate Paths

The paths in the parameters refer to paths **inside the plugin container**, not on your host. Chainlaunch automatically mounts the necessary certificates based on the `key` and `peers` configuration.

## API Endpoints

Once deployed, the plugin exposes these endpoints:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/health` | GET | Health check |
| `/api/v1/query` | GET | Query chaincode |
| `/api/v1/invoke` | POST | Invoke chaincode |
| `/api/v1/channel` | GET | Get channel info |
| `/metrics` | GET | Prometheus metrics |

### Example API Usage

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Query chaincode
curl "http://localhost:8080/api/v1/query?chaincode=mycc&function=query&args=a"

# Invoke chaincode
curl -X POST http://localhost:8080/api/v1/invoke \
  -H "Content-Type: application/json" \
  -d '{
    "chaincode": "mycc",
    "function": "invoke",
    "args": ["a", "b", "10"]
  }'
```

## Monitoring

The plugin exposes Prometheus metrics at `/metrics`:

```bash
# View metrics
curl http://localhost:8080/metrics

# Example metrics:
# - hlf_api_endpoint_latency_seconds
# - hlf_transaction_queue_size
# - hlf_network_health_status
# - http_requests_total
# - go_goroutines
```

## Updating the Plugin

To update the plugin configuration:

1. Modify `variables.tf` or `terraform.tfvars`
2. Run `terraform apply`

Terraform will:
- Stop the current deployment
- Redeploy with new parameters
- Wait for the new deployment to be ready

## Cleanup

```bash
# Stop and remove plugin deployment
terraform destroy

# This will:
# 1. Stop the plugin deployment (containers)
# 2. Delete the plugin definition
```

Note: This only removes the plugin deployment, not your Fabric network.

## Troubleshooting

### Plugin Fails to Deploy

**Check plugin logs:**
```bash
docker logs $(docker ps -q -f name=hlf-plugin-api)
```

**Common issues:**
- Invalid peer endpoints
- Certificate path mismatches
- Network connectivity issues
- Port conflicts

### API Returns Errors

**Verify peer connectivity:**
```bash
# From inside the plugin container
docker exec -it $(docker ps -q -f name=hlf-plugin-api) sh
ping peer0.org1.example.com
```

**Check certificates:**
```bash
# Verify certificate paths inside container
docker exec $(docker ps -q -f name=hlf-plugin-api) ls -la /crypto/
```

### Deployment Stuck in "deploying"

**Check Chainlaunch logs:**
```bash
# If Chainlaunch is running in Docker
docker logs chainlaunch

# Check deployment status directly
curl -u admin:admin123 http://localhost:8100/api/v1/plugins/hlf-plugin-api/deployment-status
```

## Advanced Usage

### Multi-Peer Deployment

To use multiple peers for redundancy:

```hcl
# terraform.tfvars
peer0_name              = "peer0.org1.example.com"
peer1_name              = "peer1.org1.example.com"
peer0_tls_ca_cert_path  = "/crypto/peer0/tls/ca.crt"
peer1_tls_ca_cert_path  = "/crypto/peer1/tls/ca.crt"
```

The plugin will automatically:
- Query both peers using the data source
- Extract their external endpoints
- Configure the API to use both peers for redundancy

### Custom Plugin Development

To create your own plugin:

1. **Create plugin.yaml** with your Docker Compose configuration
2. **Define parameters** using JSON Schema
3. **Use Go templates** in dockerCompose.contents for dynamic values
4. **Add metrics endpoints** if your service exposes Prometheus metrics
5. **Document** in the documentation section

Example template variables:
- `{{ .parameters.yourParam }}` - Access deployment parameters
- `{{ .volumeMounts }}` - Auto-mounted volumes for certificates
- `{{ range .parameters.list }}...{{ end }}` - Loop over arrays

## Understanding the Two-Step Approach

This example combines both plugin definition and deployment. In production, you might want to separate them:

### Why Separate?

**Plugin Definition** ([plugin-definition](../plugin-definition/)):
- ✅ **Reusable**: Register once, deploy many times
- ✅ **Shared**: Team members can deploy without the YAML file
- ✅ **Versioned**: Track plugin definitions separately
- ✅ **Centralized**: Manage plugin catalog independently

**Plugin Deployment** ([plugin-deployment](../plugin-deployment/)):
- ✅ **Environment-specific**: Different parameters per environment
- ✅ **Multiple instances**: Deploy same plugin with different configs
- ✅ **Ephemeral**: Deploy/destroy without affecting definition
- ✅ **Isolated state**: Separate Terraform state files

### Combined (This Example)

Good for:
- ✅ Quick start and demos
- ✅ Single-tenant deployments
- ✅ Development environments
- ✅ Simple use cases

### Separated (Recommended for Production)

```bash
# Team lead: Register plugin once
cd plugin-definition
terraform apply

# Developers: Deploy as needed
cd plugin-deployment
terraform apply -var="plugin_name=hlf-plugin-api"

# Multiple deployments from same definition
terraform apply -var="plugin_name=hlf-plugin-api" -var="api_port=8080"
terraform apply -var="plugin_name=hlf-plugin-api" -var="api_port=8081"
```

## Related Examples

- **[Plugin Definition](../plugin-definition/)** - Register plugin only (Step 1)
- **[Plugin Deployment](../plugin-deployment/)** - Deploy plugin only (Step 2)
- [Fabric Network](../fabric-network/) - Create a Fabric network first
- [Fabric Peer](../fabric-peer/) - Deploy peer nodes
- [Notifications](../notifications-mailpit/) - Set up notifications for plugin failures

## Additional Resources

- [Plugin YAML Reference](../../docs/plugin-yaml-reference.md)
- [Chainlaunch Plugin API](../../swagger.yaml) - See `/plugins` endpoints
- [HLF API Plugin Repository](https://github.com/kfsoftware/plugin-hlf-api)
