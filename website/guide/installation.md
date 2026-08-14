# Installation

## Binary releases

Pre-built binaries are available on the [GitHub Releases](https://github.com/oni1997/gatewayx/releases) page for Linux, macOS, and Windows (amd64/arm64).

Replace `v0.4.3` with the latest version:

```bash
# Linux (amd64)
curl -L https://github.com/oni1997/gatewayx/releases/download/v0.4.3/gatewayx_0.4.3_linux_amd64.tar.gz | tar xz
sudo mv gatewayx gatewayx-cli /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/oni1997/gatewayx/releases/download/v0.4.3/gatewayx_0.4.3_darwin_arm64.tar.gz | tar xz
sudo mv gatewayx gatewayx-cli /usr/local/bin/
```

Or always download the latest:

```bash
VERSION=$(curl -s https://api.github.com/repos/oni1997/gatewayx/releases/latest | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4)
curl -L "https://github.com/oni1997/gatewayx/releases/download/${VERSION}/gatewayx_${VERSION#v}_linux_amd64.tar.gz" | tar xz
sudo mv gatewayx gatewayx-cli /usr/local/bin/
```

Verify:

```bash
gatewayx-cli version
```

## Docker / Podman

```bash
podman pull ghcr.io/oni1997/gatewayx:latest
podman run -d -p 8080:8080 -p 9090:9090 \
  -v $(pwd)/gatewayx.yaml:/etc/gatewayx/gatewayx.yaml \
  ghcr.io/oni1997/gatewayx:latest
```

## Debian / RPM

```bash
# .deb (Debian/Ubuntu)
sudo dpkg -i gatewayx_0.4.3_linux_amd64.deb

# .rpm (Fedora/RHEL)
sudo rpm -i gatewayx_0.4.3_linux_amd64.rpm
```

## From source

```bash
git clone https://github.com/oni1997/gatewayx.git
cd gatewayx
make build        # builds bin/gatewayx and bin/gatewayx-cli
sudo make install
```
