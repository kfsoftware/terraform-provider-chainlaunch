# Bidirectional Node Invitations - Multi-Provider Example

This example demonstrates **bidirectional node sharing** between two Chainlaunch instances using provider aliases in a **single Terraform project**.

## Overview

This example manages both Chainlaunch instances in one Terraform configuration using provider aliases.

### Benefits

- ✅ **Single source of truth** - All invitation logic in one place
- ✅ **Automatic dependencies** - Terraform handles the order of operations
- ✅ **Atomic operations** - Both instances are configured together
- ✅ **Easier to maintain** - One `terraform apply` manages both sides

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Terraform Project                            │
│                                                                 │
│  ┌─────────────────────┐         ┌─────────────────────┐      │
│  │   Provider: node1   │         │   Provider: node2   │      │
│  │  localhost:8100     │         │  localhost:8104     │      │
│  └─────────────────────┘         └─────────────────────┘      │
│           │                                   │                │
│           │ 1. Generate Invitation            │                │
│           ├──────────────────────────────────>│                │
│           │                                   │ 2. Accept      │
│           │                                   │                │
│           │                3. Generate Invite │                │
│           │<──────────────────────────────────┤                │
│           │ 4. Accept                         │                │
│           │                                   │                │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

Two Chainlaunch instances must be running:

- **Node 1**: http://localhost:8100 (admin/admin123)
- **Node 2**: http://localhost:8104 (admin/admin)

## Usage

### Quick Start

```bash
cd examples/node-invitations

# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Apply the configuration
terraform apply -auto-approve

# View the connection summary
terraform output connection_summary
```

### Expected Output

```
Apply complete! Resources: 4 added, 0 changed, 0 destroyed.

Outputs:

connection_summary = <<EOT

╔══════════════════════════════════════════════════════════════╗
║        Bidirectional Node Connection Summary                ║
╚══════════════════════════════════════════════════════════════╝

Node 1 → Node 2:
  Invitation:  eyJhbGciOiJFUzI1NiIs...
  Accepted:    ✅ YES

Node 2 → Node 1:
  Invitation:  eyJhbGciOiJFUzI1NiIs...
  Accepted:    ✅ YES

Connection Status: ✅ FULLY ESTABLISHED

EOT
bidirectional_connection_established = true
```

## Resources Created

This example creates **6 resources** and uses **8 data sources**:

### Resources

1. **`chainlaunch_node_invitation.node1_to_node2`**
   - Provider: `node1`
   - Generates invitation from Node 1
   - Bidirectional: `true` (default)

2. **`chainlaunch_node_accept_invitation.node2_accepts_node1`**
   - Provider: `node2`
   - Accepts Node 1's invitation
   - Depends on: `node1_to_node2`

3. **`chainlaunch_node_invitation.node2_to_node1`**
   - Provider: `node2`
   - Generates invitation from Node 2
   - Bidirectional: `true` (default)

4. **`chainlaunch_node_accept_invitation.node1_accepts_node2`**
   - Provider: `node1`
   - Accepts Node 2's invitation
   - Depends on: `node2_to_node1`

5. **`chainlaunch_external_nodes_sync.node2_sync_from_node1`**
   - Provider: `node2`
   - Syncs all external nodes from Node 1 to Node 2
   - Runs on every `terraform apply` to keep nodes synchronized
   - Depends on: `node2_accepts_node1`

6. **`chainlaunch_external_nodes_sync.node1_sync_from_node2`**
   - Provider: `node1`
   - Syncs all external nodes from Node 2 to Node 1
   - Runs on every `terraform apply` to keep nodes synchronized
   - Depends on: `node1_accepts_node2`

### Data Sources

After syncing, the example queries synced external nodes using data sources:

**Node 1 External Nodes:**
- `chainlaunch_external_fabric_organizations.node1_orgs` - Organizations from Node 2
- `chainlaunch_external_fabric_peers.node1_peers` - Peers from Node 2
- `chainlaunch_external_fabric_orderers.node1_orderers` - Orderers from Node 2
- `chainlaunch_external_besu_nodes.node1_besu` - Besu nodes from Node 2

