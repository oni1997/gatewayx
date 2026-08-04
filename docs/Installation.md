# Installation

## Binary Releases

Download pre-built binaries from the [GitHub Releases](https://github.com/oni1997/gatewayx/releases) page.

```bash
# Linux (amd64)
curl -L https://github.com/oni1997/gatewayx/releases/latest/download/gatewayx_linux_amd64.tar.gz | tar xz
sudo mv gatewayx /usr/local/bin/

# macOS (arm64)
curl -L https://github.com/oni1997/gatewayx/releases/latest/download/gatewayx_darwin_arm64.tar.gz | tar xz
sudo mv gatewayx /usr/local/bin/
```

## Docker

```bash
docker pull ghcr.io/oni1997/gatewayx:latest
```

## Docker Compose

```yaml
version: "3.8"
services:
  gatewayx:
    image: ghcr.io/oni1997/gatewayx:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    volumes:
      - ./gatewayx.yaml:/etc/gatewayx/gatewayx.yaml
    command: ["-c", "/etc/gatewayx/gatewayx.yaml"]
```

## Kubernetes

```bash
helm repo add gatewayx https://charts.gatewayx.dev
helm install gatewayx gatewayx/gatewayx
```

## From Source

```bash
git clone https://github.com/oni1997/gatewayx.git
cd gatewayx
make build
sudo make install
```
