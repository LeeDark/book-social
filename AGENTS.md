# AGENTS.md

## Project

This is a Go Book Social learning project.

Architecture:
- modular monolith
- layered architecture
- HTTP handlers
- services / use cases
- repositories
- MPA / server-side templates
- SQLite for now, PostgreSQL may be added later

## Current technical direction

- Keep the project simple and educational.
- Prefer clear Go code over clever abstractions.
- Avoid large refactoring unless the task explicitly asks for it.
- Keep package boundaries clean.
- Do not introduce heavy dependencies without a strong reason.

## Testing

- Use the standard Go testing package.
- Prefer table-driven tests.
- Use `httptest` for HTTP handlers.
- Use fake repositories/services for unit tests.
- Avoid database integration tests unless explicitly requested.
- `go test ./...` must pass before finishing.

## UI

- This is an MPA project.
- Do not introduce frontend frameworks.
- Keep templates simple.
- Do not over-test HTML markup.

## Before changing code

1. Inspect the existing structure.
2. Explain briefly what you plan to change.
3. Make the smallest reasonable change.
4. Run or explain the relevant tests.

## After changing code

Summarize:
- files changed
- tests added/updated
- commands run
- anything intentionally left for later