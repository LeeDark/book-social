# Technical Backlog

This document collects technical ideas, experiments, cleanup opportunities, and possible future
tasks that are not committed release work.

`docs/roadmap.md` remains the source of truth for the current priority and accepted public scope
through v0.5. The full release sequence lives in `docs/private/roadmap.md` when available. An item
in this backlog is not a promise and should not interrupt the active roadmap stage. Review backlog
items when planning a release, after learning from an experiment, or when their revisit conditions
become true. Promote an item to the relevant roadmap only with a concrete user or operator outcome,
scope, verification plan, and definition of done.

## Candidate format

Each item should record:

- status and current decision;
- why it is not roadmap work now;
- evidence or historical notes;
- concrete conditions that justify reconsidering it.

Remove resolved items or keep their final decision in the relevant architecture/task-history
document. Do not use this file as an unbounded duplicate of the roadmap.

## Current planning boundary

v0.2.6 Registration/Login/Logout, v0.3.0 Private Library Foundation, and v0.3.1 Reading-State
Lifecycle are closed. No later release scope is active yet. No backlog item is promoted by this
document; defer unrelated API, rendering, and infrastructure ideas until a concrete revisit
condition is met.

## Code hygiene linked to releases

Status: scheduled cleanup from the 2026-11-04 review of the full code history at `85c7899`.

Current decision: each item below is attached to the release whose roadmap work touches the same
code, or to that release's closure. Do it in the same change as the named roadmap item; it does
not add a separate roadmap outcome.

Why it is not roadmap work: none of these changes is visible to a user or operator on its own,
which the roadmap's planning principles require of release scope.

Revisit when planning each release. If the linked roadmap item moves to another release, move its
hygiene item with it. Remove an item once it is done, and record the reason for a dropped item in
`docs/ai/task-history.md`.

### v0.3.3: with release closure

Roadmap:
[v0.3.3 — Reliability and Release Closure](roadmap.md#v033--reliability-and-release-closure).

- Remove `github.com/a-h/templ` and `maragu.dev/gomponents` from the `depguard` allowlist
  (`.golangci.yml:43`, `:48`). Neither module has been in `go.mod` since `5c7edc1`, and the
  [Templ or gomponents](#typed-server-side-rendering-templ-or-gomponents) item requires a decision
  before either returns.
- Re-panic `http.ErrAbortHandler` in the project's `Recoverer`, as chi's own recoverer does
  (`internal/http/middleware/recovery.go:18`).
- Call `stop()` once the root context is done, so that a second interrupt ends the process during
  the five-second shutdown (`cmd/web/main.go:31`).
- With the library ID-range item: resolve a book ID from its slug without loading authors, genres
  and covers. `library.Service.Add` uses the catalog's full `GetBookBySlug` today
  (`internal/modules/library/service.go:39`).
- With the login/email item: remove the unused `users.Service.Register`, which only tests call
  (`internal/modules/users/service.go:91`), and require `RegistrationSessionStore` in the
  `NewService` parameter type instead of checking it with a type assertion at run time
  (`internal/modules/users/service.go:23`).

### v0.3.4: with the catalog expansion

Roadmap: [v0.3.4 — Public Demo Polish](roadmap.md#v034--public-demo-polish).

- With the cover-bundle limit item: define the cover file limits (5 MiB, 4096 px per side,
  12,000,000 pixels, portrait JPEG, PNG or static WebP) once in Go and use that definition in both
  `internal/storage/covers/image.go` and `internal/modules/books/local_cover.go:57-59`. Migration
  `000008` keeps the database-side `CHECK` mirror for both dialects.
- With the featured-books item: share one batch helper per storage package that loads authors and
  genres for a list of book IDs. Book and library repositories hold four copies today
  (`internal/storage/sqlite/books_repository.go:168`, `:236`, `library_repository.go:363`, `:429`;
  `internal/storage/postgresql/books_repository.go:139`, `:182`, `library_repository.go:313`,
  `:378`). If v0.3.4 does not change the card read model, move this item to
  [v0.4.0](roadmap.md#v040--catalog-and-locale-contracts).

### v0.3.5: with stage hardening

Roadmap: [v0.3.5 — Hetzner Cloud Hosted Stage](roadmap.md#v035--hetzner-cloud-hosted-stage).

- With the missing-book request limit item: replace the `users.ErrUnauthenticated` import in
  `books` with a local error, so that the catalog module no longer depends on `users`
  (`internal/modules/books/missing_book.go:10`, `missing_book_handler.go:14`).

### Not attached to a release

- Lint test files. `.golangci.yml:6` sets `run.tests: false`, so about 14,000 lines of tests are
  not linted. Run the linter once with `tests: true` to measure the findings. Attach this item to
  a release when the linter set grows, or when a test defect appears that a linter would have
  caught.

## Public catalog JSON API

Status: deferred; no active implementation.

Current decision: the catalog remains MPA-first. A public, read-only `/api/v1/books` slice is not
roadmap work until a concrete consumer exists.

Revisit only when a named same-origin interactive feature, partner integration, or other external
consumer needs machine-readable public catalog data. Define the exact read contract, pagination,
error model, compatibility policy, tests, and OpenAPI scope then. Do not add CORS, client tokens,
or rate limiting in anticipation of that use case.

## Typed server-side rendering: Templ or gomponents

Status: deferred; no active implementation.

Current decision: keep `html/template` as the only application rendering path. The completed Templ
and gomponents spikes were removed from executable code during v0.2.3 cleanup so that catalog
changes do not require maintaining three parallel `BookCard` models and routes.

What the experiments showed:

- Templ gives typed component contracts and HTML-like source, but adds code generation, generated
  files, tooling, and adapter types.
- gomponents uses normal Go code and needs no generator, but larger HTML structures are less easy
  to scan and still require adapter types at the package boundary.
- `html/template` remains the simplest fit for the current MPA, where reusable UI is still small
  and existing partials are adequate.

Historical evidence:

- [Templ BookCard spike](ai/templ-spike-book-card.md)
- [Templ and gomponents comparison](ai/frontend-rendering-spike-book-card.md)

Revisit only if at least one of these becomes true:

- several non-trivial components are repeated across pages and template partial contracts become
  difficult to maintain;
- template-field mistakes or refactors cause recurring defects that typed components would catch;
- a planned UI change has enough scope to measure one typed approach against `html/template`.

If revisited, select one concrete component and one candidate technology. Define the expected
benefit and removal criteria before adding a dependency; do not restore both experimental stacks
or public spike routes by default.
