# Demo Catalog Cover Dataset v1

This inventory selects one reviewed project demo front illustration for each of
the 109 seeded book slugs. It includes the five user-approved pilot covers and
104 illustrations visually reviewed by Codex at the user's request. They are
not covers of particular published editions. The manifest records the selected
files and SHA-256 hashes; source paths are relative to `data/demo-covers/`.
Generation prompts and review evidence are in the pilot provenance and
`../batch-01/README.md` and `../batch-01/review-notes.md`.

To assemble the 218 delivery images in a new directory from a retained copy:

```bash
python3 assets/demo-covers/catalog-v1/assemble.py data/demo-covers /tmp/book-social-cover-delivery
```

The output directory must not exist. The assembler checks every source hash
before copying and refuses to overwrite an existing output. Restore missing source
files from the separately retained originals before retrying. If a previous run
was interrupted, keep it for inspection and choose a new output directory; do not
replace reviewed originals. The manifest, prompts, review decisions and source
files must be backed up together; `data/demo-covers/` alone is not a backup.

The assembler produces both `inventory.json` and a version 2 installer
`manifest.json`. The pinned SHA-256 of this dataset's installer manifest is:

```text
28329d42d6ad6db7aaffac97cd5f220a7db50f45eebd3133bef9aa1ad293973d
```

Stop all cover writers and use a migrated, seeded database, an existing
administrator ID and an explicit, existing cover directory (the installer never
creates it; create it yourself only for a new environment). Follow the backup
and recovery procedure in [cover operations](../../../docs/cover_operations.md).
For an isolated SQLite environment, substitute your absolute paths and
administrator ID:

```bash
APP_ENV=dev APP_DB_DSN=/absolute/path/catalog.db APP_COVERS_DIR=/absolute/path/covers \
  go run ./cmd/coverctl install -writers-stopped \
  -bundle /tmp/book-social-cover-delivery \
  -manifest-sha256 28329d42d6ad6db7aaffac97cd5f220a7db50f45eebd3133bef9aa1ad293973d \
  -actor-id 1
APP_ENV=dev APP_DB_DSN=/absolute/path/catalog.db APP_COVERS_DIR=/absolute/path/covers \
  go run ./cmd/coverctl verify -writers-stopped
```

The installer checks that the entire catalog matches the 109 unique manifest
slugs before publishing. It validates every file and checksum, then records
front-cover links and audits in one transaction. Repeating an unchanged install
produces zero changes. A failed SQL transaction can leave unreferenced files;
retry the same verified bundle with writers stopped. This does not delete or
replace accepted source artwork. A test using the actual seed on a disposable
SQLite database installed 109 covers, installed zero on repeat, and verified 109
references. The same install, repeat and verification checks passed on a disposable PostgreSQL
database. HTTP and browser evidence and its limits are recorded in the
[full-catalog verification record](../../../docs/cover_catalog_verification_v0_3_2.md).
