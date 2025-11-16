# backend-avito

[![lint](https://github.com/nerfthisdev/backend-avito/actions/workflows/ci.yml/badge.svg)](https://github.com/nerfthisdev/backend-avito/actions/workflows/ci.yml)

[Load testing](#load-testing)

Как описано в условии все поднимается
```
docker compose up
```
но если вы хотите запустить отдеальные части задания все описано в [Makefile](./Makefile)

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

### Docker Compose (full service)

```bash
docker compose up --build
```

This builds the service image, runs Postgres, applies all migrations automatically, and exposes the HTTP API on `localhost:8080`. Stop everything with `docker compose down`.

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

Integration tests (require DB + `DB_DSN_TEST`):
```bash
go test -tags integration ./test/integration/...
```

### Command reference
```bash
make help  # prints the summary of available targets
```

## Load testing


Я тестировал с помощью k6, сценарий можете увидеть в папке loadtest -> [loadtest/scenario.js](./loadtest/scenario.js)

Tool: k6  
Scenario: 10 VUs, 1 minute, ~274 HTTP RPS (≈ 50x higher than target RPS=5).

Each iteration:
- POST /team/add (idempotent, may return 201 or 400 TEAM_EXISTS)
- POST /pullRequest/create (may return 201 or 409 PR_EXISTS)
- POST /pullRequest/merge (idempotent, returns 200)

Results:
- Total iterations: 5490 (~91/s)
- Total HTTP requests: 16470 (~274/s)

Latency (http_req_duration):
- avg:   2.84 ms
- p95:   7.53 ms
- max:  16.97 ms

Checks:
- 100% of functional checks passed:
- team.add is 201 or 400
- pr.create is 201 or 409
- pr.merge is 200

### Notes
 - k6 сообщает `http_req_failed ≈ 62%`, так как считает все ответы `4xx` ошибками.
 - В данном сценарии статусы `400 TEAM_EXISTS` и `409 PR_EXISTS` являются ожидаемыми доменными результатами, возникающими при многократных запросах с одинаковыми идентификаторами.
 - Checks 100%
### Conclusion
 - Даже при нагрузке **примерно в 50 раз выше требуемой**, сервис остаётся значительно быстрее SLI-порога в **300 мс**: показатель **p95 < 10 мс**.
 - Функциональное поведение полностью корректно (checks = 100%).
 - Сервис демонстрирует значительный запас производительности и готов к эксплуатации в условиях заявленной нагрузки.
