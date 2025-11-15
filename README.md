# backend-avito

[![lint](https://github.com/nerfthisdev/backend-avito/actions/workflows/ci.yml/badge.svg)](https://github.com/nerfthisdev/backend-avito/actions/workflows/ci.yml)

## Local Setup

### Requirements
- Go toolchain (build/test)
- Docker + Docker Compose (local Postgres)
- [goose](https://github.com/pressly/goose) CLI (migrations)
- [air](https://github.com/air-verse/air) (live reload)

### Environment
Create a `.env` in the repo root (or export the variables) before running `make` targets:

```env
DB_DSN=postgres://admin:admin@localhost:5432/avito_db?sslmode=disable
DB_DSN_TEST=postgres://admin:admin@localhost:5432/avito_db?sslmode=disable

# Goose
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=${DB_DSN}
GOOSE_MIGRATION_DIR=./migrations
GOOSE_TABLE=goose_migrations
```

## Workflows

### Database & migrations
```bash
make db-up           # start Postgres via docker compose
make db-down         # stop the container

make migrate-up      # apply all migrations
make migrate-down    # roll back one migration
make migrate-status  # show migration history/status
make migrate-new name=create_users_table  # scaffold a migration
```

### Development server
```bash
make dev  # runs air for live reload; make sure the DB is running
```

### Build & clean
```bash
make build  # outputs ./bin/pr-assigner
make clean  # removes ./bin and ./tmp artifacts
```

### Tests
```bash
make test  # runs go test ./...
```

### Command reference
```bash
make help  # prints the summary of available targets
```