**Node 2 External Nodes:**
- `chainlaunch_external_fabric_organizations.node2_orgs` - Organizations from Node 1
- `chainlaunch_external_fabric_peers.node2_peers` - Peers from Node 1
- `chainlaunch_external_fabric_orderers.node2_orderers` - Orderers from Node 1
- `chainlaunch_external_besu_nodes.node2_besu` - Besu nodes from Node 1

## How It Works

### Step 1: Node 1 Generates Invitation

```hcl
resource "chainlaunch_node_invitation" "node1_to_node2" {
  provider = chainlaunch.node1
  # bidirectional defaults to true
}
```

### Step 2: Node 2 Accepts

```hcl
resource "chainlaunch_node_accept_invitation" "node2_accepts_node1" {
  provider = chainlaunch.node2
  invitation_jwt = chainlaunch_node_invitation.node1_to_node2.invitation_jwt
}
```

Terraform automatically waits for Step 1 to complete before running Step 2.

### Step 3: Node 2 Generates Invitation

```hcl
resource "chainlaunch_node_invitation" "node2_to_node1" {
  provider = chainlaunch.node2
}
```

### Step 4: Node 1 Accepts

```hcl
resource "chainlaunch_node_accept_invitation" "node1_accepts_node2" {
  provider = chainlaunch.node1
  invitation_jwt = chainlaunch_node_invitation.node2_to_node1.invitation_jwt
}
```

### Step 5: Sync External Nodes

After invitations are accepted, sync external nodes from remote instances:

```hcl
# Node 2 syncs nodes from Node 1
resource "chainlaunch_external_nodes_sync" "node2_sync_from_node1" {
  provider     = chainlaunch.node2
  peer_node_id = "node1"

  depends_on = [chainlaunch_node_accept_invitation.node2_accepts_node1]
}

# Node 1 syncs nodes from Node 2
resource "chainlaunch_external_nodes_sync" "node1_sync_from_node2" {
  provider     = chainlaunch.node1
  peer_node_id = "node2"

  depends_on = [chainlaunch_node_accept_invitation.node1_accepts_node2]
}
```

**Important:** This resource runs on **every** `terraform apply` to keep nodes synchronized.

### Step 6: Query External Nodes

After syncing, use data sources to query the external nodes:

```hcl
# Query organizations synced to Node 2
data "chainlaunch_external_fabric_organizations" "node2_orgs" {
  provider   = chainlaunch.node2
  depends_on = [chainlaunch_external_nodes_sync.node2_sync_from_node1]
}

# Query peers synced to Node 2
data "chainlaunch_external_fabric_peers" "node2_peers" {
  provider   = chainlaunch.node2
  depends_on = [chainlaunch_external_nodes_sync.node2_sync_from_node1]
}

# Query orderers synced to Node 2
data "chainlaunch_external_fabric_orderers" "node2_orderers" {
  provider   = chainlaunch.node2
  depends_on = [chainlaunch_external_nodes_sync.node2_sync_from_node1]
}

# Query Besu nodes synced to Node 2
data "chainlaunch_external_besu_nodes" "node2_besu" {
  provider   = chainlaunch.node2
  depends_on = [chainlaunch_external_nodes_sync.node2_sync_from_node1]
}
```

View the synced nodes:

```bash
# View all external organizations
terraform output node2_external_organizations

# View all external peers
terraform output node2_external_peers

# View sync summary
terraform output node2_sync_summary
```

### Step 7: Use External Nodes Data

The example includes practical demonstrations of how to use the synced external nodes data:

**Example 1: Extract specific fields**

```hcl
locals {
  # Get list of all external peer endpoints
  node2_external_peer_endpoints = [
    for peer in data.chainlaunch_external_fabric_peers.node2_peers.peers :
    peer.external_endpoint
  ]

  # Get list of all MSP IDs
  node2_external_msp_ids = [
    for org in data.chainlaunch_external_fabric_organizations.node2_orgs.organizations :
    org.msp_id
  ]

  # Get list of all Besu enode URLs
  node2_external_besu_enodes = [
    for node in data.chainlaunch_external_besu_nodes.node2_besu.nodes :
    node.enode_url
  ]
}
```

**Example 2: Filter nodes by criteria**

