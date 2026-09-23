# Whale Tracker

A high-performance, multi-chain crypto wallet tracker and market intelligence platform built with Go, gRPC, grpc-gateway, and React.

## Features

- **Multi-Chain Balance Tracking:** Monitor native balances and contract token holdings across BTC, ETH, Solana, Tron, Polygon, Arbitrum, Base, XRP, and custom tokens.
- **Config-Driven Token Registry:** Declarative asset and contract address configuration via `config.yaml`.
- **Market Data & Caching:** CoinGecko integration paired with Redis caching for real-time price feeds and metadata lookup.
- **Event-Driven Messaging:** Broadcast live balance updates and price changes through RabbitMQ.
- **Automated Alerts:** User-configurable alert triggers for price shifts and balance movements.

## Project Structure

```text
wallet/
├── .github/workflows/  # CI/CD deployment pipelines
├── backend/            # Go services, DB migrations, Docker & K8s configs
├── docs/               # System architecture, deployment guides, and tech notes
├── frontend/           # React + TypeScript + Vite web app
├── proto/              # Protocol buffer definitions (Buf)
└── tools/              # Bruno, scripts and tooling
```

## Tech Stack

- **Backend:** Go 1.22+, gRPC, `grpc-gateway` v2, Protocol Buffers / Buf
- **Database & Messaging:** PostgreSQL, Redis, RabbitMQ
- **Authentication & Ops:** Keycloak (OIDC), Caddy, Docker Compose, Kubernetes
- **Observability:** Prometheus, Grafana
- **Frontend:** React, TypeScript, Vite

## Quick Start

### 1. Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Node.js & npm
- [Buf CLI](https://buf.build/)

### 2. Local Infrastructure & Backend

```bash
cd backend

# Copy environment variables
cp .env.dev.example .env

# Start backing services (Postgres, Redis, RabbitMQ, Keycloak)
docker compose -f compose.dev.yml up -d

# Start backend service (database migrations run automatically on startup)
go run main.go
```

The backend exposes:
- **gRPC Server:** Port configured in `config.yaml`
- **HTTP/REST Gateway:** Port configured in `config.yaml`
- **Swagger Documentation:** Available via the HTTP docs handler

### 3. Frontend Setup

```bash
cd frontend
npm install
npm run dev
```
![1](/docs/assets/demo.png)

## Protocol Buffer Generation

Re-generate Go, gRPC, gateway, and TypeScript code from `.proto` definitions:

```bash
buf dep update
buf generate
```

## Documentation

- 📐 **[Architecture Overview](docs/ARCHITECTURE.md):** System design, core services, and event-driven workflows.
- 🚀 **[Deployment Guide](docs/DEPLOYMENT.md):** Kubernetes configuration, secrets management, CI/CD, and troubleshooting.
- 🛠️ **[Technical Notes](docs/TECH_NOTES.md):** Manual CLI workflows, DB migration commands, and dev setup tips.

## API Documentation & Testing

- **Swagger / OpenAPI:** Served locally at `http://localhost:<http-port>/docs` when the backend is running.
- **Bruno Collection:** A pre-configured collection with all gRPC-gateway REST endpoints is available in `tools/bruno/`.