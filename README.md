## Crypto Wallet Tracker

A Go-based backend for tracking crypto prices, wallet activity, and future alerting workflows. The project is still in its early stages, but the foundation is already in place for price collection, caching, persistence, and API exposure.

This repository is built as a backend-first project with plans to expand into a full web and mobile experience.

### What it does

Right now, the service includes:

- A Go HTTP server with versioned API endpoints
- Background price ingestion for the top 250 coins from CoinGecko
- Redis and PostgreSQL caching/persistence for recent price snapshots and metadata
- Wallet tracking across chains such as SOL, ETH, TRX, ADA, and BTC
- Docker Compose support for local infrastructure

### Current goals

The project is aimed at becoming a crypto monitoring platform for:

- wallet tracking
- whale movement monitoring
- price alerts
- future mobile/web app integration

### Tech stack

- Go
- Gin Web Framework
- PostgreSQL
- Redis
- Docker Compose
- Viper for configuration
- Logrus for logging

### Project structure

```text
.
├── cmd/
├── config/
├── internal/
│   ├── api/
│   ├── cache/
│   ├── client/
│   ├── core/
│   └── db/
├── bootstrap/
├── main.go
├── docker-compose.yml
└── config.yaml
```

### Getting started

### Prerequisites

- Go 1.26+
- Docker and Docker Compose

### 1. Start infrastructure

```bash
docker compose up -d postgres redis
```

### 2. Run the app

```bash
go run .
```

The server will start on port 8080 by default.

## API

The backend exposes a versioned API under `/api/v1`.

### Public endpoints

- `GET /api/v1/coins`
  - List available supported coins and tokens.
- `GET /api/v1/coins/:id`
  - Get metadata for a specific coin.
- `GET /api/v1/prices`
  - Retrieve current price data for tracked assets.
- `GET /api/v1/prices/:id`
  - Retrieve current price data for a specific asset.
- `GET /health`
  - Check service health and uptime.

### Wallet endpoints (protected)

The wallet endpoints require bearer token authentication via the configured identity provider.

- `GET /api/v1/wallets`
  - List saved wallets.
- `POST /api/v1/wallets`
  - Add a new wallet to the tracker.
- `PUT /api/v1/wallets`
  - Update wallet details.
- `GET /api/v1/wallets/balance`
  - Fetch aggregated wallet balances.
- `DELETE /api/v1/wallets`
  - Remove a wallet from tracking.

#### Add wallet example

```json
{
  "chain": "sol",
  "address": "CFMQzGS8M8wpvcWs1udJ2XgzXEmVf31bmYxBwxxxxxx",
  "token_symbol": "XAUT",
  "label": "XAUt0"
}
```

#### Wallet list response example

```json
{
  "wallet": [
    {
      "id": "b2318e12-4c93-4c04-b6c8-e7f6c84a7f98",
      "address": "DDcdDmDPYw595wAR1jYNHZQTFNi8BGisd2bVa3xxxxxx",
      "chain": "SOL",
      "token_symbol": "GOOGLX",
      "label": "test",
      "created_at": "2026-07-20T21:56:40.130046+03:00",
      "updated_at": "2026-07-20T21:56:40.130046+03:00",
      "balance_crypto": 15.5829108,
      "balance_usd": 5501.390828832001,
      "change_24h_percent": 1.44692
    },
    {
      "id": "bc78b8be-96cf-4d0d-bb97-7db100ce9d07",
      "address": "CFMQzGS8M8wpvcWs1udJ2XgzXEmVf31bmYxBxxxxxx",
      "chain": "SOL",
      "token_symbol": "XAUT",
      "label": "XAUt0",
      "created_at": "2026-07-20T22:30:45.995933+03:00",
      "updated_at": "2026-07-20T22:30:45.995933+03:00",
      "balance_crypto": 0.006881,
      "balance_usd": 27.60223697,
      "change_24h_percent": 0.11673
    }
  ],
  "total": 2,
  "total_balance_usd": 5528.993065802001
}
```

## Configuration

The application uses a YAML config file. You can adjust settings in [config.yaml](config.yaml) for:

- server port
- database connection details
- Redis settings
- Alchemy API access

## Roadmap

### Phase 1
- finish core price ingestion
- improve API endpoints
- add wallet-related models and handlers

### Phase 2
- alert system for price thresholds and whale activity
- better persistence and query support

### Phase 3
- build the app experience on top of the backend

## Notes

This project is still early and evolving. The focus right now is on building a solid backend foundation before adding the broader product experience.

## License

This project is currently under active development. A license will be added as the project matures.


# Database migration
```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate -path $PWD/internal/db/migrations -database "postgres://tracker_admin:super_secure_password@localhost:5432/whale_tracker?sslmode=disable" up
```