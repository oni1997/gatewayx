# Kubernetes Deployment

This directory contains Kubernetes manifests and Helm charts for deploying GatewayX.

## Quick Deploy

```bash
kubectl apply -f deploy/kubernetes/
```

## Helm

```bash
helm install gatewayx ./deploy/helm/gatewayx
```

Kubernetes support is scheduled for Phase 7 of development.
