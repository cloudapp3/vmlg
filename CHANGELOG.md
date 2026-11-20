# Changelog

All notable changes to **vmlg** will be documented in this file.

## [Unreleased]

### Added

- **DNS Dig Probe**: `GET /api/v1/probe/dig` with full dig semantics — A/AAAA/CNAME/MX/NS/TXT/PTR record queries showing per-record TTLs, CNAME chains and the complete answer section, via the node system resolver or a user-supplied resolver (SSRF-guarded, dual token-bucket rate limited, UDP with automatic TCP retry on truncation). New `vmlg_dig` MCP tool, Web UI dig tab with deep-link support (`?tab=dig&target=&rrtype=&resolver=`), and hub→agent tunnel forwarding.
- **HTTP/HTTPS Probe**: `GET /api/v1/probe/http` controlled GET reachability probe — status code, TTFB, full redirect chain (per-hop URL/status/latency), TLS certificate summary (issuer/validity/SANs), and complete response headers. Response body is capped at 1 MiB for timing only and never returned or rendered; per-hop SSRF re-checks on every redirect; new `vmlg_http` MCP tool, Web UI http tab (`?tab=http&url=`), tunnel forwarding, and `[probe] http_timeout_ms`.
- **Node Health Panel**: heartbeats now carry a coarse per-node snapshot (load 1/5/15, memory %, agent uptime, version) surfaced on node cards, `/api/v1/nodes`, `vmlg_list_nodes` (MCP), and a new per-node `vmlg_node_up` Prometheus gauge. A zero-disk in-memory ring buffer keeps one hour of load samples (15s default) for card sparklines — restart clears it. `[node_status] public = false` restricts the public surface to online/offline + version.

## [v0.1.0] - 2026-11-18

### Added

- **Single-Binary Dual-Mode Architecture**: `standalone` single-node server, central `hub` portal, and outbound WebSocket `agent` daemon in one pure-Go (`CGO_ENABLED=0`) binary.
- **Multi-Node Concurrent Comparison**: One-click latency, packet loss, and return-route backbone comparison across all online regional nodes via SSE (`/api/v1/probe/compare`).
- **Streaming MTR & Ping Engine**: Native Go ICMPv4/v6 and TCP Ping, multi-cycle live MTR, automatic Linux `IP_RECVERR` (`MSG_ERRQUEUE`) unprivileged container fallback, and carrier/IXP hop enrichment (`CN2 GIA/GT`, `CU 9929`, `CMIN2`, `CT 163`, `CU 4837`, `CMI`, `IIJ`, `SoftBank`, `HKIX`, `DE-CIX`, `AMS-IX`, `Equinix`, `JPIX`, `Any2IX`).
- **Zero-Disk Speedtest & Ephemeral iPerf3 Pool**: In-memory `ChaCha8` incompressible test files (`10MB.bin`, `100MB.bin`, `1GB.bin`), multi-stream HTML5 download/upload/ping/jitter benchmark, HMAC speedtest tickets, and TTL-bounded one-shot `iperf3` port leases.
- **Pluggable Read-Only BGP Looking Glass**: Native BIRD 2/3 Unix socket (`restrict` mode), FRRouting (`vtysh` JSON), GoBGP, and `mock` drivers with client-side pure SVG AS-Path topology rendering and BGP Community decoding.
- **Anti-Abuse Security Guard**: Strict RFC 1123 & `net/netip` validation, SSRF/DNS-rebinding protection, trusted proxy header extraction, and dual-dimension token buckets (`Per-Client` + `Per-Target /24 & /48`).
- **Ecosystem & Deployment**: Prometheus `/metrics` endpoint, `vmlg mcp` stdio JSON-RPC server (local & remote `-hub-http`), `vmlg update` SHA-256 verified self-updater, multi-arch Docker image (`ghcr.io/cloudapp3/vmlg` with `iperf3` pre-installed), Docker Compose stacks, systemd units, and 6-language README documentation.
