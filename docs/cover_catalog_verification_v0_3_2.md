# Full-Catalog Demo Cover Verification (#29)

This record covers the 109-book demo artwork dataset, not the earlier #28 pilot acceptance.
The [inventory and assembly procedure](../assets/demo-covers/catalog-v1/README.md) identify
selected artwork and the pinned installer manifest. Five pilot illustrations were approved by
the user; Codex reviewed the remaining 104 at the user's request. The artwork is project demo
illustration, not a representation of specific published editions. Review of artwork does not
verify legacy book descriptions or other metadata.

## Installation and inventory

On 2026-10-31, a fresh SQLite database with migrations and the actual development seed contained
109 books. The assembled dataset was checked against the seed: 109 unique slugs, 218 existing
WebP files, matching SHA-256 hashes, and the documented 600×900 / 240×360 dimensions.
`coverctl install` changed 109 front covers, a repeat changed zero, and `coverctl verify`
verified 109 local references. Focused Go tests and `GOCACHE=/tmp/book-social-go-cache go test ./...`
passed; the full-catalog SQLite test also rejects a changed catalog inventory.

On 2026-10-31, a separate disposable PostgreSQL 18 container received migrations and the real
seed. The same install, repeat and verify operations returned 109 changes, zero changes and 109
verified references. SQL reported 109 front rows for 109 distinct books. The container and
temporary PostgreSQL application server were stopped after verification.

## HTTP and browser evidence

On SQLite, automated HTTP checks requested all 109 detail pages and both image variants per
book, confirming their response hashes and WebP content type. Headless Chrome rendered the
catalog at 1440×1000 and a book page at 390×844. A browser-driven HTMX author filter changed
the 109-card list to five cards, updated the URL and loaded all five images. An isolated test
account added three books to the private library; their selected thumbnails were present and
loaded in the browser. The user was shown the temporary SQLite catalog and reacted positively;
that reaction is not recorded as explicit acceptance of issue #29 or the release.

On PostgreSQL, HTTP checks covered the same 109 detail pages and 218 image responses; the
selected files and hashes matched. A filtered HTMX partial returned five cover references and
headless Chrome rendered the desktop catalog. This PostgreSQL pass did not repeat authenticated
library interaction, mobile browser review, backup/restore or container recreation for the full
109-cover dataset. The earlier pilot recovery and lifecycle evidence remains in its own records.

## Remaining boundaries

The image files are ignored by Git and currently retained locally; an independent durable backup
location has not been established. This record does not close issue #29, accept the entire v0.3.2
release, or claim deployment. Any final release decision should consider these boundaries.
