# Project Context

Book Social is a learning Go web project.

Main goals:
- practice Go web development
- learn layered architecture
- build a modular monolith
- practice MPA / SSR with templates
- start with SQLite
- later compare SQLite and PostgreSQL
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
- SQLite is the active local database.
- Docker/Compose are supported as a basic local development setup for v0.1.
- Docker/Compose are not production-ready infrastructure.
