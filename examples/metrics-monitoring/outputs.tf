output "prometheus_url" {
  description = "Prometheus web UI URL"
  value       = "http://localhost:${chainlaunch_metrics_prometheus.monitoring.port}"
}

output "prometheus_status" {
  description = "Prometheus instance status"
  value       = chainlaunch_metrics_prometheus.monitoring.status
}

output "prometheus_version" {
  description = "Deployed Prometheus version"
  value       = chainlaunch_metrics_prometheus.monitoring.version
}
