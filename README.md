# vmlg — Modern Single-Binary Multi-Node Network & BGP Looking Glass

[![CI](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml)
[![Docker](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cloudapp3/vmlg.svg)](https://pkg.go.dev/github.com/cloudapp3/vmlg)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**English** · [简体中文](docs/README.zh-CN.md) · [繁體中文](docs/README.zh-TW.md) · [日本語](docs/README.ja.md) · [Deutsch](docs/README.de.md) · [Русский](docs/README.ru.md)

**vmlg** is a self-hosted, pure-Go (`CGO_ENABLED=0`), zero-Node-build Network and BGP Looking Glass portal designed for VPS providers, IXPs, AS operators, DN42 networks, and multi-region homelabs. One static ~8.6 MB binary embeds a modern multi-language web UI, a single-node server (`standalone`), a central portal (`hub`), a lightweight regional daemon (`agent`), and an AI-ready `stdio` MCP server—with zero external database or Redis required.

---

## Key Features

- **Single Binary, Dual-Mode Architecture (`standalone` + `hub`/`agent`)**
  - **Standalone**: Run on a single host with zero configuration (`vmlg serve -listen :8080`).
  - **Hub + Agent**: Regional agents connect outbound over a multiplexed WebSocket tunnel (`ws://` / `wss://`) to a central Hub, working seamlessly behind NAT or firewalls without requiring public management ports on every node.
- **One-Click Multi-Node Latency & Return-Route Comparison**
  - Automatically detects the visitor's public IP and ASN. With a single click on **Probe My IP**, all online regional nodes concurrently measure RTT, jitter, packet loss, and return-route hops to the visitor, sorting datacenters by live latency in one unified view.
- **Streaming MTR & Ping with Carrier Backbone & IXP Enrichment**
  - Real-time Server-Sent Events (SSE) streaming with automatic `X-Accel-Buffering: no` headers to bypass reverse-proxy buffering.
  - Native Go `ICMPv4/v6` and `TCP` Ping plus multi-cycle live MTR; automatically falls back to unprivileged Linux kernel `IP_RECVERR` (`MSG_ERRQUEUE`) hop tracing inside rootless containers lacking `CAP_NET_RAW`.
  - Built-in offline prefix, ASN, and rDNS classification for **`CN2 GIA` / `CN2 GT` / `CU 9929` / `CU 10099` / `CMIN2` / `CT 163` / `CU 4837` / `CMI` / `IIJ` / `SoftBank` / `NTT`** and major Internet Exchanges (**`HKIX` / `DE-CIX` / `AMS-IX` / `LINX` / `Equinix` / `JPIX` / `Any2IX`**).
- **Zero-Disk Speedtest Streams, HTML5 Benchmark & Ephemeral iPerf3 Leases**
  - `/speedtest/files/{10MB,100MB,1GB}.bin` streams incompressible `ChaCha8` pseudo-random blocks directly from memory with **0 bytes of disk footprint** and optional per-stream bandwidth throttling.
  - Built-in multi-stream HTML5 Download, Upload (zero-allocation discard sink), Ping, and Jitter benchmark with live throughput charts.
  - **Controlled iPerf3 Port Lease Pool**: Dynamically allocates an isolated port (e.g., `5201-5210`) running a one-shot `iperf3 -s -1` subprocess with strict TTL auto-reclamation, preventing public `iperf3` servers from being permanently saturated.
- **Pluggable Read-Only BGP RIB & Pure SVG AS-Path Topology**
  - Optional read-only drivers for **BIRD 2/3** (native `/run/bird/bird.ctl` Unix socket speaking `restrict` mode with zero process fork), **FRRouting** (`vtysh` JSON), **GoBGP**, and demo `mock` mode.
  - Client-side **pure SVG AS-Path directed topology graph** (zero Graphviz dependency on the server), Standard/Large BGP Community semantic decoding, and BGP neighbor session summaries.
- **Hardened Anti-Abuse Security Guard**
  - **Zero Shell Execution**: Every input is strictly parsed via `net/netip` and RFC 1123 hostname rules.
  - **SSRF & DNS-Rebinding Protection**: Blocks loopback, RFC1918 private, CGNAT, link-local, and cloud metadata (`169.254.169.254`) targets across all resolved DNS records (customizable via `allowed_reserved_cidrs` for DN42).
  - **Dual-Dimension Token Buckets (`Per-Client` + `Per-Target`)**: Rate-limits both individual client IPs (`/64` for IPv6) and **destination `/24` (IPv4) / `/48` (IPv6) subnets** globally—preventing distributed DDoS reflection/amplification against a single victim.
- **Observability, MCP Server & Redacted Markdown Sharing**
  - Built-in Prometheus `/metrics` endpoint (with optional Bearer token auth).
  - Built-in `vmlg mcp` stdio JSON-RPC server (`vmlg_list_nodes`, `vmlg_ping`, `vmlg_mtr`, `vmlg_compare`, `vmlg_bgp_route`) for both local execution and remote `-hub-http` multi-node portals.
  - Multi-language dark/light web UI (`English` default, with `简体中文`, `繁體中文`, `日本語`) and one-click **Redacted Forum Markdown** export (`203.0.113.x`).

---

## Quick Start

### 1. Install or Build Binary

```bash
# Install via one-line release script (verifies SHA-256 checksum)
curl -fsSL https://raw.githubusercontent.com/cloudapp3/vmlg/main/install.sh | bash

# Or install with Go toolchain (zero Node.js or CGO dependencies)
go install github.com/cloudapp3/vmlg/cmd/vmlg@latest

# Or build static binary from source
make build
```

### 2. Standalone Mode (Single Node, Zero Config)

```bash
vmlg serve \
  -listen 0.0.0.0:8080 \
  -node-id tyo-01 \
  -node-name "Tokyo · BGP Premium" \
  -location "Tokyo, JP" \
  -region "APAC"
```

Open `http://127.0.0.1:8080` in your browser. See [examples/standalone.toml](examples/standalone.toml) for full TOML configuration options.

### 3. Multi-Node Cluster Mode (`hub` + `agent`)

**Start the central Hub portal:**

```bash
export VMLG_AGENT_TOKEN="replace-with-strong-random-token"
vmlg hub \
  -listen 0.0.0.0:8080 \
  -title "Global Cloud Looking Glass" \
  -enable-local-node \
  -node-id hkg-01 \
  -node-name "Hong Kong 01" \
  -region "APAC"
```

**Connect regional Agent nodes:**

```bash
export VMLG_AGENT_TOKEN="replace-with-strong-random-token"
vmlg agent \
  -hub-url ws://hub.example.com:8080/api/v1/cluster/ws \
  -node-id lax-01 \
  -node-name "Los Angeles · CN2 GIA" \
  -location "Los Angeles, US" \
  -region "NA" \
  -tags "CN2 GIA,10Gbps" \
  -speedtest-listen 0.0.0.0:8081 \
  -public-url https://lax-speed.example.com
```

Validate any configuration file before deployment:

```bash
vmlg check -config ./examples/standalone.toml
```

### 4. Docker & Docker Compose Deployment

Official multi-arch images (`linux/amd64`, `linux/arm64`) are published to `ghcr.io/cloudapp3/vmlg:latest` with `iperf3` pre-installed and running as non-root UID `10001`.

**Option A: Docker Compose (Standalone)**

```bash
docker compose up -d
```

**Option B: Single `docker run` Command (Environment Variables)**

```bash
docker run -d \
  --name vmlg \
  --restart unless-stopped \
  --cap-drop ALL \
  --cap-add NET_RAW \
  --sysctl net.ipv4.ping_group_range="0 2147483647" \
  -p 8080:8080 \
  -p 5201-5210:5201-5210/tcp \
  -p 5201-5210:5201-5210/udp \
  -e VMLG_NODE_ID=tyo-01 \
  -e VMLG_NODE_NAME="Tokyo · BGP Premium" \
  -e VMLG_NODE_LOCATION="Tokyo, JP" \
  -e VMLG_NODE_REGION=APAC \
  ghcr.io/cloudapp3/vmlg:latest
```

> **Tip**: To probe native host IPv6 routes or mount the host's BIRD Unix socket, use `--network host` and add `-v /run/bird/bird.ctl:/run/bird/bird.ctl:ro -e VMLG_BGP_DRIVER=bird`.

**Option C: Multi-Node Hub + Agent Compose Stack**

See [docker-compose.cluster.yml](docker-compose.cluster.yml) (supports environment variables as well as Docker Secrets via `*_FILE`, such as `VMLG_AGENT_TOKEN_FILE`):

```bash
VMLG_AGENT_TOKEN="replace-with-strong-token" \
  docker compose -f docker-compose.cluster.yml --profile agent up -d
```

### 5. MCP Server (AI-Assisted Network Diagnostics)

```bash
# Local standalone diagnostics over stdio (no listening port required)
vmlg mcp

# Connect to a deployed multi-region Hub portal
vmlg mcp -hub-http https://lg.example.com
```

---

## Linux Unprivileged Socket Notes

`vmlg` automatically falls back to Linux `IP_RECVERR` (`MSG_ERRQUEUE`) hop tracing when `CAP_NET_RAW` is unavailable. On bare-metal or VM systemd deployments running under an unprivileged service account, grant raw or datagram ICMP access for native ICMP probing:

```bash
# Option A: Grant CAP_NET_RAW to the binary
sudo setcap cap_net_raw+ep /usr/local/bin/vmlg

# Option B: Enable unprivileged ICMP datagram sockets in kernel
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

---

## HTTP & SSE API Overview

| Endpoint | Description |
| --- | --- |
| `GET /healthz` | Liveness and readiness check |
| `GET /metrics` | Prometheus metrics exposition (optional `metrics_token` Bearer auth) |
| `GET /api/v1/meta` | Portal metadata, visitor IP/ASN detection, and registered node list |
| `GET /api/v1/nodes` | Live node status, active load, and short-lived HMAC speedtest tickets |
| `GET /api/v1/probe/compare` | SSE streaming multi-node concurrent latency & return-route comparison |
| `GET /api/v1/probe/mtr` | SSE streaming multi-cycle live MTR trace |
| `GET /api/v1/probe/ping` | SSE streaming single-node ICMP / TCP Ping |
| `GET /api/v1/bgp/route` | BGP RIB route lookup (paths, AS-Path, LocalPref, Communities) |
| `GET /api/v1/bgp/summary` | BGP neighbor session summary |
| `POST /api/v1/iperf3/lease` | Allocate an ephemeral one-shot iPerf3 server port lease |
| `GET /api/v1/speedtest/{ping,download}` | HTML5 speedtest latency probe and incompressible download stream |
| `POST /api/v1/speedtest/upload` | HTML5 speedtest zero-disk upload sink |
| `GET /speedtest/files/{name}` | Zero-disk in-memory test files (`10MB.bin`, `100MB.bin`, `1GB.bin`) |

---

## Development & Verification

```bash
make fmt
make vet
make race
make smoke
make build
```

## License

[MIT License](LICENSE) © 2026 cloudapp3