```hcl
locals {
  # Get only peers from a specific organization
  org1_peers = [
    for peer in data.chainlaunch_external_fabric_peers.node2_peers.peers :
    peer if peer.msp_id == "Org1MSP"
  ]

  # Get peers running Fabric v2.x
  fabric_v2_peers = [
    for peer in data.chainlaunch_external_fabric_peers.node2_peers.peers :
    peer if can(regex("^2\\.", peer.version))
  ]

  # Get Besu nodes with metrics enabled
  besu_nodes_with_metrics = [
    for node in data.chainlaunch_external_besu_nodes.node2_besu.nodes :
    node if node.metrics_enabled
  ]
}
```

**Example 3: Count and statistics**

```hcl
locals {
  node2_stats = {
    total_organizations = length(data.chainlaunch_external_fabric_organizations.node2_orgs.organizations)
    total_peers         = length(data.chainlaunch_external_fabric_peers.node2_peers.peers)
    total_orderers      = length(data.chainlaunch_external_fabric_orderers.node2_orderers.orderers)
    total_besu_nodes    = length(data.chainlaunch_external_besu_nodes.node2_besu.nodes)
  }
}
```

View the processed data:

```bash
# View extracted endpoints
terraform output node2_external_peer_endpoints
terraform output node2_external_orderer_endpoints

# View filtered results
terraform output org1_peers
terraform output fabric_v2_peers

# View statistics
terraform output node2_stats
```

**Example output:**

```json
node2_external_peer_endpoints = [
  "peer0.org1.example.com:7051",
  "peer1.org1.example.com:7051"
]

node2_stats = {
  "total_besu_nodes" = 0
  "total_orderers" = 3
  "total_organizations" = 2
  "total_peers" = 5
}

org1_peers = [
  {
    "external_endpoint" = "peer0.org1.example.com:7051"
    "id" = 1
    "msp_id" = "Org1MSP"
    "name" = "peer0.org1.example.com"
    "version" = "2.5.0"
    # ... other fields
  }
]
```

## Practical Use Cases

The example includes several commented-out practical examples showing how to use external nodes data in real resources:

### Use Case 1: Monitor External Peers with Prometheus

```hcl
resource "chainlaunch_metrics_job" "external_peers_monitoring" {
  provider = chainlaunch.node2

  job_name = "external-fabric-peers"
  targets = [
    for peer in data.chainlaunch_external_fabric_peers.node2_peers.peers :
    "${peer.external_endpoint}:9443"
  ]
  metrics_path    = "/metrics"
  scrape_interval = "15s"
}
```

### Use Case 2: Configure Besu Bootnodes

```hcl
resource "chainlaunch_besu_node" "node2_besu_node" {
  provider   = chainlaunch.node2
  name       = "besu-node-2"
  network_id = var.besu_network_id

  # Use external Besu nodes as bootnodes
  bootnodes = [
    for node in data.chainlaunch_external_besu_nodes.node2_besu.nodes :
    node.enode_url
  ]

  p2p_host = "0.0.0.0"
  p2p_port = 30303
}
```

### Use Case 3: Join Channel with External Orderers

```hcl
resource "chainlaunch_fabric_join_node" "join_orderer_to_channel" {
  provider   = chainlaunch.node2
  network_id = var.fabric_channel_id
  node_type  = "orderer"

  # Combine local orderer with external orderers from Node 1
  node_ids = concat(
    [chainlaunch_fabric_orderer.local_orderer.id],
    [
      for orderer in data.chainlaunch_external_fabric_orderers.node2_orderers.orderers :
      orderer.id
    ]
  )
}
```

### Use Case 4: Multi-Org Endorsement Policy

```hcl
resource "chainlaunch_fabric_chaincode_definition" "shared_chaincode" {
  provider     = chainlaunch.node2
  chaincode_id = chainlaunch_fabric_chaincode.mycc.id
  version      = "1.0"
  sequence     = 1
  docker_image = "myregistry/mycc:1.0"

  # Endorsement policy requiring signatures from both local and external orgs
  endorsement_policy = "OR(${join(",", concat(
    ["'Org2MSP.peer'"],  # Local organization
    [
      for org in data.chainlaunch_external_fabric_organizations.node2_orgs.organizations :
      "'${org.msp_id}.peer'"
    ]
  ))})"
}
```

