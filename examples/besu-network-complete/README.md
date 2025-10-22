# Complete Hyperledger Besu Network Example

This comprehensive example demonstrates how to deploy a complete Hyperledger Besu (Ethereum) network with cryptographic keys, network creation, and node deployment using Terraform variables.

## Overview

This example creates a production-ready Besu network with:
- **Cryptographic keys** (secp256k1 curve for Ethereum)
- **Besu network** with configurable consensus mechanism
- **Multiple Besu nodes** (configurable)
- **Bootnode configuration** for peer discovery
- **Metrics endpoints** for monitoring (Prometheus)
- **JWT authentication** (optional)
- **Complete network lifecycle management**

All resources are defined in a single `main.tf` file, with configuration managed through variables.

## Features

### ✅ Complete Key-to-Network Workflow
- Automatically creates secp256k1 keys for each node
- Creates Besu network with your chosen consensus mechanism
- Deploys nodes with proper key assignments
- Configures bootnode for peer discovery

### ✅ Flexible Consensus Options
- **QBFT** (Quantum Byzantine Fault Tolerant) - Recommended for production
- **IBFT2** (Istanbul Byzantine Fault Tolerance 2.0)
- **Clique** (Proof of Authority for development)

### ✅ Production-Ready Defaults
- Besu 24.5.1 (latest stable)
- Prometheus metrics enabled
- Proper network isolation (internal/external IPs)
- Environment variable customization
- JWT authentication support

### ✅ Fully Variable-Driven
- Add or remove nodes by modifying `nodes` variable
- Configure consensus mechanism
- Adjust ports, IPs, and endpoints
- Customize all node properties

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                   Hyperledger Besu Network                       │
│                      (Chain ID: 1337)                           │
│                    QBFT Consensus                               │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Node 0 (Bootnode)                                       │  │
│  │  ┌────────────┐                                          │  │
│  │  │ secp256k1  │  RPC: 8545    P2P: 30303               │  │
│  │  │    Key     │  Metrics: 9545                          │  │
│  │  └────────────┘                                          │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Node 1                                                  │  │
│  │  ┌────────────┐                                          │  │
│  │  │ secp256k1  │  RPC: 8546    P2P: 30304               │  │
│  │  │    Key     │  Metrics: 9546                          │  │
│  │  └────────────┘  Boot: node0                            │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Node 2                                                  │  │
│  │  ┌────────────┐                                          │  │
│  │  │ secp256k1  │  RPC: 8547    P2P: 30305               │  │
│  │  │    Key     │  Metrics: 9547                          │  │
│  │  └────────────┘  Boot: node0                            │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

1. **Chainlaunch instance** running and accessible
2. **Terraform** installed (>= 1.0)
3. **Chainlaunch Terraform Provider** built and configured
4. **Available ports** for RPC, P2P, and metrics endpoints
5. **Docker** (if using mode = "docker")

## Quick Start

### 1. Copy the Example Configuration

```bash
cd examples/besu-network-complete
cp terraform.tfvars.example terraform.tfvars
```

### 2. Customize Your Network

Edit `terraform.tfvars` to configure your network. The default configuration creates:
- **3 Besu nodes** (node0 as bootnode)
- **QBFT consensus** (Byzantine Fault Tolerant)
- **Chain ID 1337** (development)
- **Prometheus metrics** enabled
- **Docker deployment**

### 3. Review the Plan

```bash
terraform init
terraform plan
```

Expected resources:
- 3 cryptographic keys (secp256k1)
- 1 Besu network
- 3 Besu nodes

**Total:** 7 resources

### 4. Deploy the Network

```bash
terraform apply
```

This will:
1. Create secp256k1 keys for each node
2. Create the Besu network with QBFT consensus
3. Deploy all nodes
4. Configure node0 as bootnode
5. Connect remaining nodes to bootnode

### 5. Verify the Deployment

```bash
# View all outputs
terraform output

# View RPC endpoints
terraform output rpc_endpoints

# View node information
terraform output nodes

# Test RPC connection to node0
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

## Configuration Examples

### Example 1: 3-Node QBFT Network (Default)

**Recommended for production** - Byzantine Fault Tolerant consensus

```hcl
consensus_mechanism = "qbft"

nodes = {
  node0 = {
    external_ip = "localhost"
    internal_ip = "127.0.0.1"
    p2p_port    = 30303
    rpc_port    = 8545
    is_bootnode = true
  }
  node1 = { ... }
  node2 = { ... }
}
```

### Example 2: 4-Node IBFT2 Network

**Alternative BFT consensus** - Similar to QBFT

```hcl
consensus_mechanism = "ibft2"

