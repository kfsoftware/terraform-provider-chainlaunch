# Prometheus Metrics Monitoring Example

This example demonstrates how to deploy Prometheus and automatically synchronize Chainlaunch node metrics for monitoring.

## Overview

This example shows:
- **Prometheus Deployment**: Deploy Prometheus in Docker or binary mode
- **Automatic Synchronization**: Metrics jobs that automatically collect all node endpoints
- **Multi-Node Monitoring**: Monitor multiple peers and orderers simultaneously
- **Dynamic Configuration**: Add/remove nodes and metrics automatically update

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Terraform Configuration                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┐                                          │
│  │   Prometheus     │                                          │
│  │   v2.45.0:9090   │                                          │
│  └────────┬─────────┘                                          │
│           │                                                     │
│           │ scrapes                                             │
│           │                                                     │
│    ┌──────┴──────────────────────────────┐                    │
│    │                                      │                    │
│    ▼                                      ▼                    │
│  ┌────────────────┐                  ┌────────────────┐       │
│  │ Metrics Job:   │                  │ Metrics Job:   │       │
│  │ fabric-peers   │                  │ fabric-orderers│       │
│  │                │                  │                │       │
│  │ • peer0:9443   │                  │ • orderer0:9440│       │
│  │ • peer1:9444   │                  │ • orderer1:9441│       │
│  └────────────────┘                  └────────────────┘       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Chainlaunch instance running on http://localhost:8100
- Docker installed (for docker deployment mode)
- Admin credentials (default: admin/admin123)

## Usage

### Quick Start

```bash
cd examples/metrics-monitoring

# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Deploy Prometheus and create monitored nodes
terraform apply -auto-approve

# View monitoring summary
terraform output monitoring_summary
```

### Expected Output

```
Apply complete! Resources: 7 added, 0 changed, 0 destroyed.

Outputs:

monitoring_summary = <<EOT

╔══════════════════════════════════════════════════════════════╗
║           Prometheus Monitoring Setup                       ║
╚══════════════════════════════════════════════════════════════╝

Prometheus:
  URL:     http://localhost:9090
  Version: v2.45.0
  Status:  running

Monitored Nodes:
  Peers:    2 nodes
  Orderers: 1 nodes

Scrape Jobs:
  - fabric-peers (2 targets)
  - fabric-orderers (1 targets)

Next Steps:
  1. Open Prometheus UI: http://localhost:9090
  2. Check targets: http://localhost:9090/targets
  3. Query metrics: http://localhost:9090/graph

EOT

prometheus_url = "http://localhost:9090"
```

## Configuration

### Prometheus Settings

```hcl
# Use latest Prometheus version
prometheus_version = "v2.47.0"

# Custom port
prometheus_port = 9091

# Faster scraping
scrape_interval = 10
```

### Node Count

```hcl
# Create 5 peers
peer_count = 5

# Create 3 orderers
orderer_count = 3
```

### Deployment Mode

```hcl
# Use binary mode instead of Docker
deployment_mode = "binary"

# Use host networking (recommended for production)
network_mode = "host"
```

### Custom Monitoring

```hcl
# Enable monitoring of external services
enable_custom_monitoring = true

custom_monitoring_targets = [
  "my-app.example.com:8080",
  "my-database.example.com:9187"
]
```

## Resources Created

### 1. `chainlaunch_metrics_prometheus`

Deploys Prometheus instance:
- **Singleton**: Only one Prometheus instance per Chainlaunch
- **Port**: Default 9090
- **Version**: Configurable (default: v2.45.0)
- **Mode**: Docker or binary deployment

### 2. `chainlaunch_metrics_job`

Creates Prometheus scrape jobs:
- **Dynamic targets**: Automatically collects node endpoints
- **Auto-sync**: Updates when nodes are added/removed
- **Configurable**: Custom scrape intervals and metrics paths

### 3. Fabric Nodes

Creates peers and orderers with metrics enabled:
- **Metrics enabled**: All nodes expose Prometheus metrics
- **Unique ports**: Each node gets its own metrics port
- **Standard format**: `/metrics` endpoint

## Accessing Prometheus

### Web UI

Open http://localhost:9090 in your browser

### Check Targets

View all monitored targets and their health:
http://localhost:9090/targets

### Query Metrics

Examples of useful queries:

