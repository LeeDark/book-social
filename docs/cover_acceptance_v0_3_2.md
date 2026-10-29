# Cover Verification and Pending Acceptance (#28, Step 7)

Status on 2026-10-25: automated verification passes on implementation `27c64d2`.
Step 7 is **not fully accepted**: PostgreSQL lifecycle and paired operational recovery remain pending.
Reviewed artwork, isolated SQLite installation and user-reported pilot browser checks are recorded below.
This record does not close #28 or the expanded v0.3.2 scope
and does not replace the earlier acceptance of the original #23–#26 workflows.

The [contract](cover_storage_v0_3_2.md) defines behavior; [cover operations](cover_operations.md)
defines installation, old-storage migration precautions and paired SQL/file recovery.
No implementation, migration or dependency files, or retained environments, were changed during
this pass.

## Automated Evidence

The following commands completed successfully:

```bash
GOTOOLCHAIN=go1.26.4 GOCACHE=/tmp/book-social-go-cache \
  go test -race -count=1 ./internal/coverops ./cmd/coverctl ./internal/storage/covers ./internal/http/cover
GOTOOLCHAIN=go1.26.4 GOCACHE=/tmp/book-social-go-cache make test
GOTOOLCHAIN=go1.26.4 GOCACHE=/tmp/book-social-go-cache make db/migrate/smoke
GOTOOLCHAIN=go1.26.4 GOCACHE=/tmp/book-social-go-cache go vet ./...
GOTOOLCHAIN=go1.26.4 GOCACHE=/tmp/book-social-go-cache GOLANGCI_LINT_CACHE=/tmp/book-social-lint-cache \
  bin/golangci-lint run ./... --config=./.golangci.yml
node --test internal/web/tests/*.test.cjs
```

Full tests include the race detector; lint reported zero issues. The normal full suite skips
opt-in PostgreSQL cases. Those were then run against real PostgreSQL 18 in an initially empty,
separate disposable Compose project:

```bash
GOTOOLCHAIN=go1.26.4 GOCACHE=/tmp/book-social-go-cache make test/integration \
  COMPOSE_INTEGRATION='docker compose -p book-social-cover-step7 -f compose.integration.yaml'
```

Migration `000008` passed up/down/up; `internal/testutil`, `internal/storage/postgresql`,
`internal/app` and `internal/coverops` all passed sequentially with race detection. This includes
the previously pending real PostgreSQL card/HTMX rendering and five-slug pilot install/verify
tests from steps 5–6. The existing cleanup removed its test container/database and network;
a final project listing was empty. Retained dev/stage/prod databases/volumes were not accessed.
Docker socket access required sandbox escalation. PostgreSQL installation tests are not evidence
of a PostgreSQL dump/restore rehearsal.

Also passed:

```bash
GOTOOLCHAIN=go1.26.4 CGO_ENABLED=0 GOOS=linux GOCACHE=/tmp/book-social-go-cache \
  go build -o /tmp/book-social-cover-step7-web ./cmd/web
GOTOOLCHAIN=go1.26.4 CGO_ENABLED=0 GOOS=linux GOCACHE=/tmp/book-social-go-cache \
  go build -o /tmp/book-social-cover-step7-ctl ./cmd/coverctl
docker compose -f compose.yaml -f compose.dev.yaml config --quiet
docker compose -f compose.yaml -f compose.stage.yaml config --quiet
docker compose -f compose.yaml -f compose.prod.yaml config --quiet
rg --files -g '*.go' | xargs gofmt -l
git diff --check
```

Both builds exited successfully with a non-fatal read-only module stat-cache warning. Formatting
listed no files. Compose validation checks configuration only, not mounted-file persistence.
The image build passed in step 6; no web server or app container was started for this pass.
Temporary verification logs are in `/tmp/book-social-cover-step7-*.log`; this checked-in summary,
not retention of those temporary logs, records the results.

## Coverage Boundaries

| Criterion | Evidence | Remaining boundary |
|-----------|----------|--------------------|
| Format/size/dimensions, unsafe paths, atomic immutable publication and corrupt-file refusal | `internal/storage/covers/*_test.go`, `internal/http/cover/handler_test.go` | Actual storage filesystem/permissions in the chosen environment |
| Create/upload/keep/replace/external/remove, authorization, audit rollback and temporary cleanup | `internal/app/operator_cover_http_test.go`, books service/adapter tests and shared SQLite/PostgreSQL acceptance | Browser file-input workflow and selected artwork |
| Front-only rendering, main/thumbnail choice, catalog/home/author/library and HTMX fragments | `internal/app/cover_rendering_integration_test.go`, books/library view tests | Actual mobile/desktop layout, accessibility and visual quality |
| Cached/pending image failures and dynamic HTMX fallback | `internal/web/tests/cover-images.test.cjs` | DOM doubles are not a browser/network rendering test |
| Whole bundle validation, five slug targets, idempotence, audit failures and all-reference verification | `internal/coverops/*_test.go`, `cmd/coverctl/main_test.go` | Synthetic PNGs are not approved pilot illustrations or a frozen delivery profile |
| Persistence/recovery | Storage close/reopen and fresh SQLite snapshot/file restore tests; Compose config and step-6 image build | Container restart/recreate, fresh reviewed delivery and real PostgreSQL dump/restore |