### Use Case 5: Generate Application Network Config

```hcl
locals {
  network_config = {
    organizations = {
      for org in data.chainlaunch_external_fabric_organizations.node2_orgs.organizations :
      org.msp_id => {
        msp_id           = org.msp_id
        sign_certificate = org.sign_certificate
        tls_certificate  = org.tls_certificate
      }
    }
    peers = {
      for peer in data.chainlaunch_external_fabric_peers.node2_peers.peers :
      peer.name => {
        endpoint        = peer.external_endpoint
        msp_id          = peer.msp_id
        tls_certificate = peer.tls_certificate
      }
    }
    orderers = {
      for orderer in data.chainlaunch_external_fabric_orderers.node2_orderers.orderers :
      orderer.name => {
        endpoint        = orderer.external_endpoint
        msp_id          = orderer.msp_id
        tls_certificate = orderer.tls_certificate
      }
    }
  }
}

output "network_config" {
  description = "Complete network configuration including external nodes"
  value       = local.network_config
}
```

This generates a complete network configuration that your application can consume.

## Bidirectional vs Unidirectional

### Bidirectional (Default)

When `bidirectional = true` (the default), both instances can share nodes with each other. This is the most common use case.

**Use cases:**
- Multi-org Fabric networks where all orgs collaborate equally
- Development/testing environments sharing nodes
- Consortium blockchains with equal partners

### Unidirectional

Set `bidirectional = false` if only one instance should share nodes:

```hcl
resource "chainlaunch_node_invitation" "node1_to_node2" {
  provider = chainlaunch.node1
  bidirectional = false  # Only Node 1 shares, Node 2 cannot share back
}
```

**Use cases:**
- Service provider sharing infrastructure with clients
- Main network sharing orderers with satellite organizations
- One-way data sharing scenarios

## Troubleshooting

### Both instances must be running

**Error:** `connection refused`

**Solution:**
```bash
# Check Node 1
curl -u admin:admin123 http://localhost:8100/api/v1/health

# Check Node 2
curl -u admin:admin http://localhost:8104/api/v1/health
```

### Provider alias errors

**Error:** `provider "chainlaunch.node1" not found`

**Solution:** Ensure both provider blocks are defined with correct aliases:
```hcl
provider "chainlaunch" {
  alias = "node1"
  # ...
}

provider "chainlaunch" {
  alias = "node2"
  # ...
}
```

### Acceptance fails

Check the error messages:
```bash
terraform output node1_acceptance_status
terraform output node2_acceptance_status
```

## Verification

### Check individual statuses

```bash
# Node 1's invitation
terraform output node1_invitation_id

# Node 2's invitation
terraform output node2_invitation_id

# Overall connection status
terraform output bidirectional_connection_established
```

### View JWTs (sensitive)

```bash
terraform output -raw node1_invitation_jwt
terraform output -raw node2_invitation_jwt
```

## Cleanup

```bash
terraform destroy -auto-approve
```

This removes all invitation resources from both instances.

## Comparison with Separate Configs

| Aspect | Multi-Provider (This) | Separate Configs |
|--------|----------------------|------------------|
| Files | 1 project | 2 projects |
| Apply | 1 command | 2 commands |
| Dependencies | Automatic | Manual |
| State | Single state file | Two state files |
| JWT sharing | Automatic | Manual copy/paste |
| Best for | Testing, automation | Production, separation |

## Next Steps

After establishing the connection, both instances can reference each other's nodes:

```hcl
# On Node 2, reference Node 1's orderer
data "chainlaunch_fabric_orderer" "from_node1" {
  provider = chainlaunch.node2
  id       = var.node1_orderer_id
}

# Use in channel configuration
resource "chainlaunch_fabric_network" "channel" {
  provider    = chainlaunch.node2
  name        = "mychannel"
  orderer_ids = [data.chainlaunch_fabric_orderer.from_node1.id]
}
```

## Related Examples

- [Fabric Network Complete](../fabric-network-complete/) - Full Fabric network with multiple organizations
- [Fabric Peer](../fabric-peer/) - Creating Fabric peer nodes
- [Fabric Orderer](../fabric-orderer/) - Creating Fabric orderer nodes
