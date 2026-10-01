# Roadmap

This roadmap is a working guide, not a release promise.
This public version covers releases through v0.5.

Unprioritized technical ideas and revisit conditions live in [backlog.md](backlog.md). The backlog
does not change the active priority or release scope until an item is explicitly promoted here.

## Current Planning Status

v0.2.1 through v0.2.6 are closed. v0.2.6 exposes the accepted auth foundation through the first
browser-facing flow: registration, login, idempotent logout, protected `/me`, auth navigation, and
one-request success flashes. Its manual browser smoke and release closure are complete.

v0.3.0 Private Library Foundation and v0.3.1 Reading-State Lifecycle are closed. The v0.3.1
release completed automated verification, SQLite migration smoke, and the mandatory manual browser
smoke. No later release scope is active yet.

## Planning Principles

- Each release should deliver an observable user or operator outcome, not only introduce
  technologies.
- Product value determines the release boundary; technical work supports that value.
- The modular monolith and MPA remain the default until measured requirements justify a different
  architecture.
- Accessibility, security, testing, observability, privacy, and operational readiness grow with
  every release instead of being postponed to one final hardening phase.
- Items beyond the current release are direction markers. They are refined using implementation
  results, product metrics, and user interviews.

## v0.1 Baseline

v0.1 is closed as the baseline for the next development stage.

Done:

- [x] Basic app skeleton: config, logging, errors, build info, Makefile.
- [x] chi router, middleware, routes, 404 rendering.
- [x] Modular monolith shape.
- [x] Layered books/catalog module: handler, service, repository.
- [x] SQLite schema, seed data, reset script.
- [x] Home and About pages.
- [x] Catalog page.
- [x] Book details page.
- [x] Author page.
- [x] Author and genre catalog filters.
- [x] View/page models, navigation, breadcrumbs.
- [x] Basic unit tests and small HTTP/integration-style tests.
- [x] HTMX catalog filter spike.
- [x] Templ and gomponents rendering spikes.
- [x] Initial documentation audit.
- [x] Close v0.1 as the baseline.

Still intentionally rough:

- [x] Docker/Compose local environment workflows.
- [x] No database migrations in the v0.1 runtime/reset workflow.
- [x] PostgreSQL has startup support and v0.1 catalog repository behavior.
- [x] No auth or user library features.
- [x] No search, sorting, or pagination.
- [x] No real cover image storage.

## v0.2 — Catalog and Authentication Foundation

Goal: create a safely changeable catalog and the minimal authenticated-user foundation needed by
the first personal product workflow.

v0.2 should be implemented by dependency order, not as one large mixed feature batch.

Main waves:

- Data / infrastructure: quality checks, migrations, SQLite/PostgreSQL decision, schema v0.2,
  repositories, catalog changes.
- User / auth: sessions, cookies, registration, login/logout, auth middleware, user-facing auth UI.

Release outcome:

- The normalized catalog works through the existing MPA routes.
- A user can register, log in, log out, and open a protected page.
- The database can be migrated and tested without relying on destructive manual schema resets.
- The final personal-library schema is deliberately deferred until its v0.3 use cases and state
  transitions are defined.

## v0.2.1 Quality & DB Foundation

Goal: make schema changes safe before changing the catalog model.

- [x] Add or confirm Make targets: `test`, `fmt`, `vet`, optional `lint`.
- [x] Add CI for `go test ./...`, `go vet ./...`, and lint if configured.
- [x] Decide how the database driver / config is selected.
- [x] Define migration layout for SQLite and possible PostgreSQL support.
- [x] Add migration commands using the `golang-migrate` CLI.
- [x] Clarify `reset-db` and seed workflow.
- [x] Add minimal test DB bootstrap helpers if needed for repository tests.

Definition of Done:

- [x] `make test` passes.
- [x] The project has a clear migration path for v0.2 schema work.
- [x] PostgreSQL is minimally supported at startup and has v0.1 catalog repository behavior.

