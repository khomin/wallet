# Technical & Operational Notes

Useful CLI commands, manual operational workflows, and system integration details for development and debugging.

## Database Migrations

Database migrations automatically run on application startup inside `backend/main.go` using embedded migration files. 

If you need to run, roll back, or inspect migrations manually outside the application runtime, use the `golang-migrate` CLI tool:

### Manual Execution

```bash
# Install golang-migrate with postgres driver
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run pending migrations
migrate -path backend/internal/db/migrations \
  -database "postgres://tracker_admin:super_secure_password@localhost:5432/whale_tracker?sslmode=disable" up

# Rollback last migration
migrate -path backend/internal/db/migrations \
  -database "postgres://tracker_admin:super_secure_password@localhost:5432/whale_tracker?sslmode=disable" down 1
```

## Local Development Workflows

### Re-generating Protocol Buffers

Whenever `.proto` files are modified under `proto/`:

```bash
buf dep update
buf generate
```

This updates Go code under `backend/gen` and generated TypeScript definitions for the frontend.

### Testing Keycloak Integration

- Local Keycloak UI is accessible at `http://localhost:8080` when `compose.dev.yml` is running.
- Keycloak realm configuration is stored at `backend/deploy/keycloak/realm-export.json`.

### Blockchain RPC Debugging

- Ensure `TRON_GRID_API_KEY` is specified in `backend/.env` for Tron network balance queries.
- Solana RPC node endpoint can be overridden in `backend/config.yaml` to avoid public RPC rate limits.

## API Testing with Bruno

An API collection containing pre-configured requests for all HTTP REST gateway endpoints is located at:

`tools/bruno/`

Open this folder in [Bruno](https://www.usebruno.com/) to quickly test prices, wallets, users, and alert endpoints locally.