## Required Local Acceptance

Pilot preparation update on 2026-10-25: five candidate illustrations and their draft manifest,
sources, prompts/provenance and checksums are prepared in
[`assets/demo-covers/pilot-v1`](../assets/demo-covers/pilot-v1/README.ru.md). The
[offline gallery](../assets/demo-covers/pilot-v1/review.html) shows real 240×360 thumbnails and
links to 600×900 main files. The installer deliberately rejects this historical draft.

The user subsequently approved all five illustrations and the delivery profile on 2026-10-25.
The separate [approved release](../assets/demo-covers/pilot-v1-approved/README.ru.md) passed
actual offline installation twice in a fresh SQLite environment: five changes then zero,
five audits total, five main/thumbnail pairs verified. See the
[installation protocol](cover_pilot_installation_v0_3_2.md) for hashes, commands and preservation
checks. Installation alone does not imply browser, container or recovery acceptance.

### User-Reported Browser Evidence

Recorded in documentation on 2026-10-30 from the user's confirmations in this conversation on
2026-10-26. Browser: Zen; mobile-sized responsive viewport. Exact browser version and viewport
dimensions were not supplied. This is user-reported local testing, not an agent-run browser test.

The user confirmed correct cover display on pages and all five books in their private library;
then confirmed mobile proportions, A Christmas Carol's home card and HTMX filtering. A supplied
screenshot shows Pride and Prejudice's details page at mobile width. Finally, after following the
request-blocking procedure for `/media/covers/`, the user confirmed the expected placeholders,
HTMX fallback and recovery after unblocking. Files and SQL were not removed to simulate failure.

These confirmations accept the reported pilot display/fallback checks, not every independent
criterion below. Desktop sizing, explicit keyboard/accessibility checks, main-only delivery,
external HTTPS compatibility and JavaScript-disabled placeholders were not separately reported.
At the time of those browser reports, container restart/recreation and paired recovery were unverified. Storage
close/reopen tests, repeated offline installation and Compose configuration validation are not
evidence of actual container restart or recreation.

For remaining browser/container/recovery checks, use a disposable local environment outside the
Codex sandbox. Do not experiment on retained data
or remove cover files to simulate failure. Before upgrading an old container, preserve its old
cover directory as described in [the transition warning](cover_operations.md#persistent-container-storage).

- [x] Supply five reviewed artworks, source/provenance/usage evidence, version and pinned
  manifest/archive checksums. WebP/2:3 600×900 main and 240×360 thumbnail approved for the pilot;
  user-reported mobile pilot display/fallback checks are recorded above.
- [x] Provision fresh matching migrations/seed and an explicit administrator; install twice and
  record changed/audit counts and `coverctl verify` results (offline disposable SQLite).
- [x] User confirms pilot covers on pages and all five in the private library, mobile display,
  HTMX filtering and A Christmas Carol's featured home card.
- [x] User confirms unavailable-image placeholders, including HTMX, with browser request blocking,
  and restored images after unblocking; load-error fallback requires JS.
- [ ] Separately verify main-only delivery, external HTTPS compatibility, desktop presentation,
  keyboard links and accessible labels, and missing-metadata placeholders with JavaScript disabled.
- [x] Restart/recreate SQLite app with the same named volumes: SQL/files, HTTP and stopped-writer verify passed.
  See [lifecycle protocol](cover_container_lifecycle_v0_3_2.md).
- [ ] Repeat lifecycle checks for PostgreSQL and visually inspect pages after recreation.
- [ ] Rehearse matching SQL/file backup and restore into fresh SQLite and PostgreSQL targets;
  compare checksums, migration state, reference counts, metadata and HTTP/browser results.
- [ ] Record environment, application/image revision, actual commands, bundle checksums,
  browser/version, viewport sizes, date, observed results and the user's acceptance decision.

Routes for the agreed seeded pilot:

```text
/
/books
/books?author=charles-dickens
/books/a-christmas-carol
/books/pride-and-prejudice
/books/frankenstein
/books/alices-adventures-in-wonderland
/books/twenty-thousand-leagues-under-the-seas
/authors/charles-dickens
/authors/jane-austen
/authors/mary-shelley
/authors/lewis-carroll
/authors/jules-verne
/me/library
/admin/catalog
/admin/catalog/books/new
```

Add the pilot books to the test user's library first; find their edit pages through the operator
search rather than assuming numeric IDs. The exact install/verify/backup/restore and container
recreation commands are in [cover operations](cover_operations.md). Stop all writers for maintenance,
verify the actual target and do not rerun seed against an existing catalog. No browser or
operational acceptance is inferred from the automated results above.
