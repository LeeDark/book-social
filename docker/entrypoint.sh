#!/usr/bin/env sh
set -eu

app_env="${APP_ENV:-dev}"

if [ "$app_env" = "dev" ]; then
  db_dsn="${APP_DB_DSN:-./data/book_social_dev.db}"
  db_path="${db_dsn#file:}"

  mkdir -p "$(dirname "$db_path")"

  if [ ! -s "$db_path" ]; then
    migrate \
      -path /app/db/sqlite/migrations \
      -database "sqlite://$db_path" \
      up
    sqlite3 "$db_path" < db/sqlite/seed.sql
    echo "Database initialized: $db_path"
  fi
elif [ "$app_env" = "stage" ] || [ "$app_env" = "prod" ]; then
  db_dsn="${APP_DB_DSN:?APP_DB_DSN is required for PostgreSQL environments}"
  migration_dsn="${MIGRATIONS_DATABASE_URL:-$db_dsn}"

  case "$migration_dsn" in
    *x-multi-statement=*) ;;
    *\?*) migration_dsn="${migration_dsn}&x-multi-statement=true" ;;
    *) migration_dsn="${migration_dsn}?x-multi-statement=true" ;;
  esac

  migrate \
    -path /app/db/postgresql/migrations \
    -database "$migration_dsn" \
    up

  book_count="$(psql "$db_dsn" -v ON_ERROR_STOP=1 -tA -c 'SELECT COUNT(*) FROM books')"
  if [ "$book_count" -eq 0 ]; then
    psql "$db_dsn" -v ON_ERROR_STOP=1 -f /app/db/postgresql/seed.sql
    echo "Database seeded: $db_dsn"
  fi
fi

exec "$@"
