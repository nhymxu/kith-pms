# Metrics

kith-pms exposes a Prometheus-compatible metrics endpoint at `/metrics`.

## Available metrics

| Metric | Type | Description |
|---|---|---|
| `http_requests_total` | Counter | Total HTTP requests by method, route, and status code |
| `http_request_duration_seconds` | Histogram | Request latency by method and route |
| `kith_db_size_bytes` | Gauge | SQLite database size in bytes |
| `kith_sessions_active` | Gauge | Number of non-expired sessions |
| `kith_build_info` | Gauge | Build metadata (version, commit); value always 1 |

Standard Go runtime metrics (`go_*`, `process_*`) are also included.

## Scrape with Prometheus

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: kith-pms
    static_configs:
      - targets: ['localhost:8000']
```

## Security note

`/metrics` is **unauthenticated**. Bind kith-pms to a private interface or use a reverse proxy with IP allowlisting to restrict access in production.

## Kubernetes (Prometheus Operator)

Apply the ServiceMonitor component:

```bash
kubectl apply -k deploy/k8s/components/service-monitor
```

This creates a `ServiceMonitor` that scrapes `/metrics` every 30 seconds.
