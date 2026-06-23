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

## Testing

Use:

```bash
make test
```

## Running the web server in Codex

Do not try to start the web server inside the Codex sandbox for verification.

Avoid commands like:

```bash
GOCACHE=/tmp/book-social-go-cache APP_HTTP_ADDR=:18080 go run ./cmd/web
curl -I http://localhost:18080/books
```

The Codex sandbox may not allow opening listening sockets or accessing `localhost`, so these checks can fail with environment errors such as:

```text
listen tcp :18080: socket: operation not permitted
curl: (7) Couldn't connect to server
```

These errors should be treated as sandbox/environment limitations, not as project failures.

For automated verification, prefer:

```bash
GOCACHE=/tmp/book-social-go-cache go test ./...
```

For HTTP behavior, use Go tests with `net/http/httptest` instead of starting a real server.

For visual/manual checks, report the exact routes the user should open locally, for example:

```text
/
 /books
 /books/{valid-slug}
 /books/unknown-slug
```

If a task requires browser verification, stop and ask the user to run the app locally outside the Codex sandbox.