nodes = {
  node0 = { is_bootnode = true, ... }
  node1 = { ... }
  node2 = { ... }
  node3 = { ... }
}
```

### Example 3: Minimal 2-Node Development

**Fast development setup** - Minimal resources

```hcl
consensus_mechanism = "qbft"

nodes = {
  node0 = {
    external_ip     = "localhost"
    internal_ip     = "127.0.0.1"
    p2p_port        = 30303
    rpc_port        = 8545
    is_bootnode     = true
    metrics_enabled = false
  }
  node1 = {
    external_ip     = "localhost"
    internal_ip     = "127.0.0.1"
    p2p_port        = 30304
    rpc_port        = 8546
    metrics_enabled = false
  }
}
```

### Example 4: Production Network with JWT

**Secure production deployment** with JWT authentication

```hcl
nodes = {
  node0 = {
    external_ip   = "192.168.1.10"
    jwt_enabled   = true
    jwt_algorithm = "RS256"
    is_bootnode   = true
    ...
  }
  # ... more nodes with JWT
}
```

## Variable Configuration

### Key Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `network_name` | Name of the Besu network | `besu-dev-network` |
| `chain_id` | Ethereum chain ID | `1337` |
| `consensus_mechanism` | Consensus algorithm | `qbft` |
| `block_period` | Block time in seconds | `2` |
| `epoch_length` | Blocks before vote reset | `30000` |
| `request_timeout` | Consensus round timeout (seconds) | `10` |
| `external_ip` | Default external IP for all nodes | `localhost` |
| `internal_ip` | Default internal IP for all nodes | `127.0.0.1` |
| `nodes` | Map of node configurations | 3 nodes |
| `key_provider_type` | Key storage backend | `database` |

### Node Configuration Fields

Each node requires:

```hcl
node0 = {
  external_ip      = "localhost"        # External access IP (optional - uses var.external_ip if not set)
  internal_ip      = "127.0.0.1"        # Internal IP for node (optional - uses var.internal_ip if not set)
  p2p_host         = "0.0.0.0"          # P2P listen address
  p2p_port         = 30303              # P2P port (peer discovery)
  rpc_host         = "0.0.0.0"          # RPC listen address
  rpc_port         = 8545               # RPC port (JSON-RPC API)
  mode             = "docker"           # Deployment mode
  version          = "24.5.1"           # Besu version (optional)
  min_gas_price    = 1000               # Minimum gas price in Wei (optional)
  metrics_enabled  = true               # Enable Prometheus metrics (optional)
  metrics_port     = 9545               # Metrics endpoint port (optional)
  metrics_protocol = "prometheus"       # Metrics protocol (optional)
  host_allow_list  = "*"                # RPC host whitelist (optional)
  jwt_enabled      = false              # Enable JWT auth (optional)
  jwt_algorithm    = "RS256"            # JWT algorithm (optional)
  is_bootnode      = true               # Mark as bootnode (optional)
  environment      = {                  # Environment variables (optional)
    BESU_LOGGING = "INFO"
  }
}
```

### Global IP Configuration

**Single-Host Deployment** (default):
```hcl
external_ip = "localhost"     # All nodes accessible via localhost
internal_ip = "127.0.0.1"     # All nodes bind to loopback

# Nodes inherit these values - no need to specify per node
nodes = {
  node0 = { p2p_port = 30303, rpc_port = 8545, ... }
  node1 = { p2p_port = 30304, rpc_port = 8546, ... }
}
```

**Multi-Host Deployment** (override per node):
```hcl
external_ip = "192.168.1.10"  # Default (not used if all nodes override)
internal_ip = "192.168.1.10"  # Default (not used if all nodes override)

