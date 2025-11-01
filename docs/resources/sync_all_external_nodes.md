---
page_title: "chainlaunch_sync_all_external_nodes Resource - chainlaunch"
subcategory: "Node Sharing"
description: |-
  Automatically synchronizes external nodes from ALL connected remote Chainlaunch instances with failure resilience.
---

# chainlaunch_sync_all_external_nodes (Resource)

Automatically synchronizes external nodes from ALL connected remote Chainlaunch instances.

This resource discovers all connected peers and syncs their nodes automatically - no need to specify individual `peer_node_id`s. It always runs on every terraform apply to keep external nodes synchronized.

## Key Features

- **Automatic Discovery**: Discovers all connected peer nodes automatically
- **Parallel Sync**: Syncs nodes from each connected peer in parallel
- **Failure Resilient**: **Continues syncing even if individual peers fail** - partial success is useful!
- **Aggregated Results**: Tracks total nodes added/deleted across all successful syncs
- **Per-Peer Details**: Provides detailed sync results for each peer with success/error status
- **Error Visibility**: See exactly which peers failed and why

## Failure Resilience

**IMPORTANT**: This resource is designed to continue syncing from all peers even if some fail.

If you have 3 connected peers and peer1 is offline:
- ❌ peer1 sync fails → error recorded in `sync_results`
- ✅ peer2 sync succeeds → nodes imported
- ✅ peer3 sync succeeds → nodes imported
- **Result**: Partial success with 2/3 peers synced

The resource **never stops** at the first failure - it attempts to sync from ALL connected peers and reports individual results.

## Use Cases

Use this resource when:
- You want to sync from all connected peers automatically
- You don't want to manually specify each `peer_node_id`
- You have multiple peer connections and want to keep all external nodes in sync
- **Partial failures are acceptable** - sync what you can from available peers

Use `chainlaunch_external_nodes_sync` instead when:
- You want to sync from a specific peer only
- You need the apply to fail if that specific peer is unavailable
- You need fine-grained control over which peers to sync from

## Example Usage

### Basic Usage - Sync All Connected Peers

```terraform
# Accept node invitations from multiple peers
resource "chainlaunch_node_accept_invitation" "node1" {
  invitation_token = "token-from-node1"
}

resource "chainlaunch_node_accept_invitation" "node2" {
  invitation_token = "token-from-node2"
}

# Automatically sync from ALL connected peers
resource "chainlaunch_sync_all_external_nodes" "all_peers" {
  # No configuration needed - discovers peers automatically

  depends_on = [
    chainlaunch_node_accept_invitation.node1,
    chainlaunch_node_accept_invitation.node2,
  ]
}

# Output sync results
output "total_peers_synced" {
  value = length(chainlaunch_sync_all_external_nodes.all_peers.peer_node_ids)
}

output "total_fabric_peers_added" {
  value = chainlaunch_sync_all_external_nodes.all_peers.fabric_peers_added
}

# Check which peers failed (if any)
output "failed_peers" {
  value = [
    for result in chainlaunch_sync_all_external_nodes.all_peers.sync_results :
    {
      peer  = result.peer_node_id
      error = result.error
    } if !result.success
  ]
}

# Check which peers succeeded
output "successful_peers" {
  value = [
    for result in chainlaunch_sync_all_external_nodes.all_peers.sync_results :
    {
      peer        = result.peer_node_id
      peers_added = result.fabric_peers_added
    } if result.success
  ]
}
```

### Multi-Provider Setup

```terraform
provider "chainlaunch" {
  alias    = "node1"
  url      = "http://node1.example.com:8100"
  username = "admin"
  password = "admin123"
}

provider "chainlaunch" {
  alias    = "node2"
  url      = "http://node2.example.com:8100"
  username = "admin"
  password = "admin123"
}

# On Node 1: Accept invitations from other nodes
resource "chainlaunch_node_accept_invitation" "node1_connections" {
  provider = chainlaunch.node1

  invitation_token = var.invitation_from_node2
}

# On Node 1: Sync ALL external nodes from connected peers
resource "chainlaunch_sync_all_external_nodes" "node1_sync" {
  provider = chainlaunch.node1

  depends_on = [chainlaunch_node_accept_invitation.node1_connections]
}

# On Node 2: Accept invitations from other nodes
resource "chainlaunch_node_accept_invitation" "node2_connections" {
  provider = chainlaunch.node2

  invitation_token = var.invitation_from_node1
}

# On Node 2: Sync ALL external nodes from connected peers
resource "chainlaunch_sync_all_external_nodes" "node2_sync" {
  provider = chainlaunch.node2

  depends_on = [chainlaunch_node_accept_invitation.node2_connections]
}
```

## How It Works

1. **Discovery Phase**: Calls `/node/connected-peers` to get all connected peer node IDs
2. **Sync Phase**: For each connected peer, calls `/node/sync-external-nodes` with the peer's node ID
   - **Continues on failure** - if peer1 fails, still tries peer2, peer3, etc.
3. **Aggregation Phase**: Aggregates results across all peers and tracks per-peer sync status
4. **State Update**: Updates Terraform state with totals and detailed per-peer results

## Behavior

### When to Run
- Always runs on every `terraform apply` to keep nodes synchronized
- Runs on `terraform refresh` to detect drift

### Empty State Handling
If there are no connected peers, the resource creates an empty state with zero counts.

