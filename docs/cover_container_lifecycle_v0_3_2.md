# Pilot Container Lifecycle (#28, acceptance step 4)

Verified on 2026-10-30 (Europe/Kyiv; Docker timestamps below are UTC on 2026-10-29).
The user explicitly authorized a sandbox app/container run for this step only. No retained
dev/stage/prod containers, images or volumes were modified. This is actual SQLite Compose
restart/recreation evidence, not a PostgreSQL lifecycle or backup/restore rehearsal.

## Environment

- Application revision: `29b16888d8fa07a489e90c8c70201df7c517c92a`; pre-existing documentation
  edits were preserved. No application code/configuration changes.
- Docker Engine `29.9.1`, Compose `v5.6.0`; access required sandbox escalation.
- Fresh project `book-social-cover-step4-lk4lab`; initial project container/volume listings empty.
- Image `book-social-cover-step4-lk4lab:local`, built from the current Dockerfile:
  `sha256:f347346b046fdb8f6460afa0397b7692efac48981755eb671ca38f27c06de561`.
- Used actual `compose.yaml` + `compose.dev.yaml`; temporary override changed only image,
  container name and port. DB/cover mounts and entrypoint remained the normal dev configuration.
- Named volumes `book-social-cover-step4-lk4lab_sqlite-data` and
  `book-social-cover-step4-lk4lab_covers-dev-data`, mounted at `/app/data` and `/app/covers`.
- Loopback-only dynamic port (`127.0.0.1::8080`), initially `40853`; port rediscovered after
  each lifecycle operation rather than assumed stable.

Temporary override `/tmp/book-social-cover-lifecycle.lK4lAb/compose.acceptance.yaml`:

```yaml
services:
  app:
    container_name: book-social-cover-step4-lk4lab-app
    image: book-social-cover-step4-lk4lab:local
    ports: !override
      - "127.0.0.1::8080"
```

For the commands below, `dc` denotes this exact Compose invocation from the repository root:

```bash
dc() {
  docker compose -p book-social-cover-step4-lk4lab \
    -f compose.yaml -f compose.dev.yaml \
    -f /tmp/book-social-cover-lifecycle.lK4lAb/compose.acceptance.yaml "$@"
}
```

Do not reuse this project as a fresh environment: its test volumes now contain data. The initial
setup explicitly required the target SQLite DB to be absent and cover directory empty. Applied
migrations through `8|0`, seeded once, and explicitly provisioned a non-login admin fixture
(`lifecycle-audit-only@example.test`, ID 1, invalid password hash) for audit only. Used one-shot
containers with `/bin/sh` entrypoint for setup and `/app/coverctl` for maintenance, not bootstrap.
Installed approved bundle `pilot-v1-approved-2026-10-25` from a read-only mount with pinned
manifest `4f80285da167fbf1695d58b5ec96abbe125cf222f9ff1e3f78ee248b9f7de6f8`: five changes.

## Executed Lifecycle and Results

Executed `dc config --quiet`, `dc build app`, then installation and initial stopped-writer verify.
Started app with `dc up -d --no-build app`. At each running checkpoint, an HTTP probe used
`node /tmp/book-social-cover-lifecycle.lK4lAb/http-check.cjs "http://$(dc port app 8080)"`.
The probe checked status 200 and main-image URL on all five detail pages; status 200, WebP MIME
and SHA-256 against the approved manifest for all ten images; image URLs on `/books`, `/` and
the five pilot author pages; and Christmas's thumbnail in the HTMX catalog fragment obtained
with `HX-Request: true` at `/books?author=charles-dickens`.

| Checkpoint | Container ID prefix | StartedAt (UTC) | SQL dump / all ten file hashes | HTTP probe |
|------------|---------------------|-----------------|-------------------------------|------------|
| Initial | `806ef4339e66` | 22:41:24.483872976 | Baseline | Pass |
| `dc restart app` | `806ef4339e66` | 22:41:41.730151499 | Identical | Pass |
| `dc up -d --no-build --force-recreate app` | `16666f64cffd` | 22:41:53.529491314 | Identical | Pass |

Before recreation, stopped app with `dc stop app` and ran
`dc run --rm -T --no-deps --entrypoint /app/coverctl app verify -writers-stopped`.
After the final HTTP checkpoint, stopped app again and repeated the same verify. All three
maintenance verifies (initial, after restart, after recreation) reported five local cover references
including thumbnails. No app/other writer ran during maintenance.

At every checkpoint the following checks passed:

```bash
dc exec -T app sqlite3 -readonly /app/data/book_social_dev.db .dump | sha256sum
dc exec -T app /bin/sh -ec 'sha256sum /app/covers/*'
dc exec -T app sqlite3 -readonly /app/data/book_social_dev.db \
  'SELECT COUNT(id) FROM books; SELECT COUNT(id) FROM covers; SELECT COUNT(id) FROM catalog_audit_log; PRAGMA integrity_check; PRAGMA foreign_key_check;'
```

The entire SQL dump SHA-256 stayed
`6e332156b18b19f18a92ab9d6a79ffe994a978d3f5014a2c07282d933feef698`.
Counts stayed 109 books, five covers, five audits; integrity was `ok`, foreign-key check empty.
All ten canonical file hashes stayed identical. `docker inspect` showed the same two named
volume mounts before and after recreation. Restart kept container ID but changed StartedAt;
recreation changed container ID. Seed was not replayed and covers were not reinstalled.

## Retained Test Resources and Boundaries

The test app is stopped (`Exited (0)`); its two test volumes, image, container and network are
retained for inspection/next-step work. Nothing material was deleted; no `down -v` or prune ran.
Temporary override/probe files are under `/tmp/book-social-cover-lifecycle.lK4lAb`; `/tmp` is not
durable evidence storage. This checked-in protocol records observations and actual commands.

HTTP checks do not constitute a new visual browser test after recreation. Prior user-reported
Zen display/fallback acceptance remains recorded separately. PostgreSQL lifecycle repetition,
actual paired SQLite/PostgreSQL backup/restore and remaining detailed browser checks remain open.
No full unit suite was rerun: this pass changed documentation only and tested actual runtime
lifecycle directly. Compose validation, image build, HTTP/file/SQL checks, stopped-writer verify
and `git diff --check` passed. No commit, issue closure or production deployment.
