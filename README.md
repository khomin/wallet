# Whale Tracker

A crypto wallet and market intelligence platform built with Go, gRPC, grpc-gateway, and a React frontend. The backend now exposes both native gRPC services and HTTP REST endpoints generated from protobuf definitions under [proto](proto).

## What this project does

- Track wallet balances across multiple blockchains and token deployments
- Aggregate live price data from providers such as CoinGecko and chain-specific RPCs
- Keep token metadata in a config-driven registry for native assets and contracts
- Publish price and wallet events through RabbitMQ
- Cache price and wallet state in Redis
- Trigger alert workflows and email notifications
- Expose a REST API through grpc-gateway alongside the gRPC server

## Current architecture

The backend is organized around a service-oriented model with these major pieces:

- HTTP + gRPC gateway entrypoint in [backend/main.go](backend/main.go)
- Protocol definitions in [proto](proto)
- Generated Go code in [backend/gen](backend/gen)
- API handlers in [backend/internal/api/handlers](backend/internal/api/handlers)
- Core domain services in [backend/internal/core](backend/internal/core)
- Repositories and Postgres models in [backend/internal/db](backend/internal/db)
- Redis and message bus integration in [backend/internal/cache](backend/internal/cache) and [backend/internal/messaging](backend/internal/messaging)
- Frontend app in [web](web)

### Service overview

- Price service: token catalogs, market prices, streaming refreshes, metadata lookup
- Wallet service: wallet create/list/edit/delete flows and balance retrieval
- User service: users and identity-related flows
- Alert service: user alerts and notifications
- Blockchain service: ETH, Arbitrum, Base, Polygon, BNB, Solana, Bitcoin, Tron, XRP integrations

## Tech stack

- Go
- gRPC
- grpc-gateway v2
- Protocol Buffers / Buf
- PostgreSQL
- Redis
- RabbitMQ
- Keycloak
- Docker Compose
- Vite + React + TypeScript
- Swagger/OpenAPI generation from proto

## Project layout

```text
.
├── backend/
│   ├── config.yaml
│   ├── main.go
│   ├── bootstrap/
│   ├── cmd/
│   ├── gen/
│   ├── internal/
│   ├── deploy/
│   └── docker-compose.yml
├── proto/
│   ├── alert/
│   ├── price/
│   ├── user/
│   └── wallet/
├── web/
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── buf.yaml
├── buf.gen.yaml
├── resources/
└── README.md
```

## Supported chains and assets

The app is built for multi-chain wallet tracking and token-aware balances. The current configuration includes native assets and token deployments for chains such as:

- BTC
- ETH
- ARB
- SOL
- TRX
- POL / Polygon
- BNB
- XRP
- various token deployments such as USDC, USDT, WBTC, PAXG, RLUSD, and xStocks

The active token registry is configured in [backend/config.yaml](backend/config.yaml).

## gRPC and HTTP API

The project uses protobuf definitions as the source of truth. The generated code lives under [backend/gen](backend/gen) and the source files live under [proto](proto).

### Proto packages

- [proto/price/v1/price.proto](proto/price/v1/price.proto)
- [proto/wallet/v1/wallet.proto](proto/wallet/v1/wallet.proto)
- [proto/user/v1/user.proto](proto/user/v1/user.proto)
- [proto/alert/v1/alert.proto](proto/alert/v1/alert.proto)

### HTTP routes exposed via grpc-gateway

These are generated from the proto annotations and served by the gateway on the HTTP port configured in [backend/config.yaml](backend/config.yaml).

#### Price routes

- `GET /v1/coins`
- `GET /v1/coins/{id}`
- `POST /v1/coins/search`
- `GET /v1/prices`
- `GET /v1/prices/{symbol}`
- `GET /v1/prices/stream`

#### Wallet routes

- `GET /v1/wallets`
- `GET /v1/wallets/{id}`
- `POST /v1/wallets`
- `PATCH /v1/wallets/{id}`
- `DELETE /v1/wallets/{id}`

#### Additional service routes

The user and alert services are also exposed via grpc-gateway, with their generated handlers registered in [backend/main.go](backend/main.go).

## Local development

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- Node.js + npm
- Buf CLI

### Start infrastructure

```bash
cd backend
docker compose up -d
```

This starts the backing services used by the app, including Postgres, Redis, RabbitMQ, and the configured Bitcoin/Keycloak setup shown in the compose config and deployment assets.

### Start the backend

```bash
cd backend
go run .
```

The service exposes:

- gRPC on the configured gRPC port
- HTTP/gateway on the configured HTTP port
- Swagger docs under the docs handler

### Start the frontend

```bash
cd web
npm install
npm run dev
```

## Configuration

The main runtime config is [backend/config.yaml](backend/config.yaml). It includes:

- HTTP and gRPC server ports
- Postgres DSN settings
- Redis and RabbitMQ connection info
- Email settings
- RPC endpoints for supported chains
- Keycloak authorization settings
- Token registry definitions

Before using Tron support, ensure you have a valid TronGrid API key configured in the blockchain section.

## Proto generation

The project generates Go, gRPC, gateway, OpenAPI, and TypeScript artifacts from the `.proto` definitions.

```bash
# Install Buf
# macOS
brew install bufbuild/buf/buf

# Linux / WSL
# or
# go install github.com/bufbuild/buf/cmd/buf@latest

buf dep update
buf generate
```

The generation config is defined in [buf.gen.yaml](buf.gen.yaml) and the source configuration is in [buf.yaml](buf.yaml).

## Database migration notes

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate -path $PWD/backend/internal/db/migrations \
  -database "postgres://tracker_admin:super_secure_password@localhost:5432/whale_tracker?sslmode=disable" up
```

## Notes

- This project has grown beyond the initial single-service wallet tracker and now includes a full protocol-driven API surface.
- The actual runtime is a combined gRPC + gateway service, not a plain HTTP API layer.
- The gRPC gateway is the main REST-compatible entrypoint for frontend and external clients.

## Roadmap

- expand wallet analytics and tracking jobs
- improve alerting pipelines and notifications
- strengthen multi-chain coverage and token metadata accuracy
- continue hardening the API and operational workflows around Redis, RabbitMQ, and Postgres
