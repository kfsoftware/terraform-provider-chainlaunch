# Complete Fabric Network Example

This comprehensive example demonstrates how to deploy a complete Hyperledger Fabric network with fully configurable organizations, peers, orderers, and channel configuration using Terraform variables.

## Overview

This example creates a production-ready Fabric network with:
- **Multiple peer organizations** (configurable)
- **Multiple peers per organization** (configurable)
- **Multiple orderer organizations** (configurable)
- **Multiple orderers (consenters)** for Raft consensus (configurable)
- **Automatic channel creation and node joining**
- **Anchor peer configuration**
- **Chaincode deployment** (optional, full lifecycle)
- **Complete network lifecycle management**

All resources are defined in a single `main.tf` file, with configuration managed through variables.

## Features

### ✅ Fully Variable-Driven
- Add or remove organizations by modifying `peer_organizations` and `orderer_organizations` variables
- Adjust number of peers and orderers without touching the main configuration
- Configure all node properties through variables

### ✅ Production-Ready Defaults
- Etcd Raft consensus with 3 orderers
- Fabric 3.x versions (Peer 3.1.2, Orderer 3.1.1)
- Certificate auto-renewal enabled
- Operations endpoints for monitoring
- Proper environment variables

### ✅ Flexible Architecture
- Support for single or multiple orderer organizations
- Configurable anchor peers per organization
- Custom domain names and endpoints
- Environment variable customization

### ✅ Complete Chaincode Lifecycle
- Optional chaincode deployment (enable with `deploy_chaincode = true`)
- Docker image-based deployment
- Automatic install, approve, commit, and deploy
- Support for multiple chaincodes
- Custom endorsement policies
- Chaincode upgrades (increment version and sequence)

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Fabric Network                            │
│                        (mychannel)                               │
│                                                                  │
│  ┌──────────────────────┐      ┌──────────────────────┐       │
│  │   Org1MSP            │      │   Org2MSP            │       │
│  │                      │      │                      │       │
│  │  ┌──────┐ ┌──────┐  │      │  ┌──────┐ ┌──────┐  │       │
│  │  │peer0 │ │peer1 │  │      │  │peer0 │ │peer1 │  │       │
│  │  │(anch)│ │      │  │      │  │(anch)│ │      │  │       │
│  │  └──────┘ └──────┘  │      │  └──────┘ └──────┘  │       │
│  └──────────────────────┘      └──────────────────────┘       │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              OrdererOrgMSP (Consenters)                  │  │
│  │                                                           │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐              │  │
│  │  │orderer0  │  │orderer1  │  │orderer2  │              │  │
│  │  │          │  │          │  │          │              │  │
│  │  └──────────┘  └──────────┘  └──────────┘              │  │
│  │                  Etcd Raft Consensus                     │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

1. **Chainlaunch instance** running and accessible
2. **Terraform** installed (>= 1.0)
3. **Chainlaunch Terraform Provider** built and configured
4. **Available ports** for peer and orderer endpoints

## Quick Start

### 1. Copy the Example Configuration

```bash
cd examples/fabric-network-complete
cp terraform.tfvars.example terraform.tfvars
```

### 2. Customize Your Network

Edit `terraform.tfvars` to configure your network. The default configuration creates:
- **2 peer organizations** (Org1MSP, Org2MSP)
- **2 peers per organization** (4 peers total)
- **1 orderer organization** (OrdererOrgMSP)
- **3 orderers** for Raft consensus
- **1 channel** (mychannel)

### 3. Review the Plan

```bash
terraform init
terraform plan
```

Expected resources:
- 2 peer organizations
- 1 orderer organization
- 4 peer nodes
- 3 orderer nodes
- 1 fabric network (channel)
- 7 join operations (4 peers + 3 orderers)
- 2 anchor peer configurations

### 4. Deploy the Network

```bash
terraform apply
```

This will:
1. Create all organizations
2. Deploy all peer and orderer nodes
3. Create the Fabric channel
4. Join all peers and orderers to the channel
5. Configure anchor peers for cross-org communication