nodes = {
  node0 = {
    external_ip = "192.168.1.10"  # Override for this node
    internal_ip = "192.168.1.10"
    ...
  }
  node1 = {
    external_ip = "192.168.1.11"  # Different host
    internal_ip = "192.168.1.11"
    ...
  }
}
```

### Port Allocation Guide

**Standard Ports:**
- **RPC:** 8545, 8546, 8547, 8548...
- **P2P:** 30303, 30304, 30305, 30306...
- **Metrics:** 9545, 9546, 9547, 9548...

**Single-Host Setup:**
- Different ports per node (30303, 30304, 30305...)
- All nodes use same external_ip (localhost)

**Multi-Host Setup:**
- Same ports on each host (e.g., all nodes use 8545 for RPC)
- Different external_ip per node
- Differentiate nodes by IP address

### Consensus Parameters

**Block Period** - Time between blocks:
- **Default:** 2 seconds
- **Range:** 1-10 seconds
- **Impact:** Lower = faster finality, higher network overhead
- **Recommendation:** 2s for dev, 5s for production

**Epoch Length** - Blocks before vote reset:
- **Default:** 30000 blocks
- **Purpose:** Determines when validator votes reset
- **Impact:** Affects validator set changes
- **Recommendation:** Keep default (30000)

**Request Timeout** - Consensus round timeout:
- **Default:** 10 seconds
- **Purpose:** Max time to wait for consensus in each round
- **Impact:** Network resilience vs speed
- **Recommendation:** 10s (increase for slow networks)

Example configuration:
```hcl
block_period    = 2      # 2 second blocks
epoch_length    = 30000  # Reset votes every 30k blocks
request_timeout = 10     # 10 second timeout per round
```

## Consensus Mechanisms

### QBFT (Recommended)

**Quantum Byzantine Fault Tolerant** - Modern BFT consensus

- **Use Case:** Production networks, permissioned networks
- **Fault Tolerance:** (n-1)/3 Byzantine faults
- **Min Validators:** 4 (recommended)
- **Performance:** High throughput, low latency
- **Configuration:** `consensus_mechanism = "qbft"`

### IBFT2

**Istanbul Byzantine Fault Tolerance 2.0** - Classic BFT

- **Use Case:** Production networks, proven stability
- **Fault Tolerance:** (n-1)/3 Byzantine faults
- **Min Validators:** 4 (recommended)
- **Performance:** Similar to QBFT
- **Configuration:** `consensus_mechanism = "ibft2"`

### Clique

**Proof of Authority** - Simple PoA consensus

- **Use Case:** Development, testing
- **Fault Tolerance:** Limited
- **Min Validators:** 1
- **Performance:** Fast for development
- **Configuration:** `consensus_mechanism = "clique"`

## Cryptographic Keys

### secp256k1 Curve

Besu uses the **secp256k1** elliptic curve (same as Bitcoin and Ethereum mainnet):

- **Algorithm:** EC (Elliptic Curve)
- **Curve:** secp256k1
- **Use:** Node identity, block signing, transaction signing
- **Provider Support:**
  - ✅ Database (default)
  - ✅ AWS KMS
  - ❌ HashiCorp Vault (does not support secp256k1)

### Key Provider Selection

```hcl
# Option 1: Database (default)
key_provider_type = "database"

# Option 2: AWS KMS (production)
key_provider_type = "awskms"
```

## Network Operations

### Adding a New Node

1. Add the node to `nodes` variable in `terraform.tfvars`:

```hcl
nodes = {
  # ... existing nodes ...
  node3 = {
    external_ip      = "localhost"
    internal_ip      = "127.0.0.1"
    p2p_port         = 30306
    rpc_port         = 8548
    mode             = "docker"
    metrics_enabled  = true
    metrics_port     = 9548
  }
}
```

2. Apply the changes:

```bash
terraform apply
```

Terraform will:
- Create a new secp256k1 key for node3
- Deploy node3
- Automatically configure it to connect to the bootnode

### Changing Consensus Mechanism

⚠️ **Warning:** Changing consensus requires recreating the network.

1. Update `consensus_mechanism` in `terraform.tfvars`
2. Run `terraform apply` - this will destroy and recreate the network

### Scaling the Network

**Horizontal Scaling** - Add more nodes:
```hcl
# Simply add more node entries
nodes = {
  node0 = { ... }
  node1 = { ... }
  node2 = { ... }
  node3 = { ... }  # New node
  node4 = { ... }  # New node
}
```

## Outputs

After deployment, the following information is available:

### Network Information
- `network_id`: Network resource ID
- `network_name`: Network name
- `network_status`: Network status
- `network_config`: Full network configuration (chain ID, consensus)

### Key Information
- `node_keys`: All created secp256k1 keys (ID, name, algorithm, curve)

### Node Information
- `nodes`: All nodes with IDs, names, status, endpoints
- `bootnode_info`: Information about designated bootnodes
- `rpc_endpoints`: HTTP RPC endpoints for all nodes
- `metrics_endpoints`: Prometheus metrics endpoints

### Example Output

```bash
$ terraform output rpc_endpoints

{
  "node0" = "http://localhost:8545"
  "node1" = "http://localhost:8546"
  "node2" = "http://localhost:8547"
}

$ terraform output nodes

{
  "node0" = {
    "id" = "42"
    "name" = "besu-dev-network-node0"
    "status" = "RUNNING"
    "external_ip" = "localhost"
    "rpc_port" = 8545
    "p2p_port" = 30303
    "rpc_endpoint" = "http://localhost:8545"
  }
  ...
}
```

## Testing Your Network

### 1. Check Network Health

```bash
# Get block number
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Get peer count
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}'

# Get node info
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}'
```

### 2. Check Metrics

```bash
# Prometheus metrics
curl http://localhost:9545/metrics

# Check specific metrics
curl http://localhost:9545/metrics | grep besu_blockchain_height
```

### 3. Send a Transaction

```bash
# Get accounts
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1"}'

