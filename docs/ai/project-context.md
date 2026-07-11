# Project Context

Book Social is a learning Go web project.

Main goals:
- practice Go web development
- learn layered architecture
- build a modular monolith
- practice MPA / SSR with templates
- start with SQLite
- compare SQLite and PostgreSQL incrementally
- gradually add tests, auth, better catalog behavior, and selected frontend experiments

Current modules:
- home
- books/catalog
- book details
- author details
- static assets
- rendering/templates
- app skeleton: config, logging, errors

Current focus:
- minimal working features first
- clean package boundaries
- small incremental tasks
- current documentation should describe implemented behavior, not planned behavior

Current catalog behavior:
- `/books` lists books.
- `/books?author={authorSlug}` filters by author slug.
- `/books?genre={genreSlug}` filters by genre slug.
- `/books?author={authorSlug}&genre={genreSlug}` applies both filters.
- `/books/{bookSlug}` shows book details.
- `/authors/{authorSlug}` shows an author page and books by that author.

Current rendering direction:
- `html/template` is the primary rendering path.
- HTMX is present as a small progressive-enhancement spike for catalog filters.
- Templ and gomponents routes are experiments documented in spike notes, not the main frontend direction.

Current infrastructure caveat:
- `APP_ENV=dev` uses SQLite and is the active local database path.
- `APP_ENV=stage` and `APP_ENV=prod` open PostgreSQL with `APP_DB_DSN`.
- PostgreSQL catalog repository methods implement the current v0.1 SQLite behavior.
- There is no migration system yet; PostgreSQL databases are initialized manually from SQL files.
- Docker/Compose are supported as a basic local development setup for v0.1.
- Docker/Compose are not production-ready infrastructure.
