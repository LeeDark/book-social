# PostgreSQL Container Lifecycle and Browser Checks (#28)

On 2026-10-30, with explicit user authorization, exercised the isolated restored PostgreSQL
`cover_restore` and restored SQLite data described in [recovery](cover_recovery_v0_3_2.md).
Application image `sha256:f347346b046fdb8f6460afa0397b7692efac48981755eb671ca38f27c06de561`;
working revision `5e43e56`. No retained dev/stage/prod environment was accessed.

## PostgreSQL Lifecycle

Project `book-social-cover-step5-lhkev5` used PostgreSQL 18 and the app configured with
`APP_ENV=stage` and an explicit DSN for the recovered `cover_restore` database. Created a new
test-only Docker volume `book-social-cover-step5-lhkev5_pg-covers`, checked it did not exist,
then copied restored cover files into the empty volume. The source restored directory remained
intact. A temporary Compose override mounted the volume read-only at `/covers`. The DB remained
in the project's separate `postgres-data` named volume. Compose warned the manually created
cover volume was not created by Compose; the exact named volume was nevertheless mounted and
confirmed with `docker inspect`.

Actual Compose invocation from the repository root:

```bash
pc() {
  docker compose -p book-social-cover-step5-lhkev5 \
    -f /tmp/book-social-cover-recovery.LhkEv5/compose.yaml \
    -f /tmp/book-social-cover-recovery.LhkEv5/compose.http.yaml \
    -f /tmp/book-social-cover-recovery.LhkEv5/compose.pg-lifecycle.yaml "$@"
}
```

`pc restart pg-app` kept container ID `2ee5e8061fda` while changing its StartedAt from
`2026-10-29T23:58:24.845428612Z` to `2026-10-29T23:58:42.235530243Z`. Stopped the app and
ran `/app/coverctl verify -writers-stopped` with the explicit PostgreSQL DSN and the same
read-only named cover volume: five references verified. Then
`pc up -d --no-build --no-deps --force-recreate pg-app` created ID `5f71154d4ef4`, retaining
the same named cover volume and DB. Repeated verify after stopping the recreated app: five.

Before restart, after restart and after recreation, the step-4 HTTP probe passed all five detail
pages, ten WebP status/MIME/SHA-256 checks, catalog, home, five author pages and HTMX catalog
fragment. Source and restored PostgreSQL full `pg_dump --inserts --no-owner --no-privileges`
were compared after the run, excluding random restrict tokens: the only differences were the
two equivalent CHECK cast representations already documented in recovery. A preliminary
raw dump hash comparison varied because PostgreSQL generates random `\restrict` tokens; it is
not evidence of data modification. No migrations, seed or reinstallation were run.

## Desktop, Accessibility and Fallback

Chrome `155.0.8059.79`, headless at 1440×1000, rendered Pride and Prejudice's details from both
recovered SQLite and the recreated PostgreSQL app. Inspected screenshots: image proportions,
title and desktop layout were correct. Screenshots are temporary files under
`/tmp/book-social-cover-recovery.LhkEv5/`, not tracked evidence artifacts. Chrome emitted
Crashpad lock/inotify warnings but returned success and rendered the screenshots.

Chrome DevTools Protocol checked, on both engines for that page, a loaded 600px main image,
decorative `alt=""` and `aria-hidden="true"`, the accessible heading and Back to catalog link,
and that the first keyboard Tab reaches a named link with a solid outline. This is a focused
check, not a complete keyboard, screen-reader, contrast or WCAG audit.

Created an independent SQLite backup copy as `/tmp/book-social-cover-recovery.LhkEv5/fallback-fixture`
and changed only its test rows. The restored SQLite and PostgreSQL databases were not modified:

- Removed Christmas's thumbnail metadata in the copy. Browser confirmed the author-filtered
  catalog card uses the 600px main image.
- Set Pride's cover URL to `https://example.com/cover-test.webp` in the copy. With a controlled
  DevTools response containing a valid WebP, the browser displayed it. With the request blocked,
  the image was hidden and the Pride placeholder appeared. This verifies HTTPS URL rendering and
  failure behavior, not availability/CORS of an actual external provider.
- Disabled JavaScript in DevTools and opened a book without cover metadata. The browser showed
  its title placeholder and no image. Inspected a screenshot of this state.

The earlier user-confirmed Zen mobile/HTMX blocking checks remain separate. No signed-in library
flow or every pilot page was visually retested in these desktop screenshots. Browser scripts,
SQL fixture and Compose overrides remain in the private temporary test directory; `/tmp` is not
durable storage. All test app and PostgreSQL containers were stopped cleanly, and test volumes,
databases, image and screenshots were retained. No deletion, production deployment or closure.

Validation included real HTTP/browser checks, stopped-writer verify, SQL dump comparison,
SQLite fixture integrity and `git diff --check`. No application code changed and the full suite
was not rerun for this verification/documentation pass.
