# vmlg — Modernes Single-Binary Multi-Node Netzwerk- & BGP-Looking-Glass

[![CI](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml)
[![Docker](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cloudapp3/vmlg.svg)](https://pkg.go.dev/github.com/cloudapp3/vmlg)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](../LICENSE)

[English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja.md) · **Deutsch** · [Русский](README.ru.md)

**vmlg** ist ein selbstgehostetes, in reinem Go (`CGO_ENABLED=0`) geschriebenes Netzwerk- und BGP-Looking-Glass-Portal ohne Node.js-Build-Abhängigkeiten – ideal für Hosting-Anbieter, IXPs, AS-Betreiber, DN42-Netzwerke und Multi-Region-Homelabs. Eine einzige statische Binärdatei enthält Web-UI, `standalone`-Server, zentrales `hub`-Portal, leichtgewichtigen `agent` sowie einen `stdio` MCP-Server – ganz ohne Datenbank oder Redis.

---

## Hauptfunktionen

- **Single-Binary Dual-Mode (`standalone` + `hub`/`agent`)**: Sofort einsatzbereit auf einem einzelnen Host oder als verteiltes Multi-Region-Portal über ausgehende WebSocket-Tunnel (ideal auch hinter NAT/Firewalls).
- **Multi-Node Latenz- & Rückrouten-Vergleich per Klick**: Erkennt automatisch die Besucher-IP/ASN und testet parallel von allen Online-Standorten Latenz, Jitter, Paketverlust und Carrier-Routing.
- **Echtzeit-SSE-MTR & Ping mit Carrier-/IXP-Erkennung**: Nutzt unprivilegiertes Linux `IP_RECVERR` (`MSG_ERRQUEUE`) als Fallback in Rootless-Containern und erkennt u. a. **`DE-CIX` / `AMS-IX` / `LINX` / `HKIX` / `Equinix`** sowie internationale Tier-1- und Asien-Backbones.
- **Zero-Disk-Speedtest & temporäre iPerf3-Port-Leases**: Streamt nicht komprimierbare `ChaCha8`-Testdateien (`10MB.bin`, `100MB.bin`, `1GB.bin`) direkt aus dem RAM, bietet einen HTML5-Speedtest und vergibt zeitlich begrenzte One-Shot-`iperf3`-Ports.
- **Pluggable Read-Only BGP & reines SVG AS-Path-Diagramm**: Unterstützt **BIRD 2/3** (nativer Unix-Socket im `restrict`-Modus), **FRRouting** (`vtysh` JSON) und **GoBGP** inklusive clientseitiger SVG-Topologie ohne Graphviz.
- **Mehrstufiger Missbrauchsschutz (Guard)**: Kein Shell-Aufruf, strikter SSRF-/DNS-Rebinding-Schutz sowie zweidimensionale Rate-Limits (**pro Client** und **pro Ziel-`/24`- bzw. `/48`-Subnetz**).

---

## Schnellstart

```bash
# Installation über Go Toolchain
go install github.com/cloudapp3/vmlg/cmd/vmlg@latest

# Einzelknoten starten
vmlg serve -listen 0.0.0.0:8080 -node-name "Frankfurt · DE-CIX" -region "EU"

# Start mit Docker Compose (Multi-Arch inkl. vorinstalliertem iperf3)
docker compose up -d
```

Die vollständige Dokumentation und API-Übersicht finden Sie im englischen Hauptdokument [README.md](../README.md).

## Lizenz

[MIT License](../LICENSE) © 2026 cloudapp3