# Send transaction (requires funded account)
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "method":"eth_sendTransaction",
    "params":[{
      "from": "0x...",
      "to": "0x...",
      "value": "0x1"
    }],
    "id":1
  }'
```

## Monitoring

### Prometheus Configuration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'besu-node0'
    static_configs:
      - targets: ['localhost:9545']
  - job_name: 'besu-node1'
    static_configs:
      - targets: ['localhost:9546']
  - job_name: 'besu-node2'
    static_configs:
      - targets: ['localhost:9547']
```

### Key Metrics to Monitor

- `besu_blockchain_height` - Current block height
- `besu_peers_connected_total` - Number of connected peers
- `besu_transaction_pool_transactions` - Pending transactions
- `besu_synchronizer_in_sync` - Sync status

## Troubleshooting

### Issue: Node Not Starting

**Check node status:**
```bash
terraform output nodes
```

**Common causes:**
- Port conflicts
- Insufficient resources (CPU/memory)
- Docker not running
- Network connectivity issues

### Issue: Nodes Not Connecting

**Verify bootnode:**
```bash
terraform output bootnode_info
```

**Check P2P connectivity:**
```bash
# Test P2P port
nc -zv localhost 30303
```

### Issue: RPC Not Responding

**Verify RPC endpoints:**
```bash
terraform output rpc_endpoints

# Test connection
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}'
```

### Issue: Key Creation Fails

**Check key provider:**
```bash
# Verify provider supports secp256k1
terraform plan

# If using Vault, switch to database or AWS KMS
key_provider_type = "database"
```

## Best Practices

### Production Deployment

1. **Node Count:** Use at least 4 nodes for QBFT/IBFT2 (tolerates 1 Byzantine failure)
2. **Consensus:** Use QBFT or IBFT2 for production (not Clique)
3. **Monitoring:** Enable Prometheus metrics for all nodes
4. **Security:** Enable JWT authentication for Engine API
5. **Key Management:** Use AWS KMS for production key storage
6. **Networking:** Use proper firewalls and network isolation
7. **Backups:** Regular backups of blockchain data

### Development/Testing

1. **Minimal Setup:** 2-3 nodes for faster deployment
2. **Local Endpoints:** Use `localhost` for all IPs
3. **Disable Metrics:** Set `metrics_enabled = false` to reduce overhead
4. **Lower Gas Price:** Set `min_gas_price = 0` for free transactions

### Network Planning

1. **Chain ID:** Choose unique chain ID (avoid conflicts with public networks)
2. **Port Allocation:** Plan port ranges per host
3. **DNS/Hostnames:** Use consistent naming for multi-host deployments
4. **Scalability:** Start small, add nodes as needed

## Maintenance

### Updating Besu Version

1. Update `version` in `terraform.tfvars`:

```hcl
node0 = {
  version = "24.7.0"  # New version
  ...
}
```

2. Apply the change:

```bash
terraform apply
```

**Note:** Version updates cause node restarts.

### Destroying the Network

**⚠️ Warning:** This will permanently delete all network resources and blockchain data.

```bash
terraform destroy
```

Destruction order:
1. Besu nodes
2. Besu network
3. Cryptographic keys

## Resources Created

| Resource Type | Count | Description |
|---------------|-------|-------------|
| Cryptographic Keys | 3 | secp256k1 keys for node identity |
| Besu Network | 1 | Network configuration |
| Besu Nodes | 3 | Validator nodes |

**Total:** 7 Terraform resources

## Next Steps

After deploying your Besu network:

1. **Deploy Smart Contracts:** Use Hardhat, Truffle, or Remix
2. **Connect MetaMask:** Configure custom RPC (http://localhost:8545)
3. **Setup Monitoring:** Configure Prometheus + Grafana
4. **Integrate Applications:** Use Web3.js or ethers.js
5. **Configure Backups:** Regular blockchain data backups

## Additional Resources

- [Hyperledger Besu Documentation](https://besu.hyperledger.org/)
- [QBFT Consensus](https://besu.hyperledger.org/private-networks/how-to/configure/consensus/qbft)
- [IBFT2 Consensus](https://besu.hyperledger.org/private-networks/how-to/configure/consensus/ibft)
- [Besu JSON-RPC API](https://besu.hyperledger.org/public-networks/reference/api)
- [secp256k1 Curve](https://en.bitcoin.it/wiki/Secp256k1)

## Support

For issues or questions:
- Review the [Chainlaunch documentation](https://docs.chainlaunch.io)
- Check [CLAUDE.md](../../CLAUDE.md) for provider architecture details
- Open an issue in the provider repository

## License

This example is part of the Chainlaunch Terraform Provider project.
