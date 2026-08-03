#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

DB_NAME="${PGDATABASE:-book_social}"
DB_USER="${PGUSER:-book_social}"
DB_PASSWORD="${PGPASSWORD:-pa55word}"
DB_HOST="${PGHOST:-localhost}"
DB_PORT="${PGPORT:-5432}"

MIGRATE="${MIGRATE:-migrate}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$SCRIPT_DIR/migrations}"
MIGRATIONS_DATABASE_URL="${MIGRATIONS_DATABASE_URL:-postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable&x-multi-statement=true}"
SEED_PATH="$SCRIPT_DIR/seed.sql"

export PGPASSWORD="$DB_PASSWORD"

PSQL=(
    psql
    -v ON_ERROR_STOP=1
    -h "$DB_HOST"
    -p "$DB_PORT"
    -U "$DB_USER"
    -d "$DB_NAME"
)

"${PSQL[@]}" <<'SQL'
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;
SQL

"$MIGRATE" \
    -path "$MIGRATIONS_DIR" \
    -database "$MIGRATIONS_DATABASE_URL" \
    up
"${PSQL[@]}" -f "$SEED_PATH"

echo "PostgreSQL database reset: $DB_NAME"
