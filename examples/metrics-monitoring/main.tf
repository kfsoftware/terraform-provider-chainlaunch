# ==============================================================================
# Prometheus Metrics Monitoring Example
# ==============================================================================
# This example demonstrates how to:
# 1. Deploy Prometheus for monitoring
# 2. Create metrics jobs to monitor existing nodes
# 3. Automatically synchronize node endpoints with Prometheus
#
# Prerequisites:
# - Existing Fabric peers/orderers with metrics enabled
# - Or manually configure target endpoints
# ==============================================================================

terraform {
  required_providers {
    chainlaunch = {
      source  = "registry.terraform.io/chainlaunch/chainlaunch"
      version = "0.1.0"
    }
  }
}

provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

# ==============================================================================
# STEP 1: Deploy Prometheus
# ==============================================================================

resource "chainlaunch_metrics_prometheus" "monitoring" {
  version         = var.prometheus_version
  port            = var.prometheus_port
  scrape_interval = var.scrape_interval
  deployment_mode = var.deployment_mode
  network_mode    = var.network_mode
}

# ==============================================================================
# STEP 2: Configure metrics scrape jobs for existing nodes
# ==============================================================================

# Monitor Fabric peers
resource "chainlaunch_metrics_job" "fabric_peers" {
  count = length(var.peer_metrics_targets) > 0 ? 1 : 0

  job_name = "fabric-peers"
  targets  = var.peer_metrics_targets

  metrics_path    = "/metrics"
  scrape_interval = "15s"

  depends_on = [chainlaunch_metrics_prometheus.monitoring]
}

# Monitor Fabric orderers
resource "chainlaunch_metrics_job" "fabric_orderers" {
  count = length(var.orderer_metrics_targets) > 0 ? 1 : 0

  job_name = "fabric-orderers"
  targets  = var.orderer_metrics_targets

  metrics_path    = "/metrics"
  scrape_interval = "15s"

  depends_on = [chainlaunch_metrics_prometheus.monitoring]
}

# Monitor Besu nodes
resource "chainlaunch_metrics_job" "besu_nodes" {
  count = length(var.besu_metrics_targets) > 0 ? 1 : 0

  job_name = "besu-nodes"
  targets  = var.besu_metrics_targets

  metrics_path    = "/metrics"
  scrape_interval = "15s"

  depends_on = [chainlaunch_metrics_prometheus.monitoring]
}

# Monitor custom services
resource "chainlaunch_metrics_job" "custom_services" {
  count = length(var.custom_monitoring_targets) > 0 ? 1 : 0

  job_name = "custom-services"
  targets  = var.custom_monitoring_targets

  metrics_path    = var.custom_metrics_path
  scrape_interval = var.custom_scrape_interval

  depends_on = [chainlaunch_metrics_prometheus.monitoring]
}
