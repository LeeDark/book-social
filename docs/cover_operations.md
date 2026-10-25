# Cover Delivery, Persistence and Recovery (#28, Step 6)

This is a local operator procedure, not production deployment automation. Use the image and
migrations from the same application revision. Apply migrations through `000008` before maintenance.
The installer never migrates, seeds, downloads, resizes or deletes files. No approved pilot artwork
ships yet; solid-color files in tests are not demo artwork or evidence of visual acceptance.

## Persistent Container Storage

All Compose environments set `APP_COVERS_DIR=/app/covers`. Each has a separate named volume:

| Environment     | Cover volume key    | Database storage             |
|-----------------|---------------------|------------------------------|
| dev             | `covers-dev-data`   | `sqlite-data` at `/app/data` |
| stage           | `covers-stage-data` | `postgres-stage-data`        |
| prod-like local | `covers-prod-data`  | `postgres-prod-data`         |

Compose prefixes volume names with the project name. Keep the same project name and override
file when recreating an environment; another project does not automatically inherit these volumes.
`make compose/dev/down`, `make compose/stage/down` and `make compose/prod/down` now retain volumes.
`up --build --force-recreate` keeps them too. Do not use `down -v`, volume pruning or an anonymous
volume for retained environments. A database-only reset does not delete or reattach cover files.

The image prepares `/app/covers` with mode `0750`, owned by the existing `app` account; publication
creates files with mode `0640`. The current image still runs as root, as before this change;
creating
an `app` account is not a claim of non-root execution. For an explicit user override or a bind
mount,
prepare permissions for that actual UID/GID before starting. Do not fix a retained mount with an
unreviewed recursive chmod/chown. The filesystem must support hard links and directory syncing.

Existing dev installs may have files at `/app/data/covers` inside `sqlite-data`. Before upgrading,
stop writers, take the paired backup below, inventory/verify those files and copy them to the new
cover volume. Do not delete the old copy until all SQL references verify. Earlier stage/prod
containers may have cover bytes only in their writable container layer: recover that directory
from the stopped old container **before** removing/recreating it. A missing old directory/volume
cannot be repaired from SQL metadata or the Docker image alone.

## Versioned Five-Book Bundle

Keep a reviewed bundle outside the public cover directory. `manifest.json` is UTF-8 JSON, at most
64 KiB. It requires `version: 1`, a non-empty `bundle_id`, and exactly one entry for each pilot
slug:

```text
a-christmas-carol
pride-and-prejudice
frankenstein
alices-adventures-in-wonderland
twenty-thousand-leagues-under-the-seas
```

Entry shape (abbreviated; the real array must include all five):

```json
{
    "version": 1,
    "bundle_id": "pilot-v1",
    "covers": [
        {
            "slug": "a-christmas-carol",
            "reviewed": true,
            "usage_basis": "Project demo illustration; generation, references and usage evidence in artwork-provenance.json.",
            "main": {
                "file": "christmas-main.webp",
                "sha256": "REPLACE_WITH_REVIEWED_FILE_SHA256"
            },
            "thumbnail": {
                "file": "christmas-thumb.webp",
                "sha256": "REPLACE_WITH_REVIEWED_FILE_SHA256"
            }
        }
    ]
}
```

Omit `thumbnail` for main-only delivery. File paths must be flat regular filenames directly inside
the bundle; symlinks, traversal and absolute paths are refused. Unknown JSON fields are refused.
Files need lowercase SHA-256 checksums, supported extensions and fully decodable portrait bytes;
the existing size/dimension/pair validation applies. All five entries/files are validated in memory
before publishing any file. Retain source artwork, detailed generation/reference/usage evidence,
review decisions and an independently recorded checksum of the complete bundle archive. These
supporting files are retained, not copied into public storage. A `reviewed` flag is an operator
attestation, not an automated rights or visual-quality check.

Freeze and record the manifest SHA-256 during review in release/delivery records. At installation,
pin that previously reviewed value; calculating a new hash from an untrusted download and passing
it straight back is not authenticity verification. Also verify the downloaded bundle archive
against its independently recorded release checksum before unpacking into a fresh directory.

## Explicit Installation

Stop the app and every other cover writer/maintenance job. Resolve uncertain earlier commits and
take a paired backup first. `-writers-stopped` is an explicit confirmation, not an automatic lock.
Do not run two installers together. Use an existing administrator's ID for the normal audit.

Host example (set the database and storage paths explicitly):

