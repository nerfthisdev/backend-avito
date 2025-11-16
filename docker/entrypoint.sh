#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${DB_DSN:-}" && -z "${GOOSE_DBSTRING:-}" ]]; then
  echo "DB_DSN (or GOOSE_DBSTRING) must be set for migrations" >&2
  exit 1
fi

export GOOSE_DRIVER="${GOOSE_DRIVER:-postgres}"
export GOOSE_DBSTRING="${GOOSE_DBSTRING:-$DB_DSN}"
export GOOSE_MIGRATION_DIR="${GOOSE_MIGRATION_DIR:-/app/migrations}"
export GOOSE_TABLE="${GOOSE_TABLE:-goose_migrations}"

echo "Running database migrations..."
attempt=1
until goose -dir "${GOOSE_MIGRATION_DIR}" up; do
  if (( attempt >= 10 )); then
    echo "Goose failed after ${attempt} attempts, aborting" >&2
    exit 1
  fi

  echo "Migrations attempt ${attempt} failed, retrying in 3s..."
  attempt=$((attempt + 1))
  sleep 3
done

echo "Migrations applied, starting service"
exec /usr/local/bin/pr-assigner
