# Project Context

Book Social is a learning Go web project.

Main goals:
- practice Go web development
- learn layered architecture
- build a modular monolith
- practice MPA / SSR with templates
- start with SQLite
- later compare SQLite and PostgreSQL
- gradually add tests, auth, better catalog behavior, HTMX, and Templ

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

Current catalog behavior:
- `/books` lists books.
- `/books?author={authorSlug}` filters by author slug.
- `/books?genre={genreSlug}` filters by genre slug.
- `/books?author={authorSlug}&genre={genreSlug}` applies both filters.
- `/books/{bookSlug}` shows book details.
- `/authors/{authorSlug}` shows an author page and books by that author.