### 5. Verify the Deployment

```bash
# View all outputs
terraform output

# View specific information
terraform output channel_id
terraform output peers
terraform output orderers
```

## Configuration Examples

### Example 1: Two Organizations (Default)

**2 peer orgs**, **2 peers each**, **1 orderer org**, **3 orderers**

```hcl
peer_organizations = {
  org1 = {
    msp_id = "Org1MSP"
    peers = {
      peer0 = { ... }
      peer1 = { ... }
    }
  }
  org2 = {
    msp_id = "Org2MSP"
    peers = {
      peer0 = { ... }
      peer1 = { ... }
    }
  }
}
```

### Example 2: Three Organizations

**3 peer orgs**, **1 peer each**, **1 orderer org**, **3 orderers**

```hcl
peer_organizations = {
  org1 = {
    msp_id = "Org1MSP"
    peers = { peer0 = { ... } }
  }
  org2 = {
    msp_id = "Org2MSP"
    peers = { peer0 = { ... } }
  }
  org3 = {
    msp_id = "Org3MSP"
    peers = { peer0 = { ... } }
  }
}
```

See `terraform.tfvars.example` for complete configuration.

### Example 3: Multiple Orderer Organizations

**2 orderer orgs**, **2 orderers each** (4 total consenters)

```hcl
orderer_organizations = {
  orderer_org1 = {
    msp_id = "OrdererOrg1MSP"
    orderers = {
      orderer0 = { ... }
      orderer1 = { ... }
    }
  }
  orderer_org2 = {
    msp_id = "OrdererOrg2MSP"
    orderers = {
      orderer0 = { ... }
      orderer1 = { ... }
    }
  }
}
```

### Example 4: Minimal Development Setup

**2 orgs**, **1 peer each**, **1 orderer org**, **1 orderer** (solo/single node)

```hcl
peer_organizations = {
  org1 = {
    msp_id = "Org1MSP"
    peers = { peer0 = { ... } }
  }
  org2 = {
    msp_id = "Org2MSP"
    peers = { peer0 = { ... } }
  }
}

orderer_organizations = {
  orderer_org = {
    msp_id = "OrdererOrgMSP"
    orderers = { orderer0 = { ... } }
  }
}
```

## Variable Configuration

### Key Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `peer_organizations` | Map of peer organizations with their peers | 2 orgs, 2-1 peers |
| `orderer_organizations` | Map of orderer organizations with orderers | 1 org, 3 orderers |
| `channel_name` | Name of the Fabric channel | `mychannel` |
| `consensus_type` | Consensus algorithm | `etcdraft` |
| `etcdraft_options` | Raft consensus parameters | Production defaults |
| `batch_size` | Transaction batch size limits | Standard values |
| `batch_timeout` | Transaction batch timeout | `2s` |

### Port Allocation Guide

**Peer Ports:**
- Organization 1: 7000-7999 range
  - Peer 0: 7051, 7052, 7053, 9443
  - Peer 1: 7151, 7152, 7153, 9543
- Organization 2: 8000-8999 range
  - Peer 0: 8051, 8052, 8053, 10443
  - Peer 1: 8151, 8152, 8153, 10543
- Organization 3: 9000-9999 range
  - Peer 0: 9051, 9052, 9053, 11443

**Orderer Ports:**
- Organization 1: 17000-17999 range
  - Orderer 0: 17050, 17053, 17443
  - Orderer 1: 17150, 17153, 17543
  - Orderer 2: 17250, 17253, 17643
- Organization 2: 18000-18999 range
  - Orderer 0: 18050, 18053, 18443

### Peer Configuration Fields

Each peer requires:

