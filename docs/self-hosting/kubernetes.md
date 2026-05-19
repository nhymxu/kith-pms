# Kubernetes

## Prerequisites

- Kubernetes 1.28+
- kubectl + kustomize
- A container registry with the kith-pms image pushed

## Apply base manifests

```bash
# Edit deploy/k8s/base/secret.example.yaml with real values, then:
kubectl apply -f deploy/k8s/base/secret.example.yaml

# Apply the rest
kubectl apply -k deploy/k8s/base
```

## Apply with optional components

```bash
# Edit deploy/k8s/overlays/example/kustomization.yaml — set your image and domain
kubectl apply -k deploy/k8s/overlays/example
```

## Critical: single replica invariant

**Never scale beyond 1 replica.** SQLite does not support concurrent writers. Scaling to 2+ pods will corrupt the database. The deployment manifest enforces `replicas: 1` and `strategy: Recreate`.

## PVC data retention

Set `reclaimPolicy: Retain` on your StorageClass to prevent data loss when the PVC is deleted:

```bash
kubectl patch storageclass <your-class> -p '{"reclaimPolicy":"Retain"}'
```

## Upgrade

```bash
kubectl set image deployment/kith kith=ghcr.io/nhymxu/kith-pms:<new-tag> -n kith-pms
```

The `Recreate` strategy ensures the old pod stops before the new one starts.

## Optional: Prometheus scraping

Apply the service-monitor component to enable scraping via Prometheus Operator:

```bash
kubectl apply -k deploy/k8s/components/service-monitor
```

See [Metrics](metrics.md) for available metric names.
