# External Nodes Data Source Example

This example demonstrates how to query **all external nodes** (Fabric peers, orderers, and Besu nodes) using the comprehensive `chainlaunch_external_nodes` data source.

## Overview

The `chainlaunch_external_nodes` data source provides a single, efficient way to retrieve all external nodes that have been synced from connected Chainlaunch peer instances. This is more efficient than using separate data sources for each node type.

## What Are External Nodes?

External nodes are blockchain nodes (peers, orderers, Besu validators) that belong to other Chainlaunch instances but have been synced to your local instance. This enables:

- **Multi-organization Fabric networks** - Organizations can maintain their own Chainlaunch instances but participate in shared channels
- **Cross-instance collaboration** - Different teams/companies can manage their own infrastructure while collaborating
- **Besu consortium networks** - Each consortium member runs their own Chainlaunch with their validators

## Prerequisites

1. **Connected peer nodes** - You must have accepted node invitations from other Chainlaunch instances
2. **Synced external nodes** - Use `chainlaunch_sync_all_external_nodes` or `chainlaunch_external_nodes_sync` to sync nodes
3. **Active connections** - Peer connections must be established and active

## Features

### Single API Call
Unlike using separate data sources (`chainlaunch_external_fabric_peers`, `chainlaunch_external_fabric_orderers`, `chainlaunch_external_besu_nodes`), this data source makes **one API call** to retrieve all node types.

### Complete Node Information
Returns all available data for each node type:

**Fabric Peers**:
- ID, name, MSP ID
- External endpoint
- TLS and signing certificates
- Fabric version

**Fabric Orderers**:
- ID, name, MSP ID
- External endpoint
- TLS and signing certificates
- Fabric version

**Besu Nodes**:
- ID, name
- Enode URL
- P2P host and port
- Version
- Metrics configuration

## Usage

### Basic Usage

```hcl
# Sync external nodes
resource "chainlaunch_sync_all_external_nodes" "sync" {}

# Query all external nodes
data "chainlaunch_external_nodes" "all" {
  depends_on = [chainlaunch_sync_all_external_nodes.sync]
}

# Access the nodes
output "all_peers" {
  value = data.chainlaunch_external_nodes.all.fabric_peers
}
```

### Filtering by MSP ID

```hcl
# Get all peers from a specific organization
locals {
  org1_peers = [
    for peer in data.chainlaunch_external_nodes.all.fabric_peers :
    peer if peer.msp_id == "Org1MSP"
  ]
}

output "org1_peer_count" {
  value = length(local.org1_peers)
}
```

### Building Network Configurations

```hcl
# Use external orderers in channel configuration
locals {
  all_orderer_endpoints = [
    for orderer in data.chainlaunch_external_nodes.all.fabric_orderers : {
      endpoint = orderer.external_endpoint
      tls_cert = orderer.tls_certificate
    }
  ]
}
```

### Extracting Certificates

```hcl
# Export peer certificates for configuration files
output "peer_certs_by_msp" {
  value = {
    for peer in data.chainlaunch_external_nodes.all.fabric_peers :
    "${peer.msp_id}/${peer.name}" => peer.tls_certificate
  }
  sensitive = true
}
```

### Besu Enode URLs

```hcl
# Get all Besu enode URLs for bootnodes configuration
output "besu_bootnodes" {
  description = "Enode URLs for Besu bootnodes"
  value = [
    for node in data.chainlaunch_external_nodes.all.besu_nodes :
    node.enode_url
  ]
}
```

## Complete Workflow

### Instance A (Your Instance)

