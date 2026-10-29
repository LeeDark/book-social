# Paired Pilot Backup/Restore (#28, acceptance step 5)

Offline recovery verified on 2026-10-30, revision `542f705`. No application code, migrations,
dependencies or retained dev/stage/prod data changed. The initial pass was offline only because
web-app permission was limited to step 4. The user then explicitly authorized restored-state app
execution for step 5; both engines passed the HTTP checks recorded below. Visual browser review
is not inferred from HTTP responses and remains separate.

## Targets and Safety

Private workspace: `/tmp/book-social-cover-recovery.LhkEv5`, mode 0700; backup payloads mode
0600. This is a disposable rehearsal, not durable backup retention. No private SQL/files entered
Git. Sources, snapshots and restored targets were retained; no cleanup, drop or overwrite ran.

SQLite source is the stopped step-4 test container's volumes:
`book-social-cover-step4-lk4lab_sqlite-data` and `book-social-cover-step4-lk4lab_covers-dev-data`.
Confirmed its app was stopped; mounted both source volumes **read-only**, with no network.
Recovery tools used the same app image as step 4:
`sha256:f347346b046fdb8f6460afa0397b7692efac48981755eb671ca38f27c06de561`.

PostgreSQL source was provisioned separately for this rehearsal, not taken from a working stage
database. Fresh Compose project `book-social-cover-step5-lhkev5`, image `postgres:18-alpine`
(`sha256:bd1890816ae0b8ad4644f05728570d4be774e1f1490d7232f5084b52ea335183`), no host ports.
New volume: `book-social-cover-step5-lhkev5_postgres-data`; source DB `cover_source`, destination
DB `cover_restore`, test owner `cover_test`. Compose file remains at
`/tmp/book-social-cover-recovery.LhkEv5/compose.yaml`.

Applied PostgreSQL migrations through 8 and seed once to the new source, explicitly provisioned
an audit-only administrator (ID 1, `recovery-audit-only@example.test`, unusable password hash),
then installed approved pilot `pilot-v1-approved-2026-10-25` with its pinned manifest SHA-256
`4f80285da167fbf1695d58b5ec96abbe125cf222f9ff1e3f78ee248b9f7de6f8`: five changes.
Original artwork/provenance is retained in Git revision `29b1688` at
`assets/demo-covers/pilot-v1-approved`. No application or other data/cover writer ran during
either paired snapshot. PostgreSQL maintenance connections were sequential.

## SQLite Snapshot and Fresh Restore

One-shot container bypassed bootstrap, with source volumes mounted at `/source-db` and
`/source-covers`, and `sqlite-backup` mounted at `/backup`:

```bash
sqlite3 -readonly /source-db/book_social_dev.db ".backup /backup/catalog.sqlite"
tar -C /source-covers -cf /backup/covers.tar .
```

Inspected `tar -tf` (ten canonical WebPs), copied `catalog.sqlite` into the new `sqlite-restore`
directory and extracted with `tar --no-same-owner` into its initially empty `covers` directory.
Ran `PRAGMA integrity_check; PRAGMA foreign_key_check;` and queried migration/book/cover/audit
counts with host `sqlite3 -readonly`. Ran the same image's `/app/coverctl verify -writers-stopped`
with `APP_ENV=dev`, explicit `/restore/catalog.sqlite` and `/restore/covers`, and network disabled.

- Integrity `ok`; foreign-key check empty; migration `8|0`.
- 109 books, five cover rows, five audits.
- Verified five local main/thumbnail pairs.
- Entire source/restored `.dump | sha256sum` identical:
  `6e332156b18b19f18a92ab9d6a79ffe994a978d3f5014a2c07282d933feef698`.
  Thus all SQL URLs, metadata, relationships and audit rows match, not just counts.

## PostgreSQL Snapshot and Fresh Restore

Used `pg_dump` and `pg_restore` **18.4 inside the PostgreSQL container**, not the app image's
older client. In the isolated Compose project, executed:

```bash
pg_dump -U cover_test -d cover_source --format=custom --file=/tmp/cover-source.dump
createdb -U cover_test cover_restore
psql -X -v ON_ERROR_STOP=1 -U cover_test -d cover_restore \
  -c "SELECT COUNT(tablename) FROM pg_tables WHERE schemaname = 'public';"
pg_restore --exit-on-error --single-transaction --no-owner --no-privileges \
  -U cover_test --dbname=cover_restore /tmp/cover-source.dump
```

The empty destination check returned zero before restore. No migration/seed was run on the
destination. Copied the custom archive from the container into private `pg-backup/catalog.dump`;
archived the complete stopped-writer source cover directory into `pg-backup/covers.tar`.
Inspected archive names, verified backup checksums and extracted into the initially empty
`pg-restored-covers`. `cover_restore` used these separate restored files, not the source directory.

Both source and destination passed `/app/coverctl verify -writers-stopped` with explicit
`APP_ENV=stage`, database DSN and cover directory; destination covered all five main/thumbnail
pairs. Restored migration `8|false`; 109 books, five covers, five audits. No unvalidated public
constraints were returned by the query against `pg_constraint`.