Related docs:

- `README.md`
- `docs/development.md`
- `docs/database.md`
- `docs/database_v0_1.md`
- `docs/testing.md`
- `docs/ai/project-context.md`

## v0.2.2 Catalog Domain Model

Goal: move the catalog from the v0.1 single-author/single-genre shape to the normalized v0.2 shape
without prematurely fixing the future personal-library model.

Preparation:

- [x] Add a SQLite migration and seed smoke target for a disposable database.
- [x] Keep SQLite migration smoke local/manual for now; defer CI integration.
- [x] Defer a PostgreSQL service job to a separate follow-up.
- [x] Redesign disposable reset flow to run migrations up, then apply seed data.
- [x] Keep seed data as development/sample data, separate from schema migrations.
- [x] Switch Docker/Compose bootstrap to migrations plus seed.
- [x] Plan and implement the v0.2 migration sequence before repository changes.

Schema:

- [x] Add migrations for `book_authors` and `book_genres`.
- [x] Add `covers` with URL and metadata fields, including `UNIQUE(book_id, variant)`.
- [x] Migrate v0.1 data from `books.book_author_id` and `books.book_genre_id`.
- [x] Remove old `books.book_author_id` and `books.book_genre_id` after data is migrated.
- [x] Preserve unused v0.1 `library`, `shelves`, and `tags` structures as explicit legacy/demo data.
- [ ] Defer the final `library_items` schema to v0.3, where `user_id`, reading status, dates,
  uniqueness, and privacy rules are defined together.
- [ ] Defer user shelves and `library_item_tags` beyond the initial private-library scope.
- [x] Update seed data for v0.2.
- [x] Decide and document slug policy for books, authors, and genres.

Definition of Done:

- [x] Fresh database setup works with v0.2 seed data.
- [x] Existing v0.1 catalog data has a defined migration story.
- [x] Legacy sample library structures cannot accidentally be mistaken for the final user-library
  model.
- [x] SQLite migration up/down smoke passes.
- [x] Seed smoke passes after migrations.
- [x] SQLite test DB can be created from the v0.2 schema.
- [x] Covers are stored as metadata/URLs only; upload/storage is deferred.

## v0.2.3 Catalog v0.2

Goal: restore and improve catalog reads after the schema change.

