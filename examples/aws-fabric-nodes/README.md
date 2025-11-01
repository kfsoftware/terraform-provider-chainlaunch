# AWS + Chainlaunch Fabric Network Example

This example demonstrates how to deploy a Hyperledger Fabric network using **AWS EC2 instances** for infrastructure and **Chainlaunch** for blockchain node management.

## Architecture

The example provisions:
1. **AWS Infrastructure**:
   - VPC with public subnet
   - Internet Gateway and routing
   - Security groups for peer and orderer nodes
   - EC2 instances with Docker pre-installed
   - **Optional**: Chainlaunch management server (auto-installed via SSH)

2. **Chainlaunch Resources**:
   - Fabric organizations (peer and orderer)
   - Peer nodes using AWS instance public IPs
   - Orderer nodes using AWS instance public IPs

## Key Features

- **Automatic Chainlaunch Installation**: Option to install Chainlaunch on AWS via SSH
- **Dynamic IP Assignment**: Chainlaunch nodes automatically use AWS EC2 public IPs
- **Address Mapping**: External IPs are mapped to internal private IPs for Docker networking
- **Complete Network**: Multiple peer organizations and orderer cluster
- **Production Ready**: Security groups, encrypted volumes, and proper networking

## Prerequisites

1. **AWS Account** with credentials configured:
   ```bash
   export AWS_ACCESS_KEY_ID="your-access-key"
   export AWS_SECRET_ACCESS_KEY="your-secret-key"
   ```

2. **AWS EC2 Key Pair** created in your target region:
   ```bash
   aws ec2 create-key-pair --key-name fabric-nodes --query 'KeyMaterial' --output text > ~/.ssh/fabric-nodes.pem
   chmod 400 ~/.ssh/fabric-nodes.pem
   ```

3. **Chainlaunch Instance** - Either:
   - **Option A**: Use an existing Chainlaunch instance (set `install_chainlaunch_server = false`)
   - **Option B**: Let Terraform install Chainlaunch on AWS (set `install_chainlaunch_server = true`)

## Configuration

### 1. Create `terraform.tfvars`

**Option A: Using Existing Chainlaunch Instance**

```hcl
# AWS Configuration
aws_region    = "us-east-1"
key_pair_name = "fabric-nodes"  # Your EC2 key pair name
instance_type = "t3.medium"     # Adjust based on workload

# Chainlaunch Configuration - Use existing instance
install_chainlaunch_server = false
chainlaunch_url            = "https://your-chainlaunch-instance.com"
chainlaunch_username       = "admin"
chainlaunch_password       = "your-password"

# Network Configuration
project_name   = "fabric-network"
environment    = "dev"
fabric_version = "2.5.9"
```

**Option B: Auto-Install Chainlaunch on AWS**

```hcl
# AWS Configuration
aws_region    = "us-east-1"
key_pair_name = "fabric-nodes"  # Your EC2 key pair name
instance_type = "t3.medium"     # Adjust based on workload

# Chainlaunch Auto-Installation
install_chainlaunch_server = true
management_instance_type   = "t3.medium"
chainlaunch_version        = "latest"
ssh_private_key_path       = "~/.ssh/fabric-nodes.pem"
chainlaunch_username       = "admin"  # Default credentials
chainlaunch_password       = "admin123"

# Network Configuration
project_name   = "fabric-network"
environment    = "dev"
fabric_version = "2.5.9"

# Nodes Configuration (optional - defaults provided)
peer_nodes = {
  "peer0-org1" = { organization = "Org1" }
  "peer1-org1" = { organization = "Org1" }
  "peer0-org2" = { organization = "Org2" }
}

orderer_nodes = {
  "orderer0" = { organization = "OrdererOrg" }
  "orderer1" = { organization = "OrdererOrg" }
  "orderer2" = { organization = "OrdererOrg" }
}
```

### 2. Initialize and Deploy

```bash
# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Deploy the infrastructure and nodes
terraform apply

# View outputs
terraform output
```

