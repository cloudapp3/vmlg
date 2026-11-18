# vmlg — 单二进制多节点网络 Looking Glass

[![CI](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cloudapp3/vmlg.svg)](https://pkg.go.dev/github.com/cloudapp3/vmlg)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**vmlg** 是一个使用纯 Go 编写、零前端构建依赖的现代网络 Looking Glass 门户。单个 `CGO_ENABLED=0` 静态二进制文件同时内嵌响应式 Web UI、单机服务（`standalone`）、多节点聚合中心（`hub`）与轻量节点探针（`agent`），无需任何外部数据库或 Redis。

English Documentation: [docs/README.en.md](docs/README.en.md)

---

## 核心特性

- **单二进制双模架构（Standalone + Hub/Agent）**
  - 单台 VPS 一条命令即开即用：`vmlg serve --listen :8080`
  - 多地域机房通过 `vmlg agent` 主动建立反向 WebSocket 长连接注册至 `vmlg hub`，Agent 即使位于 NAT 后或无独立管理域名也能接入诊断控制面。
- **多节点同屏并发横评（买家选机房利器）**
  - 自动识别访客公网 IP 与运营商属性，点击一次「测我本机」，所有在线节点并发探测到访客 IP 的延迟、抖动、丢包率及回程路由，按延迟实时排序同屏展示。
- **实时流式 MTR / Ping + 精品骨干网与 IXP 识别**
  - 基于 SSE（Server-Sent Events）实时流式推送，自动设置 `X-Accel-Buffering: no` 穿透 Nginx / CDN 反代缓冲。
  - 原生支持 `ICMP` / `TCP` Ping 与多轮次实时 MTR，并在无 `CAP_NET_RAW` 的受限容器（如 Rootless Docker）中自动降级至 Linux 内核 `IP_RECVERR`（`MSG_ERRQUEUE`）无特权路由追踪引擎。
  - 内置离线前缀与 ASN/rDNS 识别规则，自动标记 **`CN2 GIA` / `CN2 GT` / `CU 9929` / `CU 10099` / `CMIN2` / `CT 163` / `CU 4837` / `CMI` / `IIJ` / `SoftBank`** 及 **`HKIX` / `DE-CIX` / `AMS-IX` / `Equinix` / `JPIX`** 等交换中心跳点。
- **零磁盘占用流式测速包 + HTML5 交互测速 + 受控 iPerf3 租约池**
  - `/speedtest/files/{10MB,100MB,1GB}.bin` 由内存预置的 `ChaCha8` 伪随机不可压缩数据块实时流式输出，**磁盘占用为 0**，杜绝透明压缩作弊，并支持单流带宽上限流控。
  - 内置多线程 HTML5 下载 / 上传（服务端黑洞丢弃零内存膨胀）/ Ping / Jitter 交互测速图表。
  - **iPerf3 临时端口租约池**：按需从端口池（如 `5201-5210`）分配独占端口启动单次超时回收的 `iperf3 -s -1` 子进程，彻底避免固定公网 iPerf3 端口被长期打满。
- **默认防滥用安全护栏（Guard）**
  - **零 Shell 拼接**：所有探测参数经 `net/netip` 与 RFC 1123 严格校验，杜绝命令注入。
  - **SSRF 与 DNS 重绑定防御**：默认拦截环回、RFC1918 私网、CGNAT、链路本地及云元数据地址 `169.254.169.254`（DN42 等场景可通过 `allowed_reserved_cidrs` 显式放行）。
  - **双维度令牌桶限流（Per-Client + Per-Target）**：同时限制单访客请求速率与**发往同一目标 `/24`（IPv6 `/48`）网段**的全局探测速率，防止多节点沦为针对单一受害者的反射放大器。
- **可插拔 BGP 路由表与纯 SVG AS-Path 拓扑图**
  - 按需开启只读 BGP 驱动：支持 **BIRD 2/3**（原生直连 `/run/bird/bird.ctl` Unix Socket + `restrict` 只读锁定，零 fork 开销）、**FRRouting**（`vtysh json`）、**GoBGP** 及演示 `mock` 模式。
  - 前端零 Graphviz 依赖渲染 **纯 SVG AS-Path 拓扑有向图**，支持 Best/Backup 多路径对比、Standard / Large Communities 自定义字典解码与邻居会话（Peers）状态总览。
- **可观测性与 AI 诊断生态（Prometheus + MCP Server）**
  - 原生内置 `/metrics` Prometheus 指标接口（支持可选 Bearer Token 鉴权）。
  - 内置 `vmlg mcp` Stdio JSON-RPC 服务，既可本地直连运行，也可通过 `-hub-http` 远程连接多机房 Hub 供大模型客户端调用（`vmlg_list_nodes` / `vmlg_ping` / `vmlg_mtr` / `vmlg_compare` / `vmlg_bgp_route`）。
- **中英双语 + 暗色/亮色主题 + 论坛脱敏 Markdown 一键导出**
  - 一键生成适合贴入论坛 / GitHub Issue 的结构化 Markdown 报告，默认自动掩码目标公网 IP 末段（如 `203.0.113.x`）。

---

## 快速开始

### 1. 安装与构建

```bash
# 直接通过 Go 工具链安装（零 Node.js / CGO 依赖）
go install github.com/cloudapp3/vmlg/cmd/vmlg@latest

# 或从源码编译静态二进制
make build
```

### 2. 单机模式（Standalone，零配置即跑）

```bash
./vmlg serve \
  -listen 0.0.0.0:8080 \
  -node-id tyo-01 \
  -node-name "Tokyo · BGP Premium" \
  -location "Tokyo, JP" \
  -region "APAC"
```

浏览器访问 `http://127.0.0.1:8080` 即可使用。完整配置可参考 [examples/standalone.toml](examples/standalone.toml)。

### 3. 多机房集群模式（Hub + Agent）

**在门户节点启动 Hub：**

```bash
export VMLG_AGENT_TOKEN="replace-with-strong-random-token"
./vmlg hub \
  -listen 0.0.0.0:8080 \
  -title "Global Cloud Looking Glass" \
  -enable-local-node \
  -node-id hkg-01 \
  -node-name "Hong Kong 01" \
  -region "APAC"
```

**在各地域 VPS 节点启动 Agent：**

```bash
export VMLG_AGENT_TOKEN="replace-with-strong-random-token"
./vmlg agent \
  -hub-url ws://hub.example.com:8080/api/v1/cluster/ws \
  -node-id lax-01 \
  -node-name "Los Angeles · CN2 GIA" \
  -location "Los Angeles, US" \
  -region "NA" \
  -tags "CN2 GIA,10Gbps" \
  -speedtest-listen 0.0.0.0:8081 \
  -public-url https://lax-speed.example.com
```

完整多节点配置示例见 [examples/hub.toml](examples/hub.toml) 与 [examples/agent.toml](examples/agent.toml)，上线前可执行校验：

```bash
./vmlg check -config ./examples/standalone.toml
```

### 4. Docker 与 Docker Compose 部署

官方多架构镜像（`linux/amd64`、`linux/arm64`）自动发布至 `ghcr.io/cloudapp3/vmlg:latest`，镜像内已预装 `iperf3` 并以非特权用户（UID `10001`）运行。

**方式 A：Docker Compose 一键单机部署**

```bash
docker compose up -d
```

**方式 B：`docker run` 纯环境变量启动（单机模式）**

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

> **提示**：如需探测宿主机原生 IPv6 路由或挂载宿主机 BIRD Socket，可改用 `--network host` 并附加 `-v /run/bird/bird.ctl:/run/bird/bird.ctl:ro -e VMLG_BGP_DRIVER=bird`。

**方式 C：多节点 Hub + Agent 容器化编排**

见 [docker-compose.cluster.yml](docker-compose.cluster.yml)（支持环境变量与 `*_FILE` Docker Secrets 注入，例如 `VMLG_AGENT_TOKEN_FILE`）：

```bash
# 启动中心门户 Hub（含本地节点）与测试 Agent
VMLG_AGENT_TOKEN="replace-with-strong-token" \
  docker compose -f docker-compose.cluster.yml --profile agent up -d
```

### 5. MCP Server（AI 辅助多节点网络排障）

```bash
# 本地单机诊断模式（无需监听端口）
vmlg mcp

# 远程连接已部署的多节点 Hub 门户
vmlg mcp -hub-http https://lg.example.com
```

---

## Linux 非特权运行建议

`vmlg` 在无特权容器环境下已内置内核 `IP_RECVERR` 路由追踪与无特权连接探测回退；在生产裸机或虚拟机以普通用户运行 systemd 服务时，赋予 `CAP_NET_RAW` 或开启 `ping_group_range` 可启用纯正原生 ICMP 收发引擎：

```bash
# 方式 A：允许二进制使用原生 ICMP 套接字
sudo setcap cap_net_raw+ep /usr/local/bin/vmlg

# 方式 B：开启内核非特权 ICMP Datagram 支持
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

---

## HTTP / SSE 接口概览

| 方法与路径 | 说明 |
| --- | --- |
| `GET /healthz` | 存活探测 |
| `GET /metrics` | Prometheus 指标导出（支持 `metrics_token` Bearer 鉴权） |
| `GET /api/v1/meta` | 门户元信息、访客 IP/ASN 识别与在线节点列表 |
| `GET /api/v1/nodes` | 节点实时在线状态、负载与短期 HMAC 测速凭据 |
| `GET /api/v1/probe/compare` | SSE 流式多节点并发延迟与回程线路横评 |
| `GET /api/v1/probe/mtr` | SSE 流式单节点多轮次实时 MTR 路由追踪 |
| `GET /api/v1/probe/ping` | SSE 流式单节点 ICMP / TCP Ping 探测 |
| `GET /api/v1/bgp/route` | 查询节点 BGP 路由表（RIB 路径、AS-Path、Communities） |
| `GET /api/v1/bgp/summary` | 查询节点 BGP 邻居会话状态 |
| `POST /api/v1/iperf3/lease` | 申请节点临时独占 iPerf3 服务端端口租约 |
| `GET /api/v1/speedtest/{ping,download}` | HTML5 测速延迟与不可压缩下载流 |
| `POST /api/v1/speedtest/upload` | HTML5 测速上传接收端（零落盘丢弃） |
| `GET /speedtest/files/{name}` | 零磁盘内存流式测速包（`10MB.bin` / `100MB.bin` / `1GB.bin`） |

---

## 开发与验证

```bash
make fmt
make vet
make race
make smoke
make build
```

## 开源协议

[MIT License](LICENSE) © 2026 cloudapp3
