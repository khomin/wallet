# Whale Tracker

A crypto wallet tracker built with Go, gRPC and React.

## Key Features

- **Multi-Chain Node Indexing:**  BTC, ETH, Solana, Tron, Polygon, Arbitrum, Base, BSC, and XRP via dedicated gRPC, JSON-RPC, and REST clients.
- **RPC Rate-Limiting & Throttling:** Built-in token-bucket rate limiters (`rps: 2`, burst control) to manage RPC throughput safely across public and private node infrastructure.
- **Config-Driven Asset Registry:** Multi-chain token and contract tracking configured via `config.yaml`.
- **Market Data Engine:** CoinGecko integration coupled with a dual-layer Redis cache for real-time asset pricing, metadata enrichment, and low-latency lookups.
- **Event-Driven Microservices:** Asynchronous RabbitMQ message bus distributing real-time balance movements, price changes, and system alerts to downstream consumers.
- **Custom Alert Engine:** User-configurable rule processor evaluating real-time wallet balance thresholds and sudden market volatility spikes.

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