**Peer Block Height**:
```promql
fabric_ledger_blockchain_height{job="fabric-peers"}
```

**Orderer Consensus Leader**:
```promql
fabric_consensus_etcdraft_is_leader{job="fabric-orderers"}
```

**Transaction Rate**:
```promql
rate(fabric_txvalidation_total[5m])
```

## Automatic Synchronization

The metrics jobs automatically synchronize with your infrastructure:

### Adding Nodes

```bash
# Add 2 more peers
terraform apply -var="peer_count=4"
```

Terraform will:
1. Create the new peers
2. Automatically update the `fabric-peers` job with new endpoints
3. Prometheus immediately starts scraping the new peers

### Removing Nodes

```bash
# Reduce to 1 peer
terraform apply -var="peer_count=1"
```

Terraform will:
1. Remove the extra peers
2. Automatically update the `fabric-peers` job
3. Prometheus stops scraping removed peers

## Advanced Usage

### Multiple Organizations

```hcl
# Create separate jobs per organization
resource "chainlaunch_metrics_job" "org1_peers" {
  job_name = "org1-peers"
  targets  = [for peer in chainlaunch_fabric_peer.org1_peers : "host.docker.internal:${peer.metrics_port}"]
}

resource "chainlaunch_metrics_job" "org2_peers" {
  job_name = "org2-peers"
  targets  = [for peer in chainlaunch_fabric_peer.org2_peers : "host.docker.internal:${peer.metrics_port}"]
}
```

### Custom Scrape Intervals

```hcl
# Fast scraping for critical services
resource "chainlaunch_metrics_job" "critical" {
  job_name        = "critical-services"
  targets         = var.critical_endpoints
  scrape_interval = "5s"  # Every 5 seconds
}

# Slow scraping for non-critical
resource "chainlaunch_metrics_job" "non_critical" {
  job_name        = "non-critical"
  targets         = var.other_endpoints
  scrape_interval = "60s"  # Every minute
}
```

### External Services

```hcl
# Monitor PostgreSQL
resource "chainlaunch_metrics_job" "postgres" {
  job_name = "postgresql"
  targets  = ["postgres-exporter.example.com:9187"]
}

# Monitor NGINX
resource "chainlaunch_metrics_job" "nginx" {
  job_name = "nginx"
  targets  = ["nginx-exporter.example.com:9113"]
}
```

## Troubleshooting

### Prometheus not starting

**Check status**:
```bash
terraform output prometheus_status
```

**Check logs** (if using Docker):
```bash
docker logs chainlaunch-prometheus
```

### Targets not showing up

**Verify metrics job**:
```bash
terraform show | grep chainlaunch_metrics_job
```

**Check Prometheus config**:
```bash
# If using Docker
docker exec chainlaunch-prometheus cat /etc/prometheus/prometheus.yml
```

### Metrics not accessible

**Test endpoint directly**:
```bash
curl http://localhost:9443/metrics  # Peer 0
curl http://localhost:9440/metrics  # Orderer 0
```

**Check node metrics enabled**:
```hcl
# Ensure this is set
metrics_enabled = true
```

## Integration with Grafana

After deploying Prometheus, you can add Grafana for visualization:

```bash
# Run Grafana with Docker
docker run -d \
  --name=grafana \
  -p 3000:3000 \
  grafana/grafana

# Open http://localhost:3000
# Add Prometheus datasource: http://host.docker.internal:9090
```

Grafana dashboards for Fabric:
- Dashboard ID 10918 - Hyperledger Fabric
- Dashboard ID 12734 - Fabric Peer Metrics

## Cleanup

```bash
# Remove all resources
terraform destroy -auto-approve
```

This will:
1. Remove metrics jobs from Prometheus
2. Delete all monitored nodes
3. Stop and remove Prometheus

## Related Examples

- [Fabric Network Complete](../fabric-network-complete/) - Full Fabric network
- [Besu Network](../besu-network/) - Besu blockchain monitoring
- [Backup with MinIO](../backup-with-minio/) - Backup configuration

## API Endpoints Used

- `POST /metrics/deploy` - Deploy Prometheus
- `GET /metrics/status` - Get Prometheus status
- `POST /metrics/job/add` - Add scrape job
- `DELETE /metrics/job/{jobName}` - Remove scrape job
- `GET /metrics/jobs` - List all jobs