```hcl
peer0 = {
  name                      = "peer0-org1"              # Unique peer name
  mode                      = "service"                 # Deployment mode (service/container)
  version                   = "3.1.2"                   # Fabric peer version
  external_endpoint         = "localhost:7051"          # External access endpoint
  listen_address            = "0.0.0.0:7051"           # Listen address (peer-to-peer)
  chaincode_address         = "0.0.0.0:7052"           # Chaincode endpoint
  events_address            = "0.0.0.0:7053"           # Events endpoint
  operations_listen_address = "0.0.0.0:9443"           # Operations/metrics endpoint
  domain_names              = ["peer0-org1", "localhost"] # Optional: TLS cert SANs
  certificate_expiration    = 365                       # Optional: cert validity (days)
  auto_renewal_enabled      = true                      # Optional: enable auto-renewal
  auto_renewal_days         = 30                        # Optional: renewal threshold
}
```

### Orderer Configuration Fields

Each orderer requires:

```hcl
orderer0 = {
  name                      = "orderer0-ordererorg"     # Unique orderer name
  mode                      = "service"                 # Deployment mode
  version                   = "3.1.1"                   # Fabric orderer version
  external_endpoint         = "localhost:17050"         # External access endpoint
  listen_address            = "0.0.0.0:17050"          # Listen address (consensus)
  admin_address             = "0.0.0.0:17053"          # Admin endpoint
  operations_listen_address = "0.0.0.0:17443"          # Operations/metrics endpoint
  domain_names              = ["orderer0-ordererorg", "localhost"] # Optional: TLS SANs
  certificate_expiration    = 365                       # Optional: cert validity
  auto_renewal_enabled      = true                      # Optional: enable auto-renewal
}
```

## Advanced Configuration

### Custom Anchor Peers

By default, the first peer (peer0) of each organization becomes the anchor peer. To customize:

```hcl
peer_organizations = {
  org1 = {
    msp_id = "Org1MSP"
    peers = { peer0 = {...}, peer1 = {...}, peer2 = {...} }
    anchor_peer_indices = [0, 2]  # peer0 and peer2 are anchors
  }
}
```

### Custom Environment Variables

Add custom environment variables to any peer or orderer:

```hcl
peer0 = {
  # ... other config ...
  environment = {
    CORE_PEER_GOSSIP_USELEADERELECTION = "true"
    CORE_PEER_GOSSIP_ORGLEADER         = "false"
    CORE_PEER_PROFILE_ENABLED          = "true"
    FABRIC_LOGGING_SPEC                = "DEBUG"  # Custom log level
    CORE_CHAINCODE_EXECUTETIMEOUT      = "300s"   # Custom timeout
  }
}
```

### Etcd Raft Tuning

Adjust Raft consensus parameters:

```hcl
etcdraft_options = {
  tick_interval          = "250ms"    # Faster heartbeats
  election_tick          = 20         # Higher election timeout
  heartbeat_tick         = 2          # More frequent heartbeats
  max_inflight_blocks    = 10         # More blocks in flight
  snapshot_interval_size = 41943040   # 40MB snapshots
}
```

## Outputs

After deployment, the following information is available:

### Channel Information
- `channel_id`: Channel resource ID
- `channel_name`: Channel name
- `channel_status`: Channel operational status

### Organization Information
- `peer_organizations`: All created peer organizations (ID, name, MSP ID)
- `orderer_organizations`: All created orderer organizations

### Node Information
- `peers`: All peer nodes with endpoints and status
- `orderers`: All orderer nodes with endpoints and status
- `anchor_peers`: Anchor peer configuration per organization

### Example Output

```bash
$ terraform output peers

{
  "org1-peer0" = {
    "id" = "42"
    "name" = "peer0-org1"
    "status" = "running"
    "external_endpoint" = "localhost:7051"
    "msp_id" = "Org1MSP"
  }
  "org1-peer1" = { ... }
  "org2-peer0" = { ... }
}
```

## Network Operations

### Adding a New Peer Organization

1. Add the organization to `peer_organizations` in `terraform.tfvars`:

```hcl
peer_organizations = {
  # ... existing orgs ...
  org3 = {
    msp_id = "Org3MSP"
    description = "Third peer organization"
    peers = {
      peer0 = {
        name = "peer0-org3"
        # ... peer configuration ...
      }
    }
    anchor_peer_indices = [0]
  }
}
```

2. Apply the changes:

