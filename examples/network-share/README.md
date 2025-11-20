# Network Sharing Example

This example demonstrates how to share Hyperledger Fabric and Besu networks with other Chainlaunch peer nodes using the `chainlaunch_network_share` resource.

## Overview

Network sharing allows you to distribute network configuration (genesis blocks, channel config, etc.) to other Chainlaunch instances that you've connected to via node invitations. This is essential for:

- **Multi-organization Fabric networks** - Share channel configuration with partner organizations
- **Besu consortium networks** - Distribute genesis configuration to consortium members
- **Cross-instance collaboration** - Enable multiple Chainlaunch deployments to participate in the same network

## Prerequisites

**Important**: Network sharing is a **Pro-only feature**. Both the sharing and receiving Chainlaunch instances must have Pro features enabled.

1. **Two or more Chainlaunch instances** running with Pro features
2. **Node invitations established** between instances (see `node-invitation` example)
3. **Connected peers** - Instances must have accepted invitations and be connected
4. **Network created locally** - The network you want to share must exist on your instance

## Workflow

### Instance A (Network Owner)

```hcl
# 1. Create network locally
resource "chainlaunch_fabric_network" "mychannel" {
  name = "mychannel"
  # ... network configuration
}

# 2. Share with connected peer
resource "chainlaunch_network_share" "share" {
  network_id   = chainlaunch_fabric_network.mychannel.id
  network_type = "fabric"
  recipients   = ["2"]  # Connection ID of Instance B

  metadata = {
    purpose = "collaboration"
  }
}
```

### Instance B (Receiving Peer)

On the receiving end, the shared network appears in the Pro interface where it can be:
- **Accepted** - Import the network and participate
- **Rejected** - Decline the shared network

## Finding Peer Node IDs (Connection IDs)

To share a network, you need the **connection ID** of the peer node. You can find this by:

1. **Via Chainlaunch UI**: Navigate to Pro → Connections to see all connected peers and their IDs
2. **Via API**: Query `/node/connected-peers` to get all connection IDs
3. **From node invitation**: The connection ID is established when accepting an invitation

Example of getting connected peers:

```bash
curl -u admin:admin123 http://localhost:8100/api/v1/node/connected-peers
```

Response:
```json
{
  "connected_peers": [
    {
      "connection_id": 2,
      "node_id": "node1",
      "address": "192.168.1.100:8100",
      "status": "connected"
    }
  ]
}
```

Use the `connection_id` (e.g., `"2"`) in the `recipients` list.

## Configuration

### Sharing a Fabric Network

```hcl
resource "chainlaunch_network_share" "share_fabric" {
  network_id   = chainlaunch_fabric_network.mychannel.id
  network_type = "fabric"

  recipients = [
    "2",  # Org2's Chainlaunch instance
    "3",  # Org3's Chainlaunch instance
  ]

  metadata = {
    purpose     = "supply-chain-channel"
    shared_date = "2024-01-15"
    contact     = "admin@org1.example.com"
  }
}
```

### Sharing a Besu Network

```hcl
resource "chainlaunch_network_share" "share_besu" {
  network_id   = chainlaunch_besu_network.consortium.id
  network_type = "besu"

  recipients = [
    "4",  # Partner 1
    "5",  # Partner 2
  ]

  metadata = {
    consortium = "energy-trading"
    version    = "1.0"
  }
}
```

## Resource Behavior

### Creation
- Sends network configuration (genesis block, etc.) to specified recipients
- Recipients receive notification about the shared network
- Share is created immediately

### Updates
- Changing `recipients` list will re-share the network
- `network_id` and `network_type` changes require resource replacement
- Metadata updates trigger re-sharing

### Deletion
- Removes share from Terraform state
- **Does NOT revoke** the share on recipient's end
- Recipients can still accept/reject previously sent shares

## Complete Multi-Instance Example

### Instance A (Org1) - Share Network

```hcl
provider "chainlaunch" {
  alias    = "org1"
  url      = "http://org1.example.com:8100"
  username = "admin"
  password = "admin123"
}

# Create network
resource "chainlaunch_fabric_network" "channel" {
  provider = chainlaunch.org1
  name     = "shared-channel"
  # ... configuration
}

# Share with Org2
resource "chainlaunch_network_share" "to_org2" {
  provider     = chainlaunch.org1
  network_id   = chainlaunch_fabric_network.channel.id
  network_type = "fabric"
  recipients   = ["2"]  # Org2's connection ID
}
```

### Instance B (Org2) - Accept Share

On the receiving instance, accept the share via:
- **Chainlaunch Pro UI**: Pro → Shared Networks → Accept
- **API Call**: `POST /pro/shared-networks/{shareId}/accept`

After accepting, the network is available for joining nodes.

## Troubleshooting

### Error: "Pro features not enabled"
**Solution**: Ensure both instances have Pro licensing enabled. Check with Chainlaunch support.

### Error: "Invalid recipient ID"
**Solution**:
1. Verify the connection ID exists: `GET /node/connected-peers`
2. Ensure peer invitation was accepted successfully
3. Check that the connection is active

### Share not appearing on recipient
**Solution**:
1. Verify network connectivity between instances
2. Check recipient's Pro → Shared Networks interface
3. Look for errors in recipient's logs
4. Ensure recipient has Pro features enabled

### Cannot update recipients
**Workaround**: The resource will re-share with new recipients. To remove recipients, update the list and apply.

## API Endpoints

- `POST /pro/sharing/network` - Share a Fabric network
- `POST /pro/sharing/besu-network` - Share a Besu network
- `GET /pro/shared-networks` - List received shares
- `POST /pro/shared-networks/{shareId}/accept` - Accept a share
- `POST /pro/shared-networks/{shareId}/reject` - Reject a share
- `GET /node/connected-peers` - List connected peer nodes

## Security Considerations

1. **Trust Model**: Only share networks with trusted partners
2. **Genesis Block Security**: Network configuration includes genesis blocks which define network parameters
3. **Metadata**: Avoid including sensitive information in metadata fields
4. **Access Control**: Recipients can view network configuration after accepting
5. **Network Isolation**: Each network share is independent

## Best Practices

1. **Use Descriptive Metadata**: Include purpose, contact info, and version
2. **Verify Connection IDs**: Always confirm peer connections before sharing
3. **Document Shares**: Keep track of which networks are shared with whom
4. **Version Networks**: Use metadata to track network versions when updating
5. **Test in Dev First**: Test network sharing in development before production

## Related Resources

- `chainlaunch_node_invitation` - Create invitations to connect peers
- `chainlaunch_node_accept_invitation` - Accept peer invitations
- `chainlaunch_fabric_network` - Create Fabric networks/channels
- `chainlaunch_besu_network` - Create Besu networks
- `chainlaunch_sync_all_external_nodes` - Sync nodes from connected peers

## See Also

- [Node Invitation Example](../node-invitation/)
- [Fabric Network Example](../fabric-network-complete/)
- [Besu Network Example](../besu-network-complete/)
- [Chainlaunch Pro Documentation](https://docs.chainlaunch.dev/pro/)
