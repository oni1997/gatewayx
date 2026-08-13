# Installation

## Binary releases

Pre-built binaries are available on the [GitHub Releases](https://github.com/oni1997/gatewayx/releases) page for Linux, macOS, and Windows (amd64/arm64).

## Docker / Podman

```bash
podman pull ghcr.io/oni1997/gatewayx:latest
```

## Homebrew

```bash
brew install oni1997/tap/gatewayx
```

## Debian / RPM

```bash
# .deb
dpkg -i gatewayx_0.4.0_linux_amd64.deb

# .rpm
rpm -i gatewayx-0.4.0.x86_64.rpm
```

## From source

```bash
git clone https://github.com/oni1997/gatewayx.git
cd gatewayx
make build
```
