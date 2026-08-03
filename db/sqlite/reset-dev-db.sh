#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../.." && pwd)"

DB_PATH="${DB_PATH:-$REPO_ROOT/data/book_social_dev.db}"
MIGRATE="${MIGRATE:-migrate}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$SCRIPT_DIR/migrations}"
MIGRATIONS_DATABASE_URL="${MIGRATIONS_DATABASE_URL:-sqlite://$DB_PATH}"
SEED_PATH="$SCRIPT_DIR/seed.sql"

mkdir -p "$(dirname "$DB_PATH")"

rm -f "$DB_PATH" "$DB_PATH-wal" "$DB_PATH-shm" "$DB_PATH-journal"

"$MIGRATE" \
    -path "$MIGRATIONS_DIR" \
    -database "$MIGRATIONS_DATABASE_URL" \
    up
sqlite3 "$DB_PATH" < "$SEED_PATH"

echo "Database reset: $DB_PATH"
