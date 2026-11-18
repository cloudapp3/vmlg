# Contributing to vmlg

Thanks for helping improve **vmlg**! Bug reports, carrier/IXP prefix improvements, BGP daemon compatibility fixes, UI refinements, documentation translations, and focused pull requests are all welcome.

Issues and pull requests in English or Chinese (中文 / 英文) are both welcome.

## Before you start

- Prefer opening an issue first for non-trivial architectural or protocol changes.
- Keep pull requests small, focused, and well-tested.
- Avoid introducing new third-party Go or frontend dependencies unless discussed beforehand (`vmlg` intentionally ships as a zero-CGO, zero-Node-build static binary).

## Local verification checklist

Before submitting a PR, ensure all checks pass locally:

```bash
make fmt
make vet
make race
make smoke
node --check internal/web/static/js/app.js
```

## Project structure

- `cmd/vmlg` — CLI entrypoint (`serve`, `hub`, `agent`, `mcp`, `healthcheck`, `update`, `check`, `version`)
- `internal/config` — TOML & environment configuration parser and validator
- `internal/guard` — SSRF/DNS-rebinding filter, trusted proxy client IP extraction, and dual-dimension token buckets
- `internal/probe` — Native ICMP/TCP Ping, multi-cycle MTR, and Linux `IP_RECVERR` unprivileged fallback
- `internal/enrich` — Carrier backbone (`CN2 GIA`, `CU 9929`, `CMIN2`, etc.), ASN, and IXP classifier
- `internal/speedtest` — Zero-disk `ChaCha8` PRNG test files, HTML5 speedtest handlers, and ephemeral `iperf3` port leases
- `internal/bgp` — Read-only `bird`, `frr`, `gobgp`, and `mock` BGP drivers
- `internal/cluster` — Hub-Agent multiplexed WebSocket reverse tunnel
- `internal/mcp` — Zero-dependency stdio JSON-RPC MCP server
- `internal/metrics` — Prometheus exposition collector
- `internal/web` — Embedded zero-build multi-language web UI

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