## How It Works

### 1. AWS Infrastructure Provisioning

First, Terraform creates AWS resources:
- VPC and networking components
- EC2 instances for each peer and orderer
- Security groups with appropriate ports open

Each EC2 instance:
- Runs Ubuntu 22.04 LTS
- Has Docker pre-installed via user_data
- Gets a public IP address
- Has encrypted root volume

### 2. Chainlaunch Node Provisioning

After AWS instances are ready, Terraform provisions Chainlaunch resources:

```hcl
# Example: Peer node using AWS instance IP
resource "chainlaunch_fabric_peer" "peer0_org1" {
  name = "peer0-org1"

  # Uses AWS instance public IP
  external_endpoint = "${aws_instance.fabric_peers["peer0-org1"].public_ip}:7051"

  # Maps external IP to internal IP for Docker networking
  address_overrides = [{
    from = "${aws_instance.fabric_peers["peer0-org1"].public_ip}:7051"
    to   = "${aws_instance.fabric_peers["peer0-org1"].private_ip}:7051"
  }]

  # ... other configuration
}
```

### 3. IP Address Flow

```
Client → Public IP (54.x.x.x:7051) → AWS Security Group →
EC2 Instance → Private IP (10.0.1.x:7051) → Docker Container
```

The `address_overrides` ensures that:
- External clients use the public IP
- Internal Docker networking uses the private IP
- Traffic flows correctly through AWS networking

## Network Ports

### Peer Nodes
- **7051**: Peer gRPC endpoint (external communication)
- **7052**: Chaincode endpoint
- **7053**: Events endpoint
- **9443**: Operations/Metrics (Prometheus)
- **22**: SSH access

### Orderer Nodes
- **7050**: Orderer gRPC endpoint
- **9443**: Operations/Metrics (Prometheus)
- **22**: SSH access

## Outputs

After successful deployment, you'll see:

```bash
# Connection strings for applications
peer_connection_strings = {
  "peer0-org1" = "grpcs://54.123.45.67:7051"
  "peer1-org1" = "grpcs://54.123.45.68:7051"
  # ...
}

# SSH commands for accessing instances
ssh_commands = {
  "ssh_peer0-org1" = "ssh -i ~/.ssh/fabric-nodes.pem ubuntu@54.123.45.67"
  # ...
}

# Network summary
network_summary = {
  peer_count     = 4
  orderer_count  = 3
  peer_orgs      = 2
  orderer_orgs   = 1
  fabric_version = "2.5.9"
  aws_region     = "us-east-1"
}
```

## Connecting to Nodes

### Via SSH

```bash
# Get SSH command from outputs
terraform output ssh_commands

# Connect to a peer
ssh -i ~/.ssh/fabric-nodes.pem ubuntu@<peer-public-ip>

# Check Docker containers
sudo docker ps

# View Fabric logs
sudo docker logs <container-id>
```

### Via Fabric Client

Use the connection strings from outputs:

```bash
# Example: Query peer
peer channel list \
  --connTimeout 10s \
  --peerAddresses grpcs://54.123.45.67:7051 \
  --tlsRootCertPaths /path/to/peer-tls-cert.pem
```

## Scaling the Network

### Add More Peers

Edit `terraform.tfvars`:

```hcl
peer_nodes = {
  "peer0-org1" = { organization = "Org1" }
  "peer1-org1" = { organization = "Org1" }
  "peer2-org1" = { organization = "Org1" }  # New peer
  "peer0-org2" = { organization = "Org2" }
  "peer0-org3" = { organization = "Org3" }  # New organization
}
```

Then apply:
```bash
terraform apply
```

### Add More Orderers

```hcl
orderer_nodes = {
  "orderer0" = { organization = "OrdererOrg" }
  "orderer1" = { organization = "OrdererOrg" }
  "orderer2" = { organization = "OrdererOrg" }
  "orderer3" = { organization = "OrdererOrg" }  # New orderer
}
```

## Cost Optimization

