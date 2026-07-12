#!/usr/bin/env sh
set -eu

app_env="${APP_ENV:-dev}"

if [ "$app_env" = "dev" ]; then
  db_dsn="${APP_DB_DSN:-./data/book_social_dev.db}"
  db_path="${db_dsn#file:}"

  mkdir -p "$(dirname "$db_path")"

  if [ ! -s "$db_path" ]; then
    sqlite3 "$db_path" < db/sqlite/schema_v0_1.sql
    sqlite3 "$db_path" < db/sqlite/seed.sql
    echo "Database initialized: $db_path"
  fi
fi

exec "$@"