```bash
APP_ENV=dev APP_DB_DSN=/absolute/path/book_social.db APP_COVERS_DIR=/absolute/path/covers \
  go run ./cmd/coverctl install -writers-stopped \
  -bundle /absolute/path/reviewed-pilot -manifest-sha256 "$PILOT_MANIFEST_SHA256" -actor-id 1
```

Both `APP_DB_DSN` and `APP_COVERS_DIR` must be explicitly set; maintenance refuses application
defaults for those targets. The SQLite database must already exist; the command will not silently
create/seed a target.
For PostgreSQL, use `APP_ENV=stage` or `prod` and the explicitly selected `APP_DB_DSN`.
Resolve every slug and check authorization before publishing. All references and ordinary
`correct`/`front_cover` audits commit together after authorization/target rechecks. Unrelated book
metadata, relationships and cover variants are preserved. Identical reinstallation reports zero
changes and appends no audits; all files are still validated and republished/reused safely first.
Storage/SQL failure retains published bytes without automatic cleanup. After an uncertain commit,
inspect the actual DB before retrying; the same bundle is idempotent once state is known.

The image includes `/app/coverctl`. For dev Compose, stop app, then use a one-shot process with
the maintenance entrypoint (not the normal bootstrap, which may run migrations/seed):

```bash
docker compose -f compose.yaml -f compose.dev.yaml stop app
docker compose -f compose.yaml -f compose.dev.yaml run --rm --no-deps \
  --entrypoint /app/coverctl --volume "$PILOT_BUNDLE_DIR:/bundle:ro" app \
  install -writers-stopped -bundle /bundle -manifest-sha256 "$PILOT_MANIFEST_SHA256" -actor-id 1
docker compose -f compose.yaml -f compose.dev.yaml run --rm --no-deps \
  --entrypoint /app/coverctl app verify -writers-stopped
docker compose -f compose.yaml -f compose.dev.yaml up -d app
```

Use the same project options as the retained environment. For stage/prod-like local, select its
matching override and keep PostgreSQL running while app is stopped. Only restart app after every
command succeeds. Do not reapply the development seed to install covers in an existing catalog.
Fresh disposable setup: migrations, matching development seed, explicit administrator provisioning,
then installation of the separately reviewed matching pilot bundle. No default administrator or
unreviewed artwork is provisioned automatically.

## Paired Backup

Stop **all** application/cover/database writers and leave them stopped throughout both snapshots.
Back up SQL and files from the same environment, not just the currently referenced pilot. Preserve
unreferenced retained files as well; cleanup is a different operation. A PostgreSQL dump alone or
files alone are not a recoverable application snapshot. Sessions and private data in SQL need
restricted permissions, encryption/retention appropriate to the environment, and must not enter Git.

Create a fresh private destination and keep its absolute path:

```bash
umask 077
cover_backup_dir=$(mktemp -d)
```

For a stopped **host SQLite** setup, use the SQLite backup API instead of copying a possibly
WAL-backed main file. Replace both absolute source paths with the verified stopped environment:

```bash
sqlite3 -readonly /absolute/path/book_social.db ".backup '$cover_backup_dir/catalog.sqlite'"
tar -C /absolute/path/covers -cf "$cover_backup_dir/covers.tar" .
```

For **dev Compose**, stop app and use a one-shot process attached to its existing volumes:

```bash
docker compose -f compose.yaml -f compose.dev.yaml stop app
docker compose -f compose.yaml -f compose.dev.yaml run --rm --no-deps \
  --entrypoint /bin/sh --volume "$cover_backup_dir:/backup" app -ec \
  'sqlite3 -readonly /app/data/book_social_dev.db ".backup /backup/catalog.sqlite"
   tar -C /app/covers -cf /backup/covers.tar .'
```

For **stage Compose**, leave PostgreSQL running, stop app and all other writers:

```bash
docker compose -f compose.yaml -f compose.stage.yaml stop app
docker compose -f compose.yaml -f compose.stage.yaml exec -T postgres \
  pg_dump -U book_social -d book_social_stage --format=custom > "$cover_backup_dir/catalog.dump"
docker compose -f compose.yaml -f compose.stage.yaml run --rm --no-deps \
  --entrypoint /bin/sh --volume "$cover_backup_dir:/backup" app -ec \
  'tar -C /app/covers -cf /backup/covers.tar .'
```

For prod-like local, select `compose.prod.yaml` and `book_social_prod`; never substitute that target
implicitly. Check every exit status before continuing. A failed dump/redirection is not a backup.
For PostgreSQL use a compatible `pg_dump`/`pg_restore` version (the Compose service is PostgreSQL
18).

