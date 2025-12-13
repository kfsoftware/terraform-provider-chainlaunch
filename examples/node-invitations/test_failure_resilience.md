# Testing Failure Resilience in sync_all_external_nodes

This document demonstrates how `chainlaunch_sync_all_external_nodes` continues syncing even when individual peers fail.

## Implementation Details

The resource is designed to be **failure-resilient**:

1. **Loops through ALL connected peers** - Uses `for _, peer := range connectedPeersResp.ConnectedPeers`
2. **Records failures without stopping** - When a peer sync fails, it:
   - Sets `success = false`
   - Records the error message
   - Sets all counts to 0 for that peer
   - **Continues to the next peer** (doesn't return or break)
3. **Always records results** - Every peer gets an entry in `sync_results`, whether it succeeded or failed
4. **Aggregates only successes** - Total counts only include peers that synced successfully

## Code Flow

```go
for _, peer := range connectedPeersResp.ConnectedPeers {
    syncBody, syncErr := r.client.DoRequest("POST", "/node/sync-external-nodes", syncReq)

    if syncErr != nil {
        // Record failure but CONTINUE to next peer
        result.Success = types.BoolValue(false)
        result.Error = types.StringValue(syncErr.Error())
        result.OrganizationsAdded = types.Int64Value(0)
        // ... set other counts to 0
    } else {
        // Handle success
        result.Success = types.BoolValue(true)
        // ... extract and aggregate counts
        totalOrgsAdded += orgsAdded
        totalPeersAdded += peersAdded
        // ...
    }

    // ALWAYS append result (success or failure)
    peerNodeIDs = append(peerNodeIDs, peer.NodeID)
    syncResults = append(syncResults, result)
    // Loop continues to next peer automatically
}
```

## Example Scenario

### Setup
- Node 1 connected to 3 peers: `peer1`, `peer2`, `peer3`
- `peer1` is offline (will fail)
- `peer2` is online and has 5 nodes
- `peer3` is online and has 3 nodes

### Expected Behavior

The resource will:
1. Try to sync from `peer1` → **Fails** (network error)
2. Try to sync from `peer2` → **Succeeds** (5 nodes synced)
3. Try to sync from `peer3` → **Succeeds** (3 nodes synced)

### Result State

```hcl
resource "chainlaunch_sync_all_external_nodes" "all" {
  # After apply, the state will show:

  peer_node_ids = ["peer1", "peer2", "peer3"]  # All 3 peers listed

  # Aggregated totals (only successful syncs)
  fabric_peers_added = 8  # 0 from peer1 + 5 from peer2 + 3 from peer3

  # Per-peer details
  sync_results = [
    {
      peer_node_id = "peer1"
      success      = false
      error        = "connection refused"
      fabric_peers_added = 0
    },
    {
      peer_node_id = "peer2"
      success      = true
      error        = ""
      fabric_peers_added = 5
    },
    {
      peer_node_id = "peer3"
      success      = true
      error        = ""
      fabric_peers_added = 3
    }
  ]
}
```

## Testing This Behavior

### Test Case 1: One Peer Offline

```terraform
# Set up 3 peer connections
resource "chainlaunch_node_accept_invitation" "peer1" {
  invitation_token = var.peer1_token
}

resource "chainlaunch_node_accept_invitation" "peer2" {
  invitation_token = var.peer2_token
}

resource "chainlaunch_node_accept_invitation" "peer3" {
  invitation_token = var.peer3_token
}

# Sync from all peers
resource "chainlaunch_sync_all_external_nodes" "all" {
  depends_on = [
    chainlaunch_node_accept_invitation.peer1,
    chainlaunch_node_accept_invitation.peer2,
    chainlaunch_node_accept_invitation.peer3,
  ]
}

# Check which peers failed
output "failed_peers" {
  value = [
    for result in chainlaunch_sync_all_external_nodes.all.sync_results :
    {
      peer  = result.peer_node_id
      error = result.error
    } if !result.success
  ]
}

# Check which peers succeeded
output "successful_peers" {
  value = [
    for result in chainlaunch_sync_all_external_nodes.all.sync_results :
    {
      peer         = result.peer_node_id
      peers_added  = result.fabric_peers_added
      nodes_added  = result.organizations_added
    } if result.success
  ]
}

# Total nodes synced (across successful peers only)
output "total_synced" {
  value = {
    organizations = chainlaunch_sync_all_external_nodes.all.organizations_added
    peers         = chainlaunch_sync_all_external_nodes.all.fabric_peers_added
    orderers      = chainlaunch_sync_all_external_nodes.all.fabric_orderers_added
  }
}
```

### Expected Output (with peer1 offline)

```
Outputs:

failed_peers = [
  {
    peer  = "peer1"
    error = "Post \"http://peer1:8100/node/sync-external-nodes\": dial tcp: connection refused"
  }
]

successful_peers = [
  {
    peer        = "peer2"
    peers_added = 5
    nodes_added = 2
  },
  {
    peer        = "peer3"
    peers_added = 3
    nodes_added = 1
  }
]

total_synced = {
  organizations = 3
  peers         = 8
  orderers      = 0
}
```

## Benefits of This Design

1. **Partial Success is Useful** - Even if some peers are down, you still get nodes from available peers
2. **Visibility** - `sync_results` shows exactly which peers failed and why
3. **Non-Blocking** - One failed peer doesn't block syncing from other healthy peers
4. **Idempotent** - Running again will retry failed peers automatically
5. **Diagnostic Info** - Error messages help troubleshoot connection issues

## Comparison with Individual Syncs

### Using `chainlaunch_external_nodes_sync` (manual per-peer)
```terraform
# If peer1 fails, Terraform stops and doesn't try peer2 or peer3
resource "chainlaunch_external_nodes_sync" "peer1" {
  peer_node_id = "peer1"  # ❌ FAILS - stops here
}

resource "chainlaunch_external_nodes_sync" "peer2" {
  peer_node_id = "peer2"  # ⏸️  Never reaches this
}
```

### Using `chainlaunch_sync_all_external_nodes` (automatic)
```terraform
# If peer1 fails, continues to peer2 and peer3
resource "chainlaunch_sync_all_external_nodes" "all" {
  # ✅ peer1 fails → recorded in sync_results
  # ✅ peer2 succeeds → 5 nodes synced
  # ✅ peer3 succeeds → 3 nodes synced
  # Result: 8 total nodes synced from 2/3 peers
}
```

## When to Use This Resource

✅ **Use `chainlaunch_sync_all_external_nodes` when:**
- You have multiple peer connections
- You want to sync from all available peers
- Partial failures are acceptable (sync what you can)
- You want automatic retry on next apply

❌ **Use `chainlaunch_external_nodes_sync` when:**
- You need to sync from a specific peer only
- You want the apply to fail if that specific peer is unavailable
- You need fine-grained control over sync timing per peer