```hcl
# 1. Connect to other instances
resource "chainlaunch_node_invitation" "invite_org2" {
  bidirectional = true
}

# Share invitation with Org2 out-of-band (email, Slack, etc.)
output "invitation_for_org2" {
  value     = chainlaunch_node_invitation.invite_org2.invitation_jwt
  sensitive = true
}

# 2. Sync external nodes from all connected peers
resource "chainlaunch_sync_all_external_nodes" "sync" {}

# 3. Query all external nodes
data "chainlaunch_external_nodes" "all" {
  depends_on = [chainlaunch_sync_all_external_nodes.sync]
}

# 4. Create a shared channel including external organizations
resource "chainlaunch_fabric_network" "shared_channel" {
  name = "multi-org-channel"

  orderer_endpoints = [
    # Local orderer
    {
      node_id     = tonumber(chainlaunch_fabric_orderer.orderer0.id)
      tls_ca_cert = chainlaunch_fabric_organization.org1.tls_ca_certificate
    }
  ]

  organizations = concat(
    # Local organization
    [{
      organization_id = tonumber(chainlaunch_fabric_organization.org1.id)
    }],
    # External organizations (discovered from synced nodes)
    [
      for org_msp_id in distinct([
        for peer in data.chainlaunch_external_nodes.all.fabric_peers :
        peer.msp_id
      ]) : {
        # External orgs are referenced by MSP ID
        msp_id = org_msp_id
      }
    ]
  )
}
```

## Data Structure

### Fabric Peers
```hcl
data.chainlaunch_external_nodes.all.fabric_peers = [
  {
    id                = 1
    external_node_id  = 10
    name              = "peer0.org2.example.com"
    msp_id            = "Org2MSP"
    external_endpoint = "peer0.org2.example.com:7051"
    version           = "2.5.0"
    sign_certificate  = "-----BEGIN CERTIFICATE-----\n..."
    tls_certificate   = "-----BEGIN CERTIFICATE-----\n..."
  },
  # ... more peers
]
```

### Fabric Orderers
```hcl
data.chainlaunch_external_nodes.all.fabric_orderers = [
  {
    id                = 2
    external_node_id  = 11
    name              = "orderer0.org2.example.com"
    msp_id            = "Org2MSP"
    external_endpoint = "orderer0.org2.example.com:7050"
    version           = "2.5.0"
    sign_certificate  = "-----BEGIN CERTIFICATE-----\n..."
    tls_certificate   = "-----BEGIN CERTIFICATE-----\n..."
  },
  # ... more orderers
]
```

### Besu Nodes
```hcl
data.chainlaunch_external_nodes.all.besu_nodes = [
  {
    id               = 3
    external_node_id = 12
    name             = "besu-validator-1"
    enode_url        = "enode://abc123...@192.168.1.100:30303"
    p2p_host         = "192.168.1.100"
    p2p_port         = 30303
    version          = "24.1.0"
    metrics_enabled  = true
    metrics_port     = 9545
  },
  # ... more Besu nodes
]
```

## Comparison: Single vs Multiple Data Sources

### Using the Comprehensive Data Source (Recommended)
```hcl
# One API call, all node types
data "chainlaunch_external_nodes" "all" {}

output "summary" {
  value = {
    peers    = length(data.chainlaunch_external_nodes.all.fabric_peers)
    orderers = length(data.chainlaunch_external_nodes.all.fabric_orderers)
    besu     = length(data.chainlaunch_external_nodes.all.besu_nodes)
  }
}
```

### Using Individual Data Sources (Legacy)
```hcl
# Three separate API calls
data "chainlaunch_external_fabric_peers" "peers" {}
data "chainlaunch_external_fabric_orderers" "orderers" {}
data "chainlaunch_external_besu_nodes" "besu" {}

output "summary" {
  value = {
    peers    = length(data.chainlaunch_external_fabric_peers.peers.peers)
    orderers = length(data.chainlaunch_external_fabric_orderers.orderers.orderers)
    besu     = length(data.chainlaunch_external_besu_nodes.besu.nodes)
  }
}
```

**Benefits of comprehensive data source**:
- ✅ Single API call (faster)
- ✅ Consistent data snapshot (all nodes at the same point in time)
- ✅ Simpler configuration
- ✅ Easier to maintain

