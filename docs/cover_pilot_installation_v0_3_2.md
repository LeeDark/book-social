# Approved Pilot Installation (#28, acceptance step 2)

Recorded on 2026-10-25, application revision `dfb59bd82c327a4baa9625cd78eae29594336e55`.
No application code, migrations or dependencies changed. This is an actual offline SQLite
installation of approved artwork, not the synthetic installation test or browser acceptance.

## Approval and Frozen Delivery

The project user explicitly answered «Да, подтверждаю.» to approval of all five illustrations
and the WebP 600×900 main / 240×360 thumbnail profile for the demo catalog pilot.
The decision, date, reviewer and scope are recorded in
[approved provenance](../assets/demo-covers/pilot-v1-approved/artwork-provenance.json).
The original draft is retained unchanged; generated PNGs and all ten WebPs are byte-identical.
Approval does not imply legal clearance or approval of the remaining 104 illustrations.

- Bundle: `pilot-v1-approved-2026-10-25`.
- [Manifest](../assets/demo-covers/pilot-v1-approved/bundle/manifest.json) SHA-256:
  `4f80285da167fbf1695d58b5ec96abbe125cf222f9ff1e3f78ee248b9f7de6f8`.
- Local complete archive: `assets/demo-covers/pilot-v1-approved.tar.gz` (ignored by Git).
- Archive SHA-256: `ec447dccefe6a93061991d5b1e2a6164182753f1e4dfc04c1617e370fcb088a4`,
  independently recorded in [the digest file](../assets/demo-covers/pilot-v1-approved.tar.gz.sha256).
- [SHA256SUMS](../assets/demo-covers/pilot-v1-approved/SHA256SUMS) covers all 20 payload files,
  including original PNGs, WebPs, prompts, provenance, manifest, gallery and README.

## Isolated Target and Commands

Created with `mktemp -d /tmp/book-social-pilot-approved.XXXXXX`:
`/tmp/book-social-pilot-approved.sJzxYR`. No retained database, container or volume was accessed.
No app/server or other writer was started. This directory remains available for inspection;
`/tmp` is not durable storage or a backup policy.

From the repository root:

```bash
migrate -path ./db/sqlite/migrations \
  -database sqlite:///tmp/book-social-pilot-approved.sJzxYR/catalog.sqlite up
sqlite3 -bail /tmp/book-social-pilot-approved.sJzxYR/catalog.sqlite < db/sqlite/seed.sql
```

Explicit disposable administrator provisioning used `sqlite3 -bail`:

```sql
PRAGMA foreign_keys = ON;

INSERT INTO users (id, first_name, login, password_hash, email, user_role_id)
VALUES (
    1, 'Pilot audit only', 'pilot-audit-only', 'disabled-not-a-password-hash',
    'pilot-audit-only@example.test',
    (SELECT id FROM roles WHERE role_name = 'admin')
);
```

This is a non-login audit fixture, not a default administrator or usable password. Browser tests
must provision a separate valid local account explicitly. The seed ran once on this fresh DB only.
Migration state was `8|0`, with 109 books, zero covers and zero audits.

```bash
mkdir /tmp/book-social-pilot-approved.sJzxYR/covers
GOTOOLCHAIN=go1.26.4 GOCACHE=/tmp/book-social-go-cache \
  go build -o /tmp/book-social-pilot-approved.sJzxYR/coverctl ./cmd/coverctl
```

Build succeeded with a non-fatal read-only module stat-cache warning. Every maintenance invocation
used these explicit targets (shown as exports for readability):

```bash
export APP_ENV=dev
export APP_DB_DSN=/tmp/book-social-pilot-approved.sJzxYR/catalog.sqlite
export APP_COVERS_DIR=/tmp/book-social-pilot-approved.sJzxYR/covers

/tmp/book-social-pilot-approved.sJzxYR/coverctl verify -writers-stopped
/tmp/book-social-pilot-approved.sJzxYR/coverctl install -writers-stopped \
  -bundle /home/lee/code/github/LeeDark/book-social/assets/demo-covers/pilot-v1-approved/bundle \
  -manifest-sha256 4f80285da167fbf1695d58b5ec96abbe125cf222f9ff1e3f78ee248b9f7de6f8 -actor-id 1
/tmp/book-social-pilot-approved.sJzxYR/coverctl verify -writers-stopped
```

The identical install command and final verify were then executed again.

| Observation | Before | After first install | After second install |
|-------------|--------|---------------------|----------------------|
| Changed front covers in invocation | — | 5 | 0 |
| Verified local covers (each includes its thumbnail) | 0 | 5 | 5 |
| Audits | 0 | 5 | 5 |

All five audits have actor `1`, entity `book`, action `correct`, summary `front_cover`.
Resolved seed IDs: Christmas `1`, Pride `49`, Frankenstein `22`, Alice `7`, Leagues `104`.
These are observations, not installation inputs or stable-ID assumptions.

## Integrity, Preservation and Checks

- Compared read-only SQLite `.dump` output before/after the first install for all 14 unrelated
  tables: authors, book_authors, book_genres, books, genres, library, library_items,
  missing_book_requests, roles, schema_migrations, sessions, shelves, tags and users.
  Their combined 48,476-character dump was identical. All 109 books and relationships were
  preserved, including the 104 non-pilot books. No cover variants existed initially; preservation
  of existing other variants remains covered by synthetic tests, not this fixture.
- Compared the complete `.dump` after first and second install: identical, including audits
  and SQLite sequences. Reinstallation made no SQL changes.
- `PRAGMA integrity_check;` returned `ok`; `PRAGMA foreign_key_check;` returned no rows.
- SQL measures all five front covers at 600×900 with 240×360 thumbnails. `coverctl verify`
  validated all referenced decoded bytes, hashes, metadata and pairs after both installations.
- `stat -c '%a %n' /tmp/book-social-pilot-approved.sJzxYR/covers/*` showed exactly ten files,
  all mode `640`. `sha256sum /tmp/book-social-pilot-approved.sJzxYR/covers/*` matched every
  canonical `<sha256>.webp` name and corresponding approved manifest entry.
- Read-only Node assertions checked five unique reviewed slugs, approval evidence, all recorded
  source/derivative byte counts and SHA-256, and byte equality with the draft.
- `sha256sum -c SHA256SUMS` in each artwork directory, archive checksum verification and
  `gzip -t assets/demo-covers/pilot-v1-approved.tar.gz` passed.
- `GOTOOLCHAIN=go1.26.4 GOCACHE=/tmp/book-social-go-cache go test -race -count=1 ./internal/coverops ./cmd/coverctl`
  and `git diff --check` passed. PostgreSQL cases remain opt-in and were not rerun here.
  No full application test run was needed for asset/metadata/documentation-only changes.

Browser/HTMX/mobile/desktop acceptance, app/container restart and recreation, actual paired
SQLite/PostgreSQL recovery and remaining artwork are still outstanding. This rehearsal does
not claim operational PostgreSQL delivery, start a server in the sandbox, or close #28.
See the [acceptance checklist](cover_acceptance_v0_3_2.md) for next steps.
