# vmlg — Modern Single-Binary Multi-Node Network Looking Glass

[![CI](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cloudapp3/vmlg.svg)](https://pkg.go.dev/github.com/cloudapp3/vmlg)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](../LICENSE)

**vmlg** is a pure-Go (`CGO_ENABLED=0`), zero-Node-build Looking Glass portal for VPS providers, BGP operators, and multi-region homelabs. One static binary embeds the web UI, `standalone` server, central `hub`, and lightweight regional `agent`—with no database or Redis required.

简体中文文档：[README.md](../README.md)

## Highlights

- **Single Binary, Dual-Mode (`standalone` + `hub`/`agent`)**: Run on one host with zero config, or connect N regional agents outbound over WebSocket to one unified portal.
- **One-Click Multi-Node Compare**: Automatically detects visitor IP/ASN and probes latency, packet loss, and return-route classification concurrently across all online regions.
- **Streaming MTR & Ping with Carrier/IXP Enrichment**: Real-time SSE streaming with unprivileged Linux `IP_RECVERR` (`MSG_ERRQUEUE`) fallback and built-in recognition for `CN2 GIA/GT`, `CU 9929`, `CMIN2`, `CT 163`, `CU 4837`, `CMI`, `IIJ`, `SoftBank`, `HKIX`, `DE-CIX`, `AMS-IX`, `Equinix`, and more.
- **Zero-Disk Speedtest & Ephemeral iPerf3 Leases**: Streams incompressible ChaCha8 PRNG test files (`10MB.bin`, `100MB.bin`, `1GB.bin`) straight from memory, runs multi-stream HTML5 download/upload benchmarks, and leases temporary one-shot `iperf3` ports with automatic TTL expiry.
- **Hardened Anti-Abuse Guard**: Zero shell execution, strict DNS-rebinding/SSRF protection, and dual-dimension token buckets (**Per-Client** + **Per-Target `/24` & `/48` subnet** rate limiting).
- **Pluggable BGP RIB & Pure SVG AS-Path Graph**: Read-only drivers for **BIRD 2/3** (native Unix control socket with `restrict` mode), **FRRouting** (`vtysh` JSON), **GoBGP**, and demo `mock` mode, with client-side SVG AS-Path DAG rendering and BGP Community decoding.
- **Prometheus `/metrics` & Stdio MCP Server**: Built-in telemetry endpoint and `vmlg mcp` stdio server (`vmlg_list_nodes`, `vmlg_ping`, `vmlg_mtr`, `vmlg_compare`, `vmlg_bgp_route`) supporting both local execution and remote `-hub-http` multi-node portals.

## Quick Start

```bash
go install github.com/cloudapp3/vmlg/cmd/vmlg@latest

# Standalone mode
vmlg serve -listen 0.0.0.0:8080 -node-name "Tokyo BGP" -region "APAC"
```

See [examples/standalone.toml](../examples/standalone.toml), [examples/hub.toml](../examples/hub.toml), and [examples/agent.toml](../examples/agent.toml) for production configurations.
