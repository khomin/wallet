# Whale Tracker

A Go-based backend for tracking crypto prices, wallet activity, and future alerting workflows. The project is still in its early stages, but the foundation is already in place for price collection, caching, persistence, and API exposure.

This repo is being built as a backend-first project for now, with plans to expand into a full app experience later.

## What it does

Right now, the service includes:

- A Go HTTP server with health and price endpoints
- Background price fetching for crypto assets
- Redis-based caching for recent values
- PostgreSQL persistence for snapshots and storage
- Docker Compose support for local infrastructure

## Current goals

The project is aimed at becoming a crypto monitoring platform for:

- wallet tracking
- whale movement monitoring
- price alerts
- future mobile/web app integration

## Tech stack

- Go
- Gin Web Framework
- PostgreSQL
- Redis
- Docker Compose
- Viper for configuration
- Logrus for logging

## Project structure

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

## Getting started

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

### Health check

```bash
curl http://localhost:8080/health
```

### Prices

```bash
curl http://localhost:8080/api/v1/prices
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