Record application/image revision, migration version/dirty state, selected environment/project,
SQL/local-reference counts, pilot bundle ID/checksum, UID/GID and snapshot timestamp. Copy the
reviewed manifest/provenance/source bundle into restricted backup storage or record its separately
verified immutable location. Seal the newly created files:

```bash
(cd "$cover_backup_dir" && sha256sum catalog.* covers.tar > SHA256SUMS)
```

Move/copy this temporary backup to approved persistent backup storage and verify the checksums
there before resuming writers. A temporary directory alone is not a retention policy.

## Restore Into Fresh Targets Only

First verify the trusted snapshot and inspect the archive contents. Do not extract untrusted
archives or overlay SQL/files on a running or retained environment:

```bash
(cd "$cover_backup_dir" && sha256sum -c SHA256SUMS)
tar -tf "$cover_backup_dir/covers.tar"
cover_restore_dir=$(mktemp -d)
mkdir "$cover_restore_dir/covers"
tar --no-same-owner -C "$cover_restore_dir/covers" -xf "$cover_backup_dir/covers.tar"
```

For **SQLite**, copy the consistent snapshot to the new target and verify integrity/references:

```bash
cp "$cover_backup_dir/catalog.sqlite" "$cover_restore_dir/catalog.sqlite"
sqlite3 -readonly "$cover_restore_dir/catalog.sqlite" 'PRAGMA integrity_check; PRAGMA foreign_key_check;'
APP_ENV=dev APP_DB_DSN="$cover_restore_dir/catalog.sqlite" APP_COVERS_DIR="$cover_restore_dir/covers" \
  go run ./cmd/coverctl verify -writers-stopped
```

Integrity must report `ok`, and the foreign-key check must return no rows. Compare SQL counts and
the migration version/dirty state with the snapshot record; `verify` checks local references,
not every database invariant or whether a zero-reference database was the intended backup.

For **PostgreSQL**, explicitly provision a separate empty restore database. Verify its destination
and that its public schema has no tables before setting `COVER_RESTORE_DATABASE_URL`. Never point
this variable at the source or a retained database; no `--clean`, drop, migration or seed is needed:

```bash
pg_restore --exit-on-error --single-transaction --no-owner --no-privileges \
  --dbname="$COVER_RESTORE_DATABASE_URL" "$cover_backup_dir/catalog.dump"
APP_ENV=stage APP_DB_DSN="$COVER_RESTORE_DATABASE_URL" APP_COVERS_DIR="$cover_restore_dir/covers" \
  go run ./cmd/coverctl verify -writers-stopped
```

`verify` reads all books/variants and validates every referenced main/thumbnail's canonical key,
decoded bytes, hash, measured metadata and pair relationship against SQL. Missing/corrupt/mismatched
files fail the command; external HTTPS images are not fetched. A zero count can be valid, so compare
with the recorded expected count. Extra unreferenced files are retained and are not a verify
failure.

Only after these checks, restore/copy into explicitly approved **fresh** persistent DB/cover
volumes,
set permissions for the actual runtime user, and verify once more with the same image before
starting app. Keep the source snapshots and previous environment recoverable until browser review
succeeds. No automatic restore/overwrite/cleanup command is provided.

## Verification and Remaining Acceptance

Automated tests use synthetic five-book bundles, validate refusal/idempotency/audit rollback,
reopen local storage and restore a SQLite snapshot plus cover files into fresh temporary targets.
Opt-in PostgreSQL tests check equivalent installation and verification. `make test/integration`
includes `internal/coverops`; it does not test PostgreSQL dump/restore or container recreation.

Outside the sandbox, in a disposable local environment with approved pilot artwork:

1. Record the bundle/image/migration checksums and initial `coverctl verify` result.
2. Install the pilot twice; expect five changed associations initially (or fewer if already
   present),
   then zero. Verify unrelated metadata/variants and audit count.
3. Restart app, then recreate it with the same project/volumes:
   `docker compose -f compose.yaml -f compose.dev.yaml up -d --force-recreate app`.
   Repeat verification and HTTP/browser checks after each. Repeat for the chosen PostgreSQL
   override.
4. Exercise the paired backup/restore above in fresh targets; compare checksums, all local
   references,
   counts and pilot URLs before starting app against the restored state.
5. Check `/books`, `/books/{pilot-slug}`, each pilot author's page, `/me/library`, and A Christmas
   Carol's home card. Verify HTMX and unavailable-image fallback without deleting retained files.

Record actual commands/results and browser/environment details. Passing unit tests or Compose
configuration validation does not accept restart/redeploy, real PostgreSQL restore, reviewed artwork
or visual presentation; #28 and the expanded release scope remain open.
