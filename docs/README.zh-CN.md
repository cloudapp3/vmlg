# vmlg — 单二进制多节点网络与 BGP Looking Glass

[![CI](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml)
[![Docker](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml)
[![Docs](https://img.shields.io/badge/docs-vmlg.bestcheapvps.org-0ea5e9)](https://vmlg.bestcheapvps.org/zh/)
[![Live Demo](https://img.shields.io/badge/Live_Demo-lg.bestcheapvps.org-22c55e)](https://lg.bestcheapvps.org)
[![Telegram](https://img.shields.io/badge/Telegram-交流群-26A5E4?logo=telegram)](https://t.me/VMPulse)
[![Go Reference](https://pkg.go.dev/badge/github.com/cloudapp3/vmlg.svg)](https://pkg.go.dev/github.com/cloudapp3/vmlg)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](../LICENSE)

[English](../README.md) · **简体中文** · [繁體中文](README.zh-TW.md) · [日本語](README.ja.md) · [Deutsch](README.de.md) · [Русский](README.ru.md)

**vmlg** 是一个使用纯 Go 编写、零前端构建依赖的现代网络与 BGP Looking Glass 门户。单个 `CGO_ENABLED=0` 静态二进制文件同时内嵌多语言响应式 Web UI、单机服务（`standalone`）、多节点聚合中心（`hub`）、轻量节点探针（`agent`）与 Stdio MCP 服务，无需任何外部数据库或 Redis。

---

## 核心特性

- **单二进制双模架构（`standalone` + `hub`/`agent`）**
  - 单台 VPS 一条命令即开即用：`vmlg serve -listen :8080`
  - 多地域机房通过 `vmlg agent` 主动建立反向 WebSocket 长连接注册至 `vmlg hub`，Agent 即使位于 NAT 后或无独立管理域名也能接入诊断控制面。
- **多节点同屏并发横评（买家选机房利器）**
  - 自动识别访客公网 IP 与运营商属性，点击一次「测我本机」，所有在线节点并发探测到访客 IP 的延迟、抖动、丢包率及回程路由，按延迟实时排序同屏展示。
- **实时流式 MTR / Ping + 精品骨干网与 IXP 识别**
  - 基于 SSE（Server-Sent Events）实时流式推送，自动设置 `X-Accel-Buffering: no` 穿透反代缓冲。
  - 原生支持 `ICMP` / `TCP` Ping 与多轮次实时 MTR，并在无 `CAP_NET_RAW` 的受限容器中自动回退至 Linux 内核 `IP_RECVERR`（`MSG_ERRQUEUE`）无特权路由追踪引擎。
  - 内置离线前缀与 ASN/rDNS 识别规则，自动标记 **`CN2 GIA` / `CN2 GT` / `CU 9929` / `CU 10099` / `CMIN2` / `CT 163` / `CU 4837` / `CMI` / `IIJ` / `SoftBank`** 及 **`HKIX` / `DE-CIX` / `AMS-IX` / `Equinix` / `JPIX`** 等交换中心跳点。
- **零磁盘占用流式测速包 + HTML5 交互测速 + 受控 iPerf3 租约池**
  - `/speedtest/files/{10MB,100MB,1GB}.bin` 由内存预置的 `ChaCha8` 伪随机不可压缩数据块实时流式输出，**磁盘占用为 0**，并支持单流带宽上限流控。
  - 内置多线程 HTML5 下载 / 上传（零内存膨胀黑洞接收）/ Ping / Jitter 交互测速图表。
  - **iPerf3 临时端口租约池**：按需从端口池（如 `5201-5210`）分配独占端口启动单次超时回收的 `iperf3 -s -1` 子进程，防止固定端口被长期打满。
- **可插拔 BGP 路由表与纯 SVG AS-Path 拓扑图**
  - 支持 **BIRD 2/3**（原生直连 `/run/bird/bird.ctl` Unix Socket + `restrict` 只读锁定）、**FRRouting**（`vtysh json`）、**GoBGP** 及演示 `mock` 模式。
  - 前端零 Graphviz 依赖渲染 **纯 SVG AS-Path 拓扑有向图**，支持 Best/Alt 多路径对比、Communities 字典解码与邻居会话（Peers）总览。
- **默认防滥用安全护栏（Guard）**
  - **零 Shell 拼接**：所有探测参数经 `net/netip` 与 RFC 1123 严格校验。
  - **SSRF 与 DNS 重绑定防御**：默认拦截环回、私网、CGNAT、链路本地及云元数据地址 `169.254.169.254`（DN42 场景可通过 `allowed_reserved_cidrs` 放行）。
  - **双维度令牌桶限流（Per-Client + Per-Target）**：同时限制单访客请求速率与**发往同一目标 `/24`（IPv6 `/48`）网段**的全局探测速率，防止多节点沦为反射放大器。
- **可观测性、MCP 服务与多语言脱敏分享**
  - 内置 `/metrics` Prometheus 指标导出与 `vmlg mcp` Stdio JSON-RPC 服务。
  - 界面默认英文并支持多语言切换（English / 简体中文 / 繁體中文 / 日本語），支持一键导出公网 IP 末段打码（`203.0.113.x`）的论坛 Markdown 报告。

---

## 快速开始

### 1. 安装与构建

```bash
# 一键脚本安装（校验 SHA-256）
curl -fsSL https://raw.githubusercontent.com/cloudapp3/vmlg/main/install.sh | bash

# 或通过 Go 工具链直接安装（零 Node.js / CGO 依赖）
go install github.com/cloudapp3/vmlg/cmd/vmlg@latest

# 或从源码编译静态二进制
make build
```

### 2. 单机模式（Standalone）

```bash
vmlg serve \
  -listen 0.0.0.0:8080 \
  -node-id tyo-01 \
  -node-name "Tokyo · BGP Premium" \
  -location "Tokyo, JP" \
  -region "APAC"
```

完整配置参考 [examples/standalone.toml](../examples/standalone.toml)。

### 3. 多机房集群模式（Hub + Agent）

```bash
# 启动中心门户 Hub
export VMLG_AGENT_TOKEN="replace-with-strong-random-token"
vmlg hub \
  -listen 0.0.0.0:8080 \
  -title "Global Cloud Looking Glass" \
  -enable-local-node \
  -node-id hkg-01 \
  -node-name "Hong Kong 01" \
  -region "APAC"

# 在各地域节点启动 Agent
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

### 4. Docker 与 Docker Compose 部署

官方多架构镜像（`linux/amd64`、`linux/arm64`）自动发布至 `ghcr.io/cloudapp3/vmlg:latest`，已预装 `iperf3` 并以非特权用户（UID `10001`）运行：

```bash
# 单节点一键启动
docker compose up -d

# 多节点 Hub + Agent 编排启动
VMLG_AGENT_TOKEN="replace-with-strong-token" \
  docker compose -f docker-compose.cluster.yml --profile agent up -d
```

参见 [docker-compose.yml](../docker-compose.yml) 与 [docker-compose.cluster.yml](../docker-compose.cluster.yml)。

### 5. MCP Server（AI 辅助排障）

```bash
vmlg mcp
vmlg mcp -hub-http https://lg.example.com
```

## 相关生态项目（`vm*` 工具矩阵）

| 项目 | 简介 |
| --- | --- |
| [vmbench](https://github.com/cloudapp3/vmbench) | 新 VPS 一键体检工具（硬件跑分、三网回程识别、流媒体解锁、IP 质量） |
| [vminfo](https://github.com/cloudapp3/vminfo) | 跨平台单二进制终端系统监控、Web 仪表盘与 MCP Server |
| [vmflow](https://github.com/cloudapp3/vmflow) | 资源受控的四层 TCP/UDP 端口转发与可嵌入 Go 网络运行时 |
| [vmpulse](https://github.com/VPSMarket/vmpulse) | 自托管多主机实时监控、测评与远程运维控制平台 |
| **vmlg** | 单二进制多节点网络测速与 BGP Looking Glass 门户 |

Telegram 社区交流群：**[t.me/VMPulse](https://t.me/VMPulse)**

## 开源协议

[MIT License](../LICENSE) © 2026 cloudapp3
