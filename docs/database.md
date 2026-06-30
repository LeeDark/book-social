# Database

Database docs are split by project stage:

- [Database v0.1](database_v0_1.md): active SQLite development schema.
- [Database v0.2 target](database_v0_2.md): planned normalized target schema.

Current state:

- SQLite is active for local development.
- There is no migration system yet.
- PostgreSQL support is planned, not implemented.
- Docker/Compose database wiring is experimental.

For local database reset:

```bash
make db-dev-reset
```