## Common Use Cases

### 1. Channel Configuration Export
Generate connection profiles or channel configuration files:

```hcl
output "connection_profile" {
  value = jsonencode({
    peers = {
      for peer in data.chainlaunch_external_nodes.all.fabric_peers :
      peer.name => {
        url         = "grpcs://${peer.external_endpoint}"
        tlsCACerts  = {
          pem = peer.tls_certificate
        }
        grpcOptions = {
          "ssl-target-name-override" = split(".", peer.name)[0]
        }
      }
    }
  })
}
```

### 2. Multi-Organization Monitoring
Track which organizations are connected:

```hcl
output "connected_organizations" {
  value = {
    msp_ids = distinct(concat(
      [for peer in data.chainlaunch_external_nodes.all.fabric_peers : peer.msp_id],
      [for orderer in data.chainlaunch_external_nodes.all.fabric_orderers : orderer.msp_id]
    ))
    peer_count_by_org = {
      for msp_id in distinct([
        for peer in data.chainlaunch_external_nodes.all.fabric_peers : peer.msp_id
      ]) : msp_id => length([
        for peer in data.chainlaunch_external_nodes.all.fabric_peers :
        peer if peer.msp_id == msp_id
      ])
    }
  }
}
```

### 3. Besu Consortium Bootnodes
Configure Besu network with external validators:

```hcl
resource "chainlaunch_besu_node" "validator" {
  # ... other configuration

  boot_nodes = [
    for node in data.chainlaunch_external_nodes.all.besu_nodes :
    node.enode_url
  ]
}
```

## Troubleshooting

### No External Nodes Returned
**Problem**: All arrays are empty
**Solution**:
1. Verify peer connections: Check that invitations were accepted
2. Trigger sync: Run `terraform apply` to execute sync resource
3. Check API: `curl -u admin:admin123 http://localhost:8100/api/v1/external-nodes`

### Stale Data
**Problem**: External nodes list is outdated
**Solution**:
```hcl
# Force re-sync by tainting the sync resource
# terraform taint chainlaunch_sync_all_external_nodes.sync
# terraform apply
```

### Missing Certificates
**Problem**: `tls_certificate` or `sign_certificate` fields are empty
**Solution**: This indicates the remote instance didn't provide certificates. Check:
- Remote node configuration
- Network connectivity
- Certificate availability on remote instance

### Performance with Many Nodes
**Problem**: Query takes long time with hundreds of nodes
**Solution**: The data source is already optimized (single API call). For very large deployments:
- Filter nodes locally using Terraform expressions
- Use `count` or `for_each` conditionally
- Consider caching outputs in local files

## API Endpoint

This data source calls:
- `GET /external-nodes` - Returns all external nodes in a single response

Response format:
```json
{
  "fabric_peers": [...],
  "fabric_orderers": [...],
  "besu_nodes": [...]
}
```

## Best Practices

1. **Always sync first**: Use `depends_on` to ensure sync completes before querying
2. **Use locals for filtering**: Define filtered lists in locals{} for reuse
3. **Check for empty arrays**: Handle cases where no external nodes exist
4. **Refresh regularly**: Re-apply periodically to pick up new nodes
5. **Sensitive outputs**: Mark certificate outputs as sensitive

## Related Resources

- `chainlaunch_sync_all_external_nodes` - Sync nodes from all connected peers
- `chainlaunch_external_nodes_sync` - Sync nodes from specific peer
- `chainlaunch_node_invitation` - Create peer invitations
- `chainlaunch_node_accept_invitation` - Accept peer invitations
- `chainlaunch_network_share` - Share networks with peers

## See Also

- [Node Invitation Example](../node-invitation/)
- [Network Sharing Example](../network-share/)
- [Fabric Network Example](../fabric-network-complete/)
- [Besu Network Example](../besu-network-complete/)
