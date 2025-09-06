# Prometheus & Grafana Monitoring Stack

This Docker Compose setup provides a complete monitoring stack with Prometheus, Grafana, and Node Exporter.

## Services

- **Prometheus**: Metrics collection and storage (Port: 9090)
- **Grafana**: Metrics visualization and dashboards (Port: 3000)
- **Node Exporter**: System metrics collection (Port: 9100)

## Quick Start

1. Start the monitoring stack:
   ```bash
   docker-compose up -d
   ```

2. Access the services:
   - Grafana: http://localhost:3000 (admin/admin)
   - Prometheus: http://localhost:9090
   - Node Exporter: http://localhost:9100

## Configuration

### Prometheus
- Configuration: `prometheus/prometheus.yml`
- Data retention: 200 hours
- Web UI enabled with lifecycle management

### Grafana
- Default credentials: admin/admin
- Prometheus datasource pre-configured
- Persistent data storage

### Adding Application Metrics

To monitor your DLMS applications, uncomment and modify the scrape configs in `prometheus/prometheus.yml`:

```yaml
- job_name: 'dlms-processor'
  static_configs:
    - targets: ['host.docker.internal:8080']

- job_name: 'dlms-consumer'
  static_configs:
    - targets: ['host.docker.internal:8081']
```

## Volumes

- `prometheus_data`: Persistent Prometheus data
- `grafana_data`: Persistent Grafana configuration and dashboards

## Network

All services run on the `monitoring` bridge network for internal communication.

## Stopping the Stack

```bash
docker-compose down
```

To remove all data:
```bash
docker-compose down -v
```
