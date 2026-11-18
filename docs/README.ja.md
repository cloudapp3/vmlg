# vmlg — シングルバイナリ・マルチノード対応ネットワーク & BGP Looking Glass

[![CI](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml)
[![Docker](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cloudapp3/vmlg.svg)](https://pkg.go.dev/github.com/cloudapp3/vmlg)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](../LICENSE)

[English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · **日本語** · [Deutsch](README.de.md) · [Русский](README.ru.md)

**vmlg** は、Pure Go（`CGO_ENABLED=0`）で実装されたセルフホスト型のネットワーク & BGP Looking Glass ポータルです。単一の静的バイナリにマルチ言語対応 Web UI、単一ノードサーバー（`standalone`）、統合ポータル（`hub`）、軽量エージェント（`agent`）、および AI クライアント向け `stdio` MCP サーバーを内蔵しており、外部データベースや Redis は不要です。

---

## 主な機能

- **シングルバイナリ・デュアルモード構成（`standalone` + `hub`/`agent`）**：1 台のサーバーで設定ファイルなしですぐに起動できるほか、各リージョンの Agent から Hub へ WebSocket トンネルで接続するマルチノード構成にも対応します。
- **マルチノード同時レイテンシ・復路ルート比較**：アクセス元のグローバル IP と ASN を自動検出し、ワンクリックで全オンライン拠点からの RTT・ジッター・パケットロス・復路キャリア判定を一括比較します。
- **リアルタイム SSE MTR / Ping とキャリア・IXP 自動識別**：SSE ストリーミングによりプロキシ配下でもリアルタイム表示。`CAP_NET_RAW` がないコンテナ環境では Linux カーネルの `IP_RECVERR`（`MSG_ERRQUEUE`）に自動フォールバックし、**`IIJ` / `SoftBank` / `NTT` / `KDDI` / `CN2 GIA` / `CU 9929` / `CMIN2`** や **`JPIX` / `BBIX` / `Equinix` / `HKIX` / `DE-CIX`** を自動ラベル付けします。
- **ディスク消費ゼロのスピードテスト & 一時 iPerf3 ポートリース**：メモリ上の非圧縮擬似乱数ストリームから `10MB.bin` / `100MB.bin` / `1GB.bin` を配信し、HTML5 ブラウザ速度測定および TTL 付きワンショット `iperf3` ポート貸出を提供します。
- **読み取り専用 BGP ドライバ & 純 SVG AS-Path トポロジ描画**：**BIRD 2/3**（Unix ソケット `restrict` モード）、**FRRouting**、**GoBGP** に対応し、Graphviz 不要でクライアントサイド SVG トポロジ図と BGP コミュニティ解析を表示します。
- **不正利用防止セキュリティガード**：SSRF / DNS リバインディング防御に加え、送信元 IP と宛先 `/24`（IPv6 `/48`）サブネットの二軸トークンバケット制限により DDoS 反射・増幅攻撃を防止します。

---

## クイックスタート

```bash
# Go ツールチェーンでインストール
go install github.com/cloudapp3/vmlg/cmd/vmlg@latest

# スタンドアロン起動
vmlg serve -listen 0.0.0.0:8080 -node-name "Tokyo · BGP" -region "APAC"

# Docker Compose で起動（linux/amd64・linux/arm64 対応、iperf3 同梱）
docker compose up -d
```

詳細な設定オプションや API リファレンスは英語のメインドキュメント [README.md](../README.md) を参照してください。

## ライセンス

[MIT License](../LICENSE) © 2026 cloudapp3
