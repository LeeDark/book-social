# Database

Database docs are split by project stage:

- [Database v0.1](database_v0_1.md): active SQLite development schema.
- [Database v0.2 target](database_v0_2.md): planned normalized target schema.

Current state:

- `APP_ENV=dev` uses SQLite and is the active local development path.
- `APP_ENV=stage` and `APP_ENV=prod` open PostgreSQL using `APP_DB_DSN`.
- There is no migration system yet.
- PostgreSQL has a connection package and v0.1 book repository implementation.
- Docker/Compose uses a dev-only SQLite database in a named volume.

For local database reset:

```bash
make db/reset
```

For Docker database reset:

```bash
make docker/down
make docker/up
```

For a PostgreSQL database, apply the v0.1 SQL files manually for now:

```bash
psql "$APP_DB_DSN" -f db/postgresql/schema_v0_1.sql
psql "$APP_DB_DSN" -f db/postgresql/seed.sql
```