Status: closed. The next release is [v0.2.4 HTTP Foundation](#v024-http-foundation).

- [x] Remove executable Templ/gomponents spikes after recording their findings; keep
  `html/template` as the application rendering path and defer reconsideration to the backlog.
- [x] Add read models for book cards with multiple authors and genres.
- [x] Add a book details read model with authors, genres, and covers.
- [x] Add author details read model with the author's books.
- [x] Update catalog page.
- [x] Update book details page.
- [x] Update author details page.
- [x] Update the home page to use the new read model.
- [x] Update MPA endpoint documentation in `docs/routes.md` or a focused endpoint doc.

Definition of Done:

- [x] Stable routes still work: `/`, `/books`, `/books/{slug}`, `/authors/{slug}`.
- [x] Repository and handler tests cover the new catalog read shape, including opt-in PostgreSQL
  parity against a disposable DSN.
- [x] Templates still use view/page models rather than raw database structs.

## v0.2.4 HTTP Foundation

Goal: prepare the HTTP layer for auth and user workflows.

Status: closed. Its dependent [v0.2.5 Auth Foundation](#v025-auth-foundation) is also closed; the
current next release is [v0.2.6 Registration/Login/Logout](#v026-registrationloginlogout).

- [x] Add or review a graceful shutdown.
- [x] Confirm a base middleware chain.
- [x] Add request ID if useful.
- [x] Confirm logging middleware behavior.
- [x] Add panic recovery/error page behavior if missing.
- [x] Add secure headers' policy.
- [x] Add a static / cache headers policy.
- [x] Decide timeout middleware scope.

Definition of Done:

- [x] HTTP middleware order is documented or obvious in code.
- [x] Handler tests cover important error/status behavior.
- [x] No real server startup is required for Codex verification; use `httptest`.

## v0.2.5 Auth Foundation

Goal: create the minimal auth/session core before building forms.

Status: closed at accepted implementation commit `41a8ddb`. The next release is
[v0.2.6 Registration/Login/Logout](#v026-registrationloginlogout). v0.2.5 does not add auth pages,
production `/me`, navigation, or flashes.

- [x] Decide session strategy; use minimal DB-backed opaque sessions for this project.
- [x] Add equivalent SQLite/PostgreSQL `sessions` storage with hashed-token lookup and expiry.
- [x] Define password hashing policy: bcrypt hash only, no plaintext password storage or logging.
- [x] Add user repository/service foundations.
- [x] Add session create/delete/load behavior with a seven-day absolute lifetime.
- [x] Add typed current-user middleware and a testable route-guard foundation.
- [x] Define minimal transaction rule: services own transactions for use cases touching multiple
  tables.
- [x] Decide and install the browser CSRF strategy: global `http.CrossOriginProtection`, no form
  token, and no insecure bypass patterns.
- [x] Define a minimal validation strategy for auth inputs, including bcrypt's 72-byte boundary.

Definition of Done:

- [x] Auth core can create users, verify passwords, create sessions, delete sessions, and load the
  current user.
- [x] Unit, repository, middleware, and migration tests cover success and refusal paths.
- [x] Password, session, cookie, request-identity, and cross-origin behavior are documented enough
  for future maintenance.

## v0.2.6 Registration/Login/Logout

Goal: expose the minimal user-facing auth workflow.

Status: closed. Registration, login, idempotent POST logout, protected `/me`, auth navigation, and
one-request success flashes are complete. Manual browser smoke and release closure are complete.

- [x] Add a registration form and handler.
- [x] Add login form and handler.
- [x] Add a logout handler.
- [x] Add `/me` or another minimal protected route.
- [x] Update navigation for anonymous and logged-in states.
- [x] Add flash messages for login/logout/register outcomes.
- [x] Add validation errors for duplicate login/email and invalid credentials.
- [x] Add handler tests with `httptest`.

Definition of Done:

- [x] Anonymous users can register and log in.
- [x] Logged-in users can log out.
- [x] Anonymous users are redirected from the protected route.
- [x] `make test` passes.
- [x] v0.2 release notes or task history are updated.

## v0.3 — Reliable User Core

Goal: turn the authenticated catalog into the first complete user product.

Core user flow:

```text
register or log in
  -> select a book from the curated demo catalog
  -> add it to the private library
  -> choose a reading status
  -> return to the library later
```

### Sub-releases

The v0.3 scope is delivered in this dependency order. A sub-release is closed only after its own
acceptance criteria and relevant tests pass; unfinished work is not silently moved to the next one.

#### v0.3.0 — Private Library Foundation

Status: closed. The implemented behavior and its remaining lifecycle boundary are recorded in
[`library_v0_3.md`](library_v0_3.md).

- [x] Confirm the closed v0.2.6 auth baseline and write the library use cases, permission rules,
  status-transition rules, application errors, and transaction boundaries.
- [x] Add the `library` module and the minimal `library_items` migration, including unique
  `user_id + book_id`, ownership, and `added_at`.
- [x] Let an authenticated user add a known catalog book and view `/me/library`.
- [x] Cover migrations, repository behavior, anonymous access, duplicate addition, unknown books,
  and user isolation on SQLite; retain an explicit opt-in PostgreSQL verification path.

Outcome: a user can build a private want-to-read list from the existing catalog.

#### v0.3.1 — Reading-State Lifecycle

Status: closed. The bounded slice delivers `want_to_read`, `reading`, and `read` transitions with
documented `started_at` and `finished_at` rules.

- [x] Add `want_to_read`, `reading`, and `read` transitions, with documented `started_at` and
  `finished_at` rules.
- [x] Show a current user's library state on book pages; support conflict-safe status changes and
  explicit-confirmation removal.
- [x] Complete MPA form validation, keyboard-accessible errors, empty states, and focused unit and
  `httptest` coverage for lifecycle and permission refusals.

Outcome: the private library supports the complete add → update → return → remove cycle.

#### v0.3.2 — Curated Catalog Operations

- [ ] Separate test fixtures, deterministic development seed data, and the curated demo catalog;
  document demo metadata provenance and permitted cover use.
- [ ] Add a clear missing-book state and simple feedback path without an external import.
- [ ] Add only the administrator role, protected routes, catalog find/create/correct operations,
  and critical-change audit records needed to maintain the demo catalog.

Outcome: a small closed cohort can use reproducible catalog data that an operator can correct
without direct SQL.

#### v0.3.3 — Reliability and Release Closure

- [ ] Add the small critical-flow end-to-end HTTP set and complete SQLite migration/seed smoke;
  run the documented PostgreSQL critical-path verification.
- [ ] Record safe structured library-operation logs and product events, retaining request IDs and
  the existing health endpoint without adding a monitoring stack.
- [ ] Finish accessibility smoke checks, release checklist, and documentation for implemented
  routes, domain rules, migrations, demo data, and verification evidence.

Outcome: the complete private-library flow is releasable and operable as a curated demonstration.

#### v0.3.4 — Public Demo Polish

- [ ] Polish the shared MPA interface: typography, layout, responsive behavior, form states, empty
  states, contrast, and keyboard navigation.
- [ ] Expand the home and About pages with the learning project's purpose, implemented
  capabilities, technical approach, and honest curated-demo-catalog limitations.
- [ ] Add one shared footer with internal pages and only confirmed external links: source code,
  documentation, roadmap, and an available contact method.
- [ ] Add baseline public-page metadata; canonical URLs, redirects, and sitemap behavior remain
  v0.4 work.

Outcome: Book Social can be shown as a clear, polished public demonstration without claiming a
public beta or a finished social service.

#### v0.3.5 — Hetzner Cloud Hosted Stage

- [ ] Record an ADR: SQLite remains the local development, test, and disposable-preview path;
  hosted stage and future production use PostgreSQL. Render and other managed PaaS options remain
  evaluated alternatives, not this release's selected path.
- [ ] Deploy one reproducible Hetzner Cloud stage: Caddy with HTTPS, one application container,
  and one PostgreSQL container on a non-public internal connection; expose only required HTTP/HTTPS
  ports and limit SSH to trusted sources.
- [ ] Keep secrets outside the repository; apply the production cookie policy, trusted-proxy
  configuration, migrations, and health/readiness checks.
- [ ] Verify the critical browser flow after deployment; document manual application deployment and
  rollback without automating production delivery.
- [ ] Establish backups: server backups as a secondary layer, a daily PostgreSQL backup to storage
  separate from the VM, and a verified restore using disposable data.
- [ ] Record actual cost, CPU/RAM/latency, and growth boundaries; defer Prometheus and Grafana to
  v0.4, and add Loki or OpenTelemetry only when operational evidence justifies them.

Outcome: a low-cost, reproducible PostgreSQL stage environment runs on Hetzner Cloud with verified
deployment, migration, backup, and restore behavior; it can accommodate later observability
components without changing the application architecture.

### Product Scope

- [ ] Introduce a focused `library` module with handler, service/use-case, and repository
  boundaries.
- [ ] Define the minimal `library_items` schema around real use cases.
- [ ] Add a book to the current user's library from catalog and book pages.
- [ ] Support the system statuses `want_to_read`, `reading`, and `read`.
- [ ] Change the status of an existing item.
- [ ] Remove an item with explicit confirmation.
- [ ] Add `/me/library` and show the current library state on book pages.
- [ ] Enforce one item per `user_id + book_id`.
- [ ] Define `added_at`, `started_at`, and `finished_at` transition rules.
- [ ] Handle anonymous access, duplicate additions, unknown books, and update conflicts.
- [ ] Add a clear missing-book state and a simple way to request or report a missing book.

### Catalog and Operator Scope

- [ ] Separate test fixtures, deterministic development seed data, and a curated demo catalog.
- [ ] Prepare a reproducible demo catalog suitable for product demonstrations and a small closed
  user cohort.
- [ ] Document data provenance and permitted cover usage for demo records.
- [ ] Protect minimal admin routes with an administrator role.
- [ ] Allow an operator to find, create, and correct the metadata needed by the demo catalog.
- [ ] Record critical administrative changes with actor, entity, action, and time.

### Engineering Scope

- [ ] Define a small application error model for not found, validation, conflict, unauthorized,
  forbidden, and internal failures.
- [ ] Keep HTTP status mapping in handlers rather than services.
- [ ] Define repository contracts from the consuming service packages.
- [ ] Let services own transaction boundaries for multi-table use cases.
- [ ] Verify equivalent application semantics for SQLite and PostgreSQL.
- [ ] Cover status transitions and permissions with unit tests.
- [ ] Cover forms, validation, anonymous access, and conflicts with `httptest`.
- [ ] Add a small end-to-end HTTP test set for the critical library flow.
- [ ] Keep new forms keyboard-accessible with labels, useful errors, and actionable empty states.

### Observability Baseline

- [ ] Keep structured logs, request IDs, and a simple health endpoint.
- [ ] Record operation name, duration, and result type for critical library operations without
  passwords, tokens, private notes, or other sensitive data.
- [ ] Define product events for registration, first book addition, status change, and missing-book
  feedback.
- [ ] Keep the monitoring infrastructure optional in this release; prepare instrumentation for
  the next stage.

### Definition of Done

- [ ] A new user can complete the full register-to-library flow without manual database changes.
- [ ] Library data is private by default and cannot be changed by another user.
- [ ] Demo data can be reproduced separately from test fixtures.
- [ ] Critical flows pass on SQLite and have an explicit PostgreSQL verification path.
- [ ] An operator can correct the catalog data required for the demonstration without direct SQL.
- [ ] `make test` passes and a short release checklist covers migrations, demo data, and smoke
  checks.

Not in v0.3: external catalog import, ratings, notes, reading progress, custom shelves/tags, social
features, a public beta, production SLOs, automated production deployment, a full monitoring stack,
a frontend framework, message brokers, gRPC, or microservices.

## v0.4 — Bilingual Public Alpha and Live Catalog

Goal: let users find real books beyond the curated demo catalog and let an operator safely import
one missing book from a controlled external source. Open registration on a separate public
environment with complete English and Ukrainian support for the available user journey.

### Sub-releases

#### v0.4.0 — Catalog and Locale Contracts

- [ ] Decide work versus edition, identifiers, languages, provenance, duplicate/merge rules, and
  the migration path for existing catalog references.
- [ ] Add the minimum schema, repository contracts, and operator error/review outcomes required by
  those decisions.
- [ ] Fix the `en`/`uk` locale, fallback, public-URL, and existing-account migration contracts
  before
  changing routes or registration.

Outcome: catalog identity and language behavior are stable before search, import, or localization.

#### v0.4.1 — Bilingual MPA Foundation

- [ ] Move user-facing messages out of templates into stable keys and English/Ukrainian catalogs;
  translate the existing auth, catalog, library, error, and feedback flows.
- [ ] Add explicit language selection, persisted preference, locale resolution and fallback, and
  localized dates, numbers, and plurals; detect missing keys in tests.
- [ ] Add language-specific public GET URLs, equivalent-page switching, canonical/hreflang, and
  localized sitemap behavior; keep existing public GET URLs through redirects.

Outcome: the existing user journey works end to end in either language before new public features.

#### v0.4.2 — Search and Navigation

- [ ] Search by title, author, and ISBN; add stable sorting, pagination, and retained useful
  filters.
- [ ] Add bilingual zero-result feedback and conversion events; preserve language and search state
  across navigation and canonical redirects.
- [ ] Measure indexed-query behavior before selecting PostgreSQL full-text search.

Outcome: a user reliably finds an existing catalog record and can add it to their private library.

#### v0.4.3 — First Controlled Import

- [ ] Implement one source-specific lookup/import by ISBN or external ID, started only by an
  authenticated operator.
- [ ] Add source provenance, rate/timeout controls, idempotency, conflict review, and explicit
  metadata/cover-rights handling.
- [ ] Add a small PostgreSQL job table only if retry or asynchronous work is actually required.

Outcome: an operator can safely resolve one missing book without making an external API part of
ordinary catalog reads.

#### v0.4.4 — Public Account and Environment Readiness

- [ ] Add verified-email activation, expiring single-use password recovery, and rate limits for
  registration and login; translate their pages, errors, and transactional emails into `en`/`uk`.
- [ ] Publish privacy/terms and a support contact in both languages; document manual user-data
  export and deletion requests until self-service workflows are available.
- [ ] Provision a persistent public PostgreSQL environment separate from the v0.3.5 stage, with
  HTTPS, secrets, safe migrations, versioned deployment, and a documented rollback path.

Outcome: open registration and persistent user data have a supportable public operating path.

#### v0.4.5 — Import and Alpha Release Closure

- [ ] Add focused migration, search, import, conflict, and failure tests; document source rules and
  manual correction paths.
- [ ] Expose HTTP, PostgreSQL, search, import, retry, and failed-job metrics with focused Grafana
  dashboards; add tracing or centralized logs only when the stated conditions are met.
- [ ] Verify bilingual critical flows, account security, backup/restore, rollback, and the public
  browser path before opening registration.

Outcome: a bilingual public alpha is available, and its first import path is ready for v0.5 supply.

### Catalog Model

- [ ] Make an ADR for the distinction between a work and an edition before external import.
- [ ] Add ISBN and source-specific external identifiers where appropriate.
- [ ] Record source, provenance, import time, and raw-source reference for imported metadata.
- [ ] Define duplicate detection, merge, correction, and canonical-record rules.
- [ ] Make edition/content language explicit and define localized metadata fallback behavior.

### Discovery

- [ ] Search by title, author, and ISBN.
- [ ] Add stable sorting and pagination.
- [ ] Preserve useful author and genre filters.
- [ ] Measure queries before deciding whether PostgreSQL full-text search is required.
- [ ] Define canonical URLs, redirects after merges/slug changes, and sitemap behavior.
- [ ] Track zero-result searches and book-to-library conversion.

### Controlled Import

- [ ] Implement one source-specific importer for one book by ISBN or external ID.
- [ ] Trigger imports through an authenticated operator workflow, not a public synchronous request.
- [ ] Make repeated imports idempotent and expose conflicts for review.
- [ ] Respect source terms, rate limits, attribution, and metadata/cover usage rules.
- [ ] Use a small PostgreSQL-backed job table only if the import flow needs retries or background
  processing.

### Localization and Account Safety

- [ ] Support `en` and `uk` across the complete available user journey, including forms, validation,
  errors, support/legal text, and verification/recovery email; keep the operator UI single-language
  until v0.5.
- [ ] Keep edition/content language separate from UI locale; do not require every book to have
  translated metadata or translate user content automatically.
- [ ] Resolve explicit URL/selection before saved preference, anonymous cookie, browser hint, and
  English fallback; do not overwrite an explicit choice with `Accept-Language`.
- [ ] Use `/en/...` and `/uk/...` for indexable public pages; preserve existing private routes and
  the public GET links through redirects.
- [ ] Verify email before normal account use, provide secure recovery, limit auth abuse, and define
  a support-owned manual export/deletion process for alpha users.

### Monitoring

- [ ] Expose Prometheus-compatible RED metrics for critical HTTP paths.
- [ ] Add metrics for PostgreSQL pools/queries, search outcomes, imports, retries, and failed jobs.
- [ ] Create the first focused Grafana dashboards.
- [ ] Add an OpenTelemetry trace only for a genuinely multi-step import path if logs and metrics
  are insufficient.
- [ ] Add Loki only when a persistent staging environment or multiple processes make centralized
  log search useful.

### Public Alpha Operations

- [ ] Keep the v0.3.5 stage for testing and operate the public alpha in a separate persistent
  PostgreSQL environment; do not use the local Compose `prod` workflow as deployment guidance.
- [ ] Build and deploy a versioned artifact, run safe migrations, and document manual deployment,
  rollback, support escalation, and incident response.
- [ ] Store backups away from the public host and complete a restore rehearsal before launch; add
  basic availability/error alerts without requiring the full monitoring stack.

### Definition of Done

- [ ] A user can reliably find catalog records by title, author, or ISBN.
- [ ] An operator can import one missing book from the selected source and safely review conflicts.
- [ ] Repeated import does not create logical duplicates.
- [ ] Work/edition, language, provenance, canonical URL, and cover-source rules are documented.
- [ ] Dashboards answer concrete availability, latency, search, and import questions.
- [ ] Open registration and the available user journey work in `en` and `uk`, including account
  verification/recovery, transactional email, errors, legal text, and support.
- [ ] Public language URLs, redirects, `lang`, canonical/hreflang, and sitemap are checked.
- [ ] The separate public environment has tested migration, rollback, backup/restore, and a manual
  data-request process; the alpha is externally available.

Not in v0.4: a universal importer framework, public bulk imports, an external search engine,
runtime dependency on an external API for ordinary catalog requests, owned media storage, or
catalog microservices. Operator UI localization, self-service account deletion, social features,
and broad acquisition remain later work.

## v0.5 — Catalog Supply and Media Foundation

Goal: scale the catalog beyond a single source-specific import while keeping metadata provenance,
rights, operator control, and media handling explicit.

The public bulk import in this release means an operator-started import of an openly available or
licensed catalog dataset with visible progress and results. It is not an anonymous endpoint that
fetches arbitrary URLs, and it is not the later personal-library CSV import.

### Sub-releases

#### v0.5.0 — Source and Asset Contracts

- [ ] Define the source registry, shared adapter contract, normalized import record, source-health
  policy, and field-level provenance/license model.
- [ ] Add the asset model and storage boundary, including rights/attribution records, content
  validation, hash-based deduplication, local development storage, and a production storage adapter.

Outcome: source and media decisions are enforceable data contracts rather than conventions in an
import script.

#### v0.5.1 — Managed Batch Pilot

- [ ] Build the shared job lifecycle: dry run, validation report, operator review, idempotent apply,
  pause/resume, bounded retry, and audit.
- [ ] Run one bounded, isolated bulk pilot using an accepted source/dataset and measure quality,
  duplicates, identifiers, languages, rejected rows, and review rate.

Outcome: the team can prove a batch-import workflow without modifying the public catalog.

#### v0.5.2 — Reviewed Publication and Licensed Covers

- [ ] Promote accepted staging records only through an explicit operator decision; provide a
  read-only public progress/result page with source, version, scope, and limitations.
- [ ] Add failure isolation, back-pressure, and resumability so repeated or interrupted batches keep
  catalog consistency.
- [ ] Store and serve one class of explicitly permitted cover assets with recorded usage rights and
  attribution; reject invalid or unlicensed assets.
- [ ] Localize the operator catalog/import/review UI, including the v0.4 single-book import and new
  batch workflows, in English and Ukrainian.

Outcome: a reviewed batch can safely enrich the public catalog and use rights-cleared covers.

#### v0.5.3 — Supply-Chain Release Closure

- [ ] Test source admission, batch failure/retry/resume, corrections, media validation, and public
  result visibility; perform a security review for raw files and external fetches.
- [ ] Document source terms, quality/license evidence, retention, operator procedures, and metrics
  for the accepted source.
- [ ] Automate deployment to the separate stage environment and verify a versioned release there
  before promoting it to the public alpha.

Outcome: the first scalable catalog-supply and media path is operable without claiming a general
purpose importer or media platform.

### Import Platform

- [ ] Define a source registry, source-specific adapter contract, normalized import record, field
  mapping, provenance, license/usage metadata, and source-health policy.
- [ ] Build the shared job lifecycle: dry run, validation report, review, idempotent apply,
  pause/resume, retry policy, audit record, and read-only public progress/result page.
- [ ] Implement one production-like bulk adapter only after source terms, territorial rights,
  update rules, and data-quality checks are accepted; evaluate an Open Library dump as the initial
  candidate rather than treating its low-volume API as a catalog backend.
- [ ] Keep Google Books or equivalent services limited to attributed, action-driven lookup or
  enrichment unless their terms expressly permit durable bulk catalog storage.
- [ ] Make unresolved work/edition, duplicate, language, identifier, and metadata conflicts visible
  for operator review; never silently overwrite a manually corrected record.

### Public Dataset Import

- [ ] Run a bounded pilot import in an isolated environment and measure ISBN coverage, duplicates,
  language coverage, rejected records, and operator-review rate before publishing a larger run.
- [ ] Publish the dataset/source, scope, start/end time, version, aggregate result, and known data
  limitations for each accepted batch; retain raw-source references only as permitted.
- [ ] Add resource limits, resumability, back-pressure, failure isolation, and idempotency so a
  repeated or interrupted batch does not corrupt the catalog.
- [ ] Preserve the existing catalog during a failed import; promotion from staging data to the
  public
  catalog requires an explicit operator decision.

### Owned Media Foundation

- [ ] Add an asset model with hash, media type, dimensions, variants, source, license/usage proof,
  attribution, lifecycle state, and links to catalog records.
- [ ] Store only owned, licensed, public-domain, or otherwise explicitly permitted cover assets;
  do not bulk-download third-party thumbnails or hotlink images as a substitute for rights.
- [ ] Add bounded upload/fetch validation: byte-level type detection, size and dimension limits,
  image decoding, content-addressed deduplication, and safe failure handling.
- [ ] Introduce a small storage boundary with local development storage and an object-storage-ready
  production adapter; do not add full-text ebook storage, image editing, or a media CDN.

### Operator UX and Delivery

- [ ] Translate catalog correction, source admission, import review, and batch-result screens into
  `en` and `uk`; preserve field-level language/source evidence in either operator locale.
- [ ] Keep stage separate from the public environment and automate stage deployment of versioned
  artifacts; retain an explicit review before public batch publication or application promotion.

### Definition of Done

- [ ] One accepted source can be imported through the shared job model without duplicate logical
  catalog records and with per-field provenance.
- [ ] A failed, resumed, or repeated batch preserves catalog consistency and has an auditable
  result.
- [ ] Users can inspect the source and limitations of published batch data.
- [ ] The application stores and serves only cover assets with recorded usage rights and
  attribution.
- [ ] Source, import, and media risks are covered by focused tests, operator documentation, and
  metrics/logs appropriate to batch processing.
- [ ] Operator import/review flows work in both languages, and the stage deployment is reproducible.

Not in v0.5: scraping protected sites; treating a third-party API as the permanent catalog
database; automatic unreviewed merge of all sources; anonymous arbitrary-URL imports; migration of
user reviews/social data; full-text ebook hosting; or a general-purpose media platform.
