ifneq (,$(wildcard .env))
	include .env
	export
endif

DB_DSN ?= postgres://admin:admin@localhost:5432/avito_db?sslmode=disable
MIGRATIONS_DIR := ./migrations
GOOSE_TABLE    := goose_migrations

GOOSE_CMD := goose"

APP_NAME := pr-assigner
OUT_DIR := bin
OUT_BIN := $(OUT_DIR)/$(APP_NAME)

.PHONY: help
help:
	@echo "make db-up        - start postgres via docker-compose"
	@echo "make db-down      - stop postgres"
	@echo "make migrate-up   - run all migrations up"
	@echo "make migrate-down - rollback one migration"
	@echo "make migrate-status - show migrations status"
	@echo "make migrate-new name=xxx - create new migration"
	@echo "make dev          - run air (live reload)"
	@echo "make build	 - builds go binary"
	@echo "make clean	 - cleans build directory"
	@echo "make test	 - runs all tests"

.PHONY: db-up
db-up:
	docker compose up -d

.PHONY: db-down
db-down:
	docker compose down

.PHONY: migrate-up
migrate-up:
	goose up

.PHONY: migrate-down
migrate-down:
	goose down

.PHONY: migrate-status
migrate-status:
	goose status

.PHONY: migrate-new
migrate-new:
ifndef name
	$(error "Usage: make migrate-new name=create_users_table")
endif
	goose -dir $(MIGRATIONS_DIR) -s create $(name) sql

.PHONY: dev
dev:
	air

.PHONY: build
build:
	@mkdir -p $(OUT_DIR)
	go build -o $(OUT_BIN) ./cmd/$(APP_NAME)
	@echo "Built: $(OUT_BIN)"

.PHONY: clean
clean:
	rm -rf $(OUT_DIR) tmp

.PHONY: test
test:
	go test ./...

