### Crypto Wallet Tracker

A Go-based backend for tracking crypto prices, wallet activity, and future alerting workflows. The project is still in its early stages.

The service includes wallet tracking across chains such as:
    `SOL, ETH, TRX, ADA, BTC`

It also supports configured token assets from the registry in [backend/config.yaml](backend/config.yaml), including major tokens such as `USDC`, `USDT`, `WBTC`, `PAXG`, and other chain-specific assets.

Note: Bitcoin wallet tracking is wired to the preconfigured Bitcoin Core node already defined in docker-compose.yml, so starting the infrastructure brings up the node automatically. Keycloak also comes with a preloaded realm configuration imported from [backend/deploy/keycloak/realm-export.json](backend/deploy/keycloak/realm-export.json).

### Dasboard previews
![1](/resources/demo.png)

### Tech stack

- Gin, PostgreSQL, Keycloak, Redis, Docker Compose, Viper

- SDK used:

```
    github.com/btcsuite/btcd
    github.com/ethereum/go-ethereum
    github.com/fbsobreira/gotron-sdk
    github.com/gagliardetto/solana-go
    github.com/blinklabs-io/gouroboros
    github.com/blinklabs-io/cardano-node-api
```

### Prerequisites

- Go 1.26+
- Docker and Docker Compose

### Configuration

The backend configuration is in [backend/config.yaml](backend/config.yaml). Before
using TRON wallet tracking, set `blockchain.tron_api_key` to a valid TronGrid API key (trongrid.io):

```yaml
blockchain:
    tron_grpc: "grpc.trongrid.io:50051"
    tron_api_key: "your-trongrid-api-key"
```

The public RPC endpoints may be rate-limited or require provider-specific access,
so replace them with your own endpoints when needed. Bitcoin also requires a
locally running node matching the credentials in the configuration.

### Launch the services

Start the supporting infrastructure from the backend workspace:

```bash
cd ./backend
docker compose up -d
```

Run the backend API in a separate terminal:

```bash
cd ./backend
go run .
```

Launch the frontend client in another terminal:

```bash
cd ./frontend
npm install
npm run dev
```

This brings up the Go backend alongside the Vite-based frontend so both layers of the application can be used locally.

#### Public endpoints

- `GET /api/v1/coins` List available supported coins and tokens.
- `GET /api/v1/coins/:id`  Get metadata for a specific coin.
- `GET /api/v1/prices` Retrieve current price data for tracked assets.
- `GET /api/v1/prices/:id` Retrieve current price data for a specific asset.
- `GET /health` Check service health and uptime.
#### Bearer token required
- `GET /api/v1/wallets` List saved wallets.
- `POST /api/v1/wallets` Add a new wallet to the tracker.
- `PUT /api/v1/wallets` Update wallet details.
- `GET /api/v1/wallets/balance` Fetch aggregated wallet balances.
- `DELETE /api/v1/wallets` Remove a wallet from tracking.

### Current goals

The project is aimed at becoming a crypto monitoring platform for:

- wallet tracking
- whale movement monitoring
- price alerts
- future mobile app integration

### Database migration notes
```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate -path $PWD/internal/db/migrations -database "postgres://tracker_admin:super_secure_password@localhost:5432/whale_tracker?sslmode=disable" up
```