```bash
terraform apply
```

### Adding More Peers to Existing Organization

1. Add the peer to the organization's `peers` map:

```hcl
org1 = {
  msp_id = "Org1MSP"
  peers = {
    peer0 = { ... }
    peer1 = { ... }
    peer2 = {  # New peer
      name = "peer2-org1"
      # ... configuration ...
    }
  }
}
```

2. Apply the changes:

```bash
terraform apply
```

### Scaling Orderers

To add more consenters for better fault tolerance:

```hcl
orderer_organizations = {
  orderer_org = {
    msp_id = "OrdererOrgMSP"
    orderers = {
      orderer0 = { ... }
      orderer1 = { ... }
      orderer2 = { ... }
      orderer3 = { ... }  # Add 4th orderer
      orderer4 = { ... }  # Add 5th orderer
    }
  }
}
```

**Note:** Raft consensus requires odd numbers of orderers (1, 3, 5, 7) for optimal fault tolerance.

## Best Practices

### Production Deployment

1. **Orderer Count**: Use at least 3 orderers for Raft consensus (tolerates 1 failure)
2. **Peer Count**: Deploy at least 2 peers per organization for high availability
3. **Anchor Peers**: Set at least one anchor peer per organization
4. **Monitoring**: Use operations endpoints (9443, 17443) for metrics
5. **Certificates**: Enable auto-renewal for production deployments
6. **Versions**: Use stable Fabric versions (3.1.x)

### Development/Testing

1. **Minimal Setup**: 2 orgs, 1 peer each, 1 orderer (faster deployment)
2. **Local Endpoints**: Use `localhost` for all external endpoints
3. **Logs**: Set `FABRIC_LOGGING_SPEC=DEBUG` for troubleshooting
4. **Port Conflicts**: Ensure no other services use the same ports

### Network Planning

1. **Plan Port Allocation**: Reserve port ranges per organization
2. **DNS/Hostnames**: Use consistent naming conventions
3. **External Access**: Configure `external_endpoint` for cross-host access
4. **Organization Design**: Consider business network structure
5. **Scalability**: Start small, scale horizontally as needed

## Troubleshooting

### Issue: Port Already in Use

**Error:** `bind: address already in use`

**Solution:** Change the port numbers in your peer/orderer configuration:
```hcl
listen_address = "0.0.0.0:7151"  # Use different port
```

### Issue: Peer/Orderer Not Starting

**Check the node status:**
```bash
terraform output peers
terraform output orderers
```

**Common causes:**
- Port conflicts
- Certificate issues
- Network connectivity
- Insufficient resources

### Issue: Channel Creation Fails

**Verify orderers are running:**
```bash
terraform output orderers
```

**Ensure at least one orderer is healthy before channel creation.**

### Issue: Peer Cannot Join Channel

**Check dependencies:**
- Channel must be created first
- Peer must be in `peer_organizations` list
- Organization must be part of channel config

### Issue: Anchor Peer Configuration Fails

**Verify:**
- Peers are joined to channel first
- `anchor_peer_indices` reference valid peer indices (0-based)
- At least one peer exists in the organization

## Maintenance

### Updating Node Versions

1. Update `version` in `terraform.tfvars`:

```hcl
peer0 = {
  version = "3.1.3"  # Updated version
  # ... other config unchanged ...
}
```

2. Apply the change:

```bash
terraform apply
```

**Note:** Version updates may cause node restarts and brief downtime.

### Certificate Renewal

Certificates auto-renew if `auto_renewal_enabled = true`. Manual renewal:

1. The Chainlaunch platform handles certificate renewal automatically
2. Monitor certificate expiration via node status
3. Adjust `auto_renewal_days` to renew earlier (default: 30 days before expiration)

### Destroying the Network

**⚠️ Warning:** This will permanently delete all network resources.

```bash
terraform destroy
```

The destruction order is automatically handled by Terraform dependencies:
1. Anchor peers configuration
2. Node join operations
3. Channel
4. Peer nodes
5. Orderer nodes
6. Organizations