### Development/Testing
```hcl
instance_type = "t3.small"   # ~$15/month per instance
# Use 1-2 peers and 1 orderer
```

### Production
```hcl
instance_type = "t3.large"   # ~$60/month per instance
# Use redundant peers and 3+ orderers
```

### Estimated Costs (us-east-1)

| Configuration | Instances | Type | Monthly Cost |
|---------------|-----------|------|--------------|
| Minimal | 2 peers + 1 orderer | t3.small | ~$45 |
| Development | 4 peers + 3 orderers | t3.medium | ~$245 |
| Production | 6 peers + 5 orderers | t3.large | ~$660 |

*Does not include data transfer, storage, or other AWS costs*

## Troubleshooting

### EC2 Instances Not Accessible

1. Check security group rules:
   ```bash
   aws ec2 describe-security-groups --group-ids <sg-id>
   ```

2. Verify instance is running:
   ```bash
   aws ec2 describe-instances --instance-ids <instance-id>
   ```

3. Test connectivity:
   ```bash
   nc -zv <public-ip> 7051
   ```

### Chainlaunch Nodes Not Connecting

1. Check if Docker is running on EC2:
   ```bash
   ssh ubuntu@<public-ip> 'sudo docker ps'
   ```

2. Verify address overrides are correct:
   ```bash
   terraform state show chainlaunch_fabric_peer.peers["peer0-org1"]
   ```

3. Check Chainlaunch logs in the web UI

### AWS Rate Limits

If you hit AWS API rate limits:
```bash
# Add delay between resource creation
terraform apply -parallelism=5
```

## Next Steps

After deploying the network:

1. **Create a Channel**: Use `chainlaunch_fabric_network` resource
2. **Join Nodes**: Use `chainlaunch_fabric_join_node` resource
3. **Deploy Chaincode**: Use the chaincode lifecycle resources
4. **Setup Monitoring**: Add Prometheus/Grafana for metrics
5. **Configure Backups**: Use `chainlaunch_backup_target` and schedules

## Clean Up

To destroy all resources:

```bash
# Destroy Chainlaunch resources first (recommended)
terraform destroy -target=chainlaunch_fabric_peer.peers
terraform destroy -target=chainlaunch_fabric_orderer.orderers
terraform destroy -target=chainlaunch_fabric_organization.peer_orgs
terraform destroy -target=chainlaunch_fabric_organization.orderer_orgs

# Then destroy AWS infrastructure
terraform destroy
```

Or destroy everything at once:
```bash
terraform destroy
```

**Warning**: This will permanently delete:
- All EC2 instances and their data
- All VPC resources
- All Chainlaunch organizations and nodes

## Security Considerations

### Production Checklist

- [ ] Restrict SSH access to specific IP ranges in security groups
- [ ] Use AWS Secrets Manager for Chainlaunch credentials
- [ ] Enable VPC Flow Logs for network monitoring
- [ ] Configure CloudWatch alarms for instance health
- [ ] Use AWS KMS for volume encryption keys
- [ ] Implement backup strategy for node data
- [ ] Enable MFA for AWS account access
- [ ] Review and minimize IAM permissions
- [ ] Use private subnets with NAT gateway for production
- [ ] Configure AWS WAF if exposing endpoints publicly

## Related Examples

- [fabric-network-complete](../fabric-network-complete/) - Complete Fabric network with channels
- [besu-network-complete](../besu-network-complete/) - Besu network deployment
- [backup-with-minio](../backup-with-minio/) - Automated backups configuration
- [metrics-monitoring](../metrics-monitoring/) - Prometheus/Grafana monitoring

## Support

For issues specific to:
- **AWS Resources**: Check [AWS Documentation](https://docs.aws.amazon.com/)
- **Chainlaunch Provider**: Open an issue on [GitHub](https://github.com/kfsoftware/terraform-provider-chainlaunch/issues)
- **Fabric Concepts**: See [Hyperledger Fabric Docs](https://hyperledger-fabric.readthedocs.io/)
