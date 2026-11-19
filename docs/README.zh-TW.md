# vmlg — 單二進制多節點網路與 BGP Looking Glass

[![CI](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml)
[![Docker](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml)
[![Live Demo](https://img.shields.io/badge/Live_Demo-lg.bestcheapvps.org-22c55e)](https://lg.bestcheapvps.org)
[![Go Reference](https://pkg.go.dev/badge/github.com/cloudapp3/vmlg.svg)](https://pkg.go.dev/github.com/cloudapp3/vmlg)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](../LICENSE)

[English](../README.md) · [简体中文](README.zh-CN.md) · **繁體中文** · [日本語](README.ja.md) · [Deutsch](README.de.md) · [Русский](README.ru.md)

**vmlg** 是一款以純 Go（`CGO_ENABLED=0`）開發、零前端建置依賴的現代化網路與 BGP Looking Glass 門戶，專為 VPS 主機商、IXP 交換中心、AS 營運者、DN42 社群與跨區自建機房打造。單一靜態執行檔同時內嵌多語系 Web UI、單機模式（`standalone`）、多節點中心（`hub`）、輕量探針（`agent`）以及 Stdio MCP 伺服器，無需外部資料庫或 Redis。

---

## 核心特色

- **單一執行檔雙模架構（`standalone` + `hub`/`agent`）**：單機零設定直接啟動，或由各區域 Agent 主動透過反向 WebSocket 連線註冊至 Hub，輕鬆穿透 NAT 環境。
- **多機房同屏並發橫評**：自動辨識訪客 IP 與電信業者，一鍵觸發全球在線節點同步測試延遲、抖動、掉包率與回程路由。
- **即時串流 MTR / Ping 與骨幹網 / IXP 識別**：透過 SSE 即時推送，無 `CAP_NET_RAW` 權限時自動回退至 Linux `IP_RECVERR` 核心追蹤引擎；內建辨識 **`CN2 GIA` / `CN2 GT` / `CU 9929` / `CMIN2` / `CT 163` / `CU 4837` / `CMI` / `IIJ` / `SoftBank` / `HiNet`** 與 **`HKIX` / `DE-CIX` / `AMS-IX` / `Equinix` / `JPIX`**。
- **零磁碟佔用測速檔、HTML5 測速與受控 iPerf3 租約池**：記憶體即時串流不可壓縮 `ChaCha8` 測速檔（`10MB.bin`、`100MB.bin`、`1GB.bin`），並動態分配限時回收的臨時 `iperf3` 連接埠。
- **可插拔唯讀 BGP 驅動與純前端 SVG AS-Path 拓撲圖**：支援 **BIRD 2/3**（原生 Unix Socket + `restrict` 模式）、**FRRouting**、**GoBGP**，無需伺服器端安裝 Graphviz 即可繪製自治系統路徑圖與解析 BGP Communities。
- **雙維度防濫用護欄（Guard）**：嚴格攔截 SSRF 與 DNS 重綁定，同時限制單一訪客與單一目標 `/24`（IPv6 `/48`）網段的探測速率。

---

## 快速開始

```bash
# Go 工具鏈安裝
go install github.com/cloudapp3/vmlg/cmd/vmlg@latest

# 單機啟動
vmlg serve -listen 0.0.0.0:8080 -node-name "Taipei · BGP" -region "APAC"

# Docker Compose 一鍵啟動（支援 linux/amd64 與 linux/arm64，內建 iperf3）
docker compose up -d
```

完整設定與多節點集群部署請參閱英文主文件 [README.md](../README.md)、[examples/standalone.toml](../examples/standalone.toml) 與 [docker-compose.cluster.yml](../docker-compose.cluster.yml)。

## 授權條款

[MIT License](../LICENSE) © 2026 cloudapp3