## Chaincode Deployment

This example includes **optional chaincode deployment** using the complete Fabric chaincode lifecycle. Set `deploy_chaincode = true` to enable.

### Chaincode Configuration

Each chaincode is configured with:

```hcl
chaincodes = {
  basic = {
    name               = "basic"                          # Chaincode name
    version            = "1.0"                            # Version
    sequence           = 1                                # Sequence number (increment for upgrades)
    docker_image       = "hyperledger/fabric-samples-basic:latest"  # Docker image
    chaincode_address  = "basic.example.com:7052"        # Service address
    endorsement_policy = ""                               # Custom policy (empty = default)
    install_on_orgs    = ["org1", "org2"]                # Install on these orgs
    approve_with_orgs  = ["org1", "org2"]                # Orgs that must approve
    commit_with_org    = "org1"                          # Org to commit
    environment_variables = {                             # Environment variables
      CORE_CHAINCODE_LOGGING_LEVEL = "info"
    }
  }
}
```

### Chaincode Deployment Workflow

When `deploy_chaincode = true`, the following steps are executed automatically:

1. **Create Chaincode Resource** - Logical chaincode entity
2. **Define Chaincode** - Version, sequence, docker image, endorsement policy
3. **Install** - Pull docker image to all peers in specified organizations
4. **Approve** - Each organization approves the definition
5. **Commit** - Commit the definition to the channel
6. **Deploy** - Start chaincode containers

### Example: Deploy Single Chaincode

```hcl
deploy_chaincode = true

chaincodes = {
  basic = {
    name               = "basic"
    version            = "1.0"
    sequence           = 1
    docker_image       = "hyperledger/fabric-samples-basic:latest"
    chaincode_address  = "basic.example.com:7052"
    endorsement_policy = ""
    install_on_orgs    = ["org1", "org2"]
    approve_with_orgs  = ["org1", "org2"]
    commit_with_org    = "org1"
  }
}
```

### Example: Deploy Multiple Chaincodes

```hcl
deploy_chaincode = true

chaincodes = {
  basic = {
    name               = "basic"
    version            = "1.0"
    sequence           = 1
    docker_image       = "hyperledger/fabric-samples-basic:latest"
    chaincode_address  = "basic.example.com:7052"
    endorsement_policy = ""
    install_on_orgs    = ["org1", "org2"]
    approve_with_orgs  = ["org1", "org2"]
    commit_with_org    = "org1"
  }

  asset_transfer = {
    name               = "asset-transfer"
    version            = "1.0"
    sequence           = 1
    docker_image       = "hyperledger/fabric-samples-asset-transfer:latest"
    chaincode_address  = "asset-transfer.example.com:7052"
    endorsement_policy = "OR('Org1MSP.peer','Org2MSP.peer')"
    install_on_orgs    = ["org1", "org2"]
    approve_with_orgs  = ["org1", "org2"]
    commit_with_org    = "org2"
  }
}
```

### Chaincode Upgrades

To upgrade a chaincode, increment the `version` and `sequence`:

```hcl
chaincodes = {
  basic = {
    name               = "basic"
    version            = "2.0"           # New version
    sequence           = 2               # Incremented sequence
    docker_image       = "hyperledger/fabric-samples-basic:v2.0"
    chaincode_address  = "basic.example.com:7052"
    endorsement_policy = ""
    install_on_orgs    = ["org1", "org2"]
    approve_with_orgs  = ["org1", "org2"]
    commit_with_org    = "org1"
  }
}
```

### Custom Endorsement Policies

Specify custom endorsement policies:

```hcl
# Any peer from either org can endorse
endorsement_policy = "OR('Org1MSP.peer','Org2MSP.peer')"

# Both orgs must endorse
endorsement_policy = "AND('Org1MSP.peer','Org2MSP.peer')"

# At least 2 out of 3 orgs must endorse
endorsement_policy = "OutOf(2, 'Org1MSP.peer', 'Org2MSP.peer', 'Org3MSP.peer')"

# Default policy (empty string)
endorsement_policy = ""
```

