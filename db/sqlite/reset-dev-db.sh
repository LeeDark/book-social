#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "$SCRIPT_DIR/../.." && pwd)"

DB_PATH="${DB_PATH:-$REPO_ROOT/data/book_social_dev.db}"
SCHEMA_PATH="$SCRIPT_DIR/schema_v0_1.sql"
SEED_PATH="$SCRIPT_DIR/seed.sql"

mkdir -p "$(dirname "$DB_PATH")"

rm -f "$DB_PATH"

sqlite3 "$DB_PATH" < "$SCHEMA_PATH"
sqlite3 "$DB_PATH" < "$SEED_PATH"

echo "Database reset: $DB_PATH"