### Error Handling - FAILURE RESILIENT
- **If syncing from a specific peer fails**, the error is recorded in `sync_results`
- **Sync continues for remaining peers** - partial success is useful!
- **Total counts only include successful syncs**
- All peers appear in `sync_results` with their individual success/failure status

### Example Failure Scenario

With 3 connected peers where peer1 is offline:

```hcl
# After apply, state shows:
peer_node_ids = ["peer1", "peer2", "peer3"]

# Only successful peers contribute to totals
fabric_peers_added = 8  # 0 from peer1 + 5 from peer2 + 3 from peer3

# Per-peer details show what happened
sync_results = [
  {
    peer_node_id       = "peer1"
    success            = false
    error              = "connection refused"
    fabric_peers_added = 0
  },
  {
    peer_node_id       = "peer2"
    success            = true
    error              = ""
    fabric_peers_added = 5
  },
  {
    peer_node_id       = "peer3"
    success            = true
    error              = ""
    fabric_peers_added = 3
  }
]
```

### Deletion
When the resource is deleted from Terraform:
- External nodes remain in the system
- Only the Terraform state is removed
- No actual cleanup is performed

## Schema

### Read-Only Attributes

#### Aggregated Totals (Successful Syncs Only)
- `organizations_added` (Number) - Total organizations added across all successful syncs
- `fabric_peers_added` (Number) - Total Fabric peers added across all successful syncs
- `fabric_peers_deleted` (Number) - Total Fabric peers deleted across all successful syncs
- `fabric_orderers_added` (Number) - Total Fabric orderers added across all successful syncs
- `fabric_orderers_deleted` (Number) - Total Fabric orderers deleted across all successful syncs
- `besu_nodes_added` (Number) - Total Besu nodes added across all successful syncs
- `besu_nodes_deleted` (Number) - Total Besu nodes deleted across all successful syncs

#### Metadata
- `id` (String) - Resource identifier (timestamp of sync)
- `last_sync_at` (String) - Timestamp of the last successful sync
- `peer_node_ids` (List of String) - List of ALL peer node IDs that were attempted (includes both successes and failures)

#### Per-Peer Details
- `sync_results` (List of Objects) - Detailed sync results for each peer (both successes and failures)
  - `peer_node_id` (String) - The peer node ID that was synced
  - `success` (Boolean) - Whether the sync was successful for this peer
  - `error` (String) - Error message if sync failed (empty string if successful)
  - `organizations_added` (Number) - Organizations added from this peer (0 if failed)
  - `fabric_peers_added` (Number) - Fabric peers added from this peer (0 if failed)
  - `fabric_peers_deleted` (Number) - Fabric peers deleted from this peer (0 if failed)
  - `fabric_orderers_added` (Number) - Fabric orderers added from this peer (0 if failed)
  - `fabric_orderers_deleted` (Number) - Fabric orderers deleted from this peer (0 if failed)
  - `besu_nodes_added` (Number) - Besu nodes added from this peer (0 if failed)
  - `besu_nodes_deleted` (Number) - Besu nodes deleted from this peer (0 if failed)

## Comparison: sync_all_external_nodes vs external_nodes_sync

| Feature | chainlaunch_sync_all_external_nodes | chainlaunch_external_nodes_sync |
|---------|-------------------------------------|--------------------------------|
| Peer Discovery | ✅ Automatic | ❌ Manual (specify peer_node_id) |
| Configuration | ✅ No configuration needed | ❌ Requires peer_node_id |
| Use Case | Sync from all connected peers | Sync from specific peer only |
| Failure Handling | ✅ **Continues on peer failure** | ❌ Stops on any error |
| Per-Peer Details | ✅ Yes (in sync_results) | ❌ No |
| Aggregated Totals | ✅ Yes | N/A (single peer) |
| Partial Success | ✅ Useful - sync what you can | N/A |

## Troubleshooting

### No Peers Synced
If `peer_node_ids` is empty:
- Check that node invitations have been accepted
- Verify peer connections with `GET /node/connected-peers`
- Ensure `depends_on` includes invitation acceptance

### Partial Sync Failures
Check the `sync_results` attribute for per-peer error messages:
```terraform
output "sync_failures" {
  value = [
    for result in chainlaunch_sync_all_external_nodes.all_peers.sync_results :
    {
      peer  = result.peer_node_id
      error = result.error
    } if !result.success
  ]
}
```

### All Peers Failed
If all peers in `sync_results` show `success = false`:
- Check network connectivity to peer instances
- Verify peer instances are running and accessible
- Check authentication credentials
- Review error messages in each `sync_results[].error`

### Sync Always Runs
This is expected behavior - the resource intentionally runs on every apply to keep external nodes synchronized.

## See Also

- [chainlaunch_external_nodes_sync](external_nodes_sync.md) - Sync from a specific peer
- [chainlaunch_node_invitation](node_invitation.md) - Generate node invitations
- [chainlaunch_node_accept_invitation](node_accept_invitation.md) - Accept node invitations
- [External Organization Data Source](../data-sources/external_fabric_organizations.md) - Query synced organizations
- [External Peer Data Source](../data-sources/external_fabric_peers.md) - Query synced peers
- [External Orderer Data Source](../data-sources/external_fabric_orderers.md) - Query synced orderers
- [Failure Resilience Test Example](../../examples/node-invitations/test_failure_resilience.md) - Detailed test scenarios
