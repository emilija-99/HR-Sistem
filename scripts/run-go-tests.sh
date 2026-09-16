#!/bin/sh
# Runs the Go test suite. Unit tests always run; DB-backed integration tests run
# when the database from the project .env is reachable.
#
# Usage: scripts/run-go-tests.sh [extra go test flags]
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT/backend"

if [ -f "$ROOT/.env" ]; then
  set -a
  # shellcheck disable=SC1091
  . "$ROOT/.env"
  set +a
fi

if [ -n "$POSTGRES_USER" ]; then
  export TEST_DATABASE_URL="host=${DB_HOST:-localhost} port=${DB_PORT:-5433} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB:-hr_db} sslmode=${DB_SSLMODE:-disable}"
fi

echo "== go test ./... =="
go test ./... "$@"