Compared full source/restored `pg_dump --inserts --no-owner --no-privileges` output, excluding
only random `\restrict`/`\unrestrict` guard tokens. A strict textual comparison initially failed:
PostgreSQL re-emitted two equivalent CHECK expressions (`ck_catalog_audit_action` and the MIME
list in `ck_covers_local_file`) with casts on each array element rather than on the whole array.
Inspected the exact two changed lines: all schema/data/sequence/index lines otherwise match.
This is a documented equivalent cast representation, not byte-identical schema text or data loss.

## Snapshot Checksums and File Evidence

Each backup has a local `SHA256SUMS`, verified with `sha256sum -c SHA256SUMS`:

| Snapshot payload | SHA-256 |
|------------------|---------|
| SQLite `catalog.sqlite` | `94a42605910f16a679788b89f10e7a7a0b73944783af8bbb93a8a6a1b4e66f95` |
| SQLite `covers.tar` | `ff3c0297326b41c7eba8c8e3ab573d235596d145b19cda273310802b99f0b21b` |
| PostgreSQL `catalog.dump` | `8e9cfba04a05e985c98cc19d69e05fd2642f0476b80cb96d2aea58b1b356e507` |
| PostgreSQL `covers.tar` | `46869c59179cf326dee0c18f503df1a2a59751af97c76522fe8c6cbade6cd0c5` |

The tar hashes differ because archive metadata/timestamps differ; restored image bytes do not.
Read-only Node assertions verified exactly ten files in each restored cover directory and compared
every SHA-256 and complete byte buffer against the approved manifest/bundle (20 restored WebPs).
All matched. The fixture contains no extra unreferenced files; their retention follows full-directory
archiving but was not exercised with an extra-file fixture here.

## Final State and Remaining Work

SQLite source remains stopped/unchanged. Test PostgreSQL was stopped cleanly (`Exited (0)`),
retaining both test DBs and its named volume/network. All private snapshots and restored files
remain under the private temporary workspace. Host chmod initially failed for container-owned
backup files; exact-file chmod through their container owner succeeded. Docker access required
sandbox escalation. No broad chmod/chown or destructive Git/Docker operation was used.

Offline SQL/file recovery checks passed for both engines. No full application test suite was rerun
for documentation-only changes; `git diff --check` passed. This is not a production backup policy:
choose durable protected backup storage separately before relying on retention. The follow-up
below completes HTTP checks with separate explicit authorization, not a silent extension of
the step-4-only permission. PostgreSQL restart/recreation and detailed visual review remain separate.

## Authorized Restored-State HTTP Follow-Up

On 2026-10-30 the user explicitly authorized the test application run for step 5. Added only a
temporary Compose override `/tmp/book-social-cover-recovery.LhkEv5/compose.http.yaml` to the
existing step-5 project. It defines `sqlite-app` and `pg-app`, both using the same image as the
offline verification, with **entrypoint `/app/book-social`** to bypass bootstrap/migrations/seed.

- SQLite: bind `/tmp/book-social-cover-recovery.LhkEv5/sqlite-restore` at `/restore`, explicit
  `APP_ENV=dev`, `APP_DB_DSN=/restore/catalog.sqlite`, `APP_COVERS_DIR=/restore/covers`.
- PostgreSQL: explicit `APP_ENV=stage`, DSN targeting `cover_restore` (not `cover_source`),
  `pg-restored-covers` mounted read-only at `/covers`, explicit `APP_COVERS_DIR=/covers`.
- Both listen inside their container at `:8080`, published only to loopback dynamic ports:
  SQLite `127.0.0.1:40925`, PostgreSQL `127.0.0.1:40924`.

From the repository root, the actual Compose invocation was:

```bash
rc() {
  docker compose -p book-social-cover-step5-lhkev5 \
    -f /tmp/book-social-cover-recovery.LhkEv5/compose.yaml \
    -f /tmp/book-social-cover-recovery.LhkEv5/compose.http.yaml "$@"
}

rc config --quiet
rc up -d --wait postgres
rc up -d --no-build --no-deps sqlite-app pg-app
node /tmp/book-social-cover-lifecycle.lK4lAb/http-check.cjs "http://$(rc port sqlite-app 8080)"
node /tmp/book-social-cover-lifecycle.lK4lAb/http-check.cjs "http://$(rc port pg-app 8080)"
rc stop sqlite-app pg-app
```

Both HTTP probes passed:

- Five detail pages return 200 and reference the approved main-image URL.
- All ten main/thumbnail URLs return 200, `image/webp`, and bytes with the approved SHA-256.
- Catalog, home and five pilot author pages return 200 with local cover URLs.
- Catalog filtering with `HX-Request: true` returns the expected Christmas thumbnail.

After stopping both apps, repeated `/app/coverctl verify -writers-stopped` against each restored
DB/file target: both verified five pairs. SQLite integrity was `ok`, foreign-key check empty,
and full dump hash remained `6e332156b18b19f18a92ab9d6a79ffe994a978d3f5014a2c07282d933feef698`.
PostgreSQL migration/counts remained 8|false and 109/5/5; full logical dump before/after HTTP was
identical excluding random restrict tokens. No schema/data changes were made by these HTTP checks.

Finally ran `rc stop postgres`; all three containers exited 0 and remain stopped. Source/restored
data, snapshots, containers and test volume/network are retained. No real browser was driven;
these probes do not assert layout, JavaScript execution, authenticated library UI or visual
quality after restore. Prior Zen pilot acceptance remains separate. No full suite was rerun;
Compose validation, both HTTP probes, offline verification and `git diff --check` passed.
