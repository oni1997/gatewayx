# Deployment

## Docker / Podman

```bash
podman run -d -p 8080:8080 -p 9090:9090 \
  -v $(pwd)/gatewayx.yaml:/etc/gatewayx/gatewayx.yaml \
  ghcr.io/oni1997/gatewayx:latest
```

## Kubernetes

```bash
helm install gatewayx ./deploy/helm/gatewayx
```

Or apply raw manifests:

```bash
kubectl apply -f deploy/kubernetes/
```

## Raspberry Pi

```bash
GOOS=linux GOARCH=arm64 go build -o bin/gatewayx ./apps/gateway
./bin/gatewayx
```

Runs comfortably on a Pi 4 with 50-100MB RAM.