### Chaincode Environment Variables

Configure chaincode runtime environment:

```hcl
environment_variables = {
  CORE_CHAINCODE_LOGGING_LEVEL = "debug"
  CORE_CHAINCODE_LOGGING_SHIM  = "warning"
  CORE_CHAINCODE_EXECUTETIMEOUT = "300s"
  CUSTOM_ENV_VAR                = "custom_value"
}
```

### Viewing Chaincode Information

After deployment, view chaincode details:

```bash
terraform output chaincodes
```

Output example:
```json
{
  "basic" = {
    "id" = "42"
    "name" = "basic"
    "network_id" = "12"
    "definition" = {
      "id" = "84"
      "version" = "1.0"
      "sequence" = 1
      "docker_image" = "hyperledger/fabric-samples-basic:latest"
      "chaincode_address" = "basic.example.com:7052"
    }
  }
}
```

## Next Steps

After deploying your network:

1. **Deploy Chaincode**: Set `deploy_chaincode = true` and configure your chaincodes
2. **Configure Backups**: See [examples/backup-with-minio](../backup-with-minio/)
3. **Setup Monitoring**: Use operations endpoints for Prometheus/Grafana
4. **Invoke Transactions**: Use Fabric SDKs or CLI to invoke chaincode functions

## Resources Created

This example creates the following resources:

### Without Chaincode (default)

| Resource Type | Count | Description |
|---------------|-------|-------------|
| Organizations (Peer) | 2 | Peer organizations with MSP |
| Organizations (Orderer) | 1 | Orderer organization with MSP |
| Peers | 4 | Peer nodes (2 per org) |
| Orderers | 3 | Orderer nodes (consenters) |
| Fabric Network | 1 | Channel configuration |
| Join Operations | 7 | 4 peer joins + 3 orderer joins |
| Anchor Peer Configs | 2 | 1 per peer org |

**Total:** 20 Terraform resources

### With Chaincode Deployment (deploy_chaincode = true)

Add per chaincode:

| Resource Type | Count | Description |
|---------------|-------|-------------|
| Chaincode | 1 | Chaincode entity |
| Chaincode Definition | 1 | Version, sequence, docker image |
| Chaincode Install | 1 | Install operation (covers all peers) |
| Chaincode Approve | 2 | 1 per approving organization |
| Chaincode Commit | 1 | Commit to channel |
| Chaincode Deploy | 1 | Start containers |

**Per Chaincode:** 7 additional resources

**Example:** With 1 chaincode = 27 total resources, with 2 chaincodes = 34 total resources

## Key Concepts

### Organizations vs Nodes

- **Organization**: A logical entity with MSP (Membership Service Provider)
- **Node**: A running peer or orderer instance belonging to an organization
- Each organization can have multiple nodes
- Organizations are identified by their MSP ID

### Channel Join Operations

- **Logical Association**: Defined in `peer_organizations` and `orderer_organizations`
- **Physical Join**: Executed by `chainlaunch_fabric_join_node` resources
- **Two-Step Process**: First associate with network config, then perform join operation

### Anchor Peers

- **Purpose**: Facilitate cross-organization gossip communication
- **Requirement**: At least one anchor peer per organization
- **Best Practice**: Use highly available peers as anchors
- **Configuration**: Set via `anchor_peer_indices` (0-based array indices)

### Raft Consensus

- **Recommended**: For production use (fault tolerant)
- **Odd Numbers**: Use 1, 3, 5, or 7 orderers for optimal fault tolerance
- **Fault Tolerance**: N orderers tolerate (N-1)/2 failures
  - 3 orderers → 1 failure tolerance
  - 5 orderers → 2 failure tolerance
  - 7 orderers → 3 failure tolerance

## Support

For issues or questions:
- Review the [Chainlaunch documentation](https://docs.chainlaunch.io)
- Check [CLAUDE.md](../../CLAUDE.md) for provider architecture details
- Open an issue in the provider repository

## License

This example is part of the Chainlaunch Terraform Provider project.
