#!/usr/bin/env bash
set -euo pipefail

DB_PATH="${DB_PATH:-./data/book_social_dev.db}"

mkdir -p "$(dirname "$DB_PATH")"

rm -f "$DB_PATH"

sqlite3 "$DB_PATH" < db/sqlite/schema_v0_1_sqlite.sql
sqlite3 "$DB_PATH" < db/sqlite/seed_sqlite.sql

echo "Database reset: $DB_PATH"