# vmlg — Современный мультиузловой сетевой и BGP Looking Glass в одном бинарнике

[![CI](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/ci.yml)
[![Docker](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml/badge.svg)](https://github.com/cloudapp3/vmlg/actions/workflows/docker.yml)
[![Live Demo](https://img.shields.io/badge/Live_Demo-lg.bestcheapvps.org-22c55e)](https://lg.bestcheapvps.org)
[![Go Reference](https://pkg.go.dev/badge/github.com/cloudapp3/vmlg.svg)](https://pkg.go.dev/github.com/cloudapp3/vmlg)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](../LICENSE)

[English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja.md) · [Deutsch](README.de.md) · **Русский**

**vmlg** — это современный сетевой и BGP Looking Glass портал на чистом Go (`CGO_ENABLED=0`) без внешних баз данных, Redis и сборки фронтенда на Node.js. Один статический бинарный файл включает многоязычный веб-интерфейс, одиночный режим (`standalone`), центральный портал (`hub`), региональный агент (`agent`) и `stdio` MCP-сервер.

---

## Ключевые возможности

- **Один бинарник, два режима (`standalone` + `hub`/`agent`)**: Мгновенный запуск на одном сервере без конфигурации или объединение десятков локаций через исходящий WebSocket-туннель (работает даже за NAT).
- **Сравнение задержек и обратных маршрутов со всех узлов в один клик**: Автоматически определяет IP и ASN посетителя и параллельно тестирует RTT, джиттер, потери пакетов и маршрут со всех активных площадок.
- **Потоковый MTR и Ping через SSE с определением магистралей и IXP**: Работает даже в непривилегированных контейнерах без `CAP_NET_RAW` через механизм ядра Linux `IP_RECVERR` (`MSG_ERRQUEUE`) и распознаёт магистральных операторов и точки обмена трафиком (**`DE-CIX`, `AMS-IX`, `LINX`, `HKIX`, `Equinix`** и др.).
- **Тестовые файлы без записи на диск, HTML5 Speedtest и аренда портов iPerf3**: Генерирует несжимаемые потоки `10MB.bin` / `100MB.bin` / `1GB.bin` прямо из памяти (`ChaCha8`) и выдаёт временные изолированные порты `iperf3` с автоочисткой по таймеру.
- **Подключаемые Read-Only драйверы BGP и SVG-граф AS-Path**: Поддержка **BIRD 2/3** (нативный Unix-сокет в режиме `restrict`), **FRRouting** и **GoBGP** с визуализацией путей на чистом SVG без Graphviz.
- **Встроенная защита от злоупотреблений (Guard)**: Защита от SSRF / DNS-rebinding и двумерное ограничение частоты запросов (**на клиента** и **на целевую подсеть `/24` / `/48`** против амплификации).

---

## Быстрый старт

```bash
# Установка через Go
go install github.com/cloudapp3/vmlg/cmd/vmlg@latest

# Одиночный запуск
vmlg serve -listen 0.0.0.0:8080 -node-name "Frankfurt · BGP" -region "EU"

# Запуск через Docker Compose (multi-arch образ со встроенным iperf3)
docker compose up -d
```

Полное руководство, параметры конфигурации и описание API доступны в основной английской документации [README.md](../README.md).

## Лицензия

[MIT License](../LICENSE) © 2026 cloudapp3
