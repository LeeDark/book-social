# Cover File Storage and Five-Book Pilot (#28)

Status: implementation contract prepared on 2026-10-25. Steps 2–6 implement local file storage,
validation, public HTTP delivery, database references/metadata, administrator uploads and shared
cover rendering/fallback, persistent Compose volume configuration and offline delivery/verification
tools. The five-artwork pilot and profile are user-approved and passed isolated SQLite installation;
the user confirmed mobile pilot display, library, home, HTMX and blocked-image fallback in Zen.
Actual [SQLite restart/recreation](cover_container_lifecycle_v0_3_2.md) passed on 2026-10-30.
Paired [offline SQLite/PostgreSQL recovery](cover_recovery_v0_3_2.md) passed on 2026-10-30.
Both restored states also passed real HTTP checks with unchanged SQL and repeated offline verify.
PostgreSQL restart/recreation, restored-state visual review and remaining detailed browser checks remain pending.
Step 7's [automated evidence and outstanding manual gates](cover_acceptance_v0_3_2.md) are recorded
separately; the final automated pass includes real SQLite/PostgreSQL cover workflows.
This document extends the original
[URL-only catalog operations contract](catalog_operations_v0_3_2.md); it does not change the
recorded acceptance of issues #23–#26. Release scope remains in [roadmap.md](roadmap.md).

## Scope and Ownership

Store prepared front-cover files outside SQL, attach them through the administrator workflow,
and display them on book details and all card contexts: catalog, home, author and private library.
Keep external HTTPS covers supported and preserve other cover variants. Missing or unavailable
images must show the existing placeholder, including after an HTMX catalog update.

The books service owns validation and attachment/replacement decisions; repositories own SQL and
the existing catalog audit transaction. A small local storage component owns safe file operations.
HTTP handlers parse bounded requests and serve validated files. This work adds no general media
platform, external-URL downloads, image editor, runtime resizing or background conversion jobs.

## Shared Pilot

Use these five stable slugs in #28 and the companion demo-artwork task. All exist in both SQL seeds;
do not select them by numeric ID or rerun seed against an existing catalog.

| Book                                   | Slug                                     | Visual review purpose                                            |
|----------------------------------------|------------------------------------------|------------------------------------------------------------------|
| A Christmas Carol                      | `a-christmas-carol`                      | Seasonal illustration; also exercises the current home selection |
| Pride and Prejudice                    | `pride-and-prejudice`                    | Romance and restrained composition                               |
| Frankenstein                           | `frankenstein`                           | Dark artwork and contrast                                        |
| Alice's Adventures in Wonderland       | `alices-adventures-in-wonderland`        | Detailed fantasy artwork and a longer title                      |
| Twenty Thousand Leagues Under the Seas | `twenty-thousand-leagues-under-the-seas` | Adventure artwork and a long title                               |

The selection is the working pilot, not approval of artwork or legacy metadata. Source files,
generation details, reference sources and permitted-use basis, review status and delivered file
checksums belong in the companion task's manifest, keyed by slug. Generated artwork must be
identified as project demo illustrations, not covers of particular editions. Existing metadata
provenance limitations in [demo_catalog_provenance.md](demo_catalog_provenance.md) remain.

The user approved all five illustrations and their delivery profile on 2026-10-25. The separate
[`pilot-v1-approved` release](../assets/demo-covers/pilot-v1-approved/README.ru.md) retains original
PNGs, prepared WebPs, prompts, provenance, review decision and pinned checksums. The
[installation protocol](cover_pilot_installation_v0_3_2.md) records actual offline SQLite installation
twice (five changes then zero), five audits and verification. The historical
[`pilot-v1` draft](../assets/demo-covers/pilot-v1/README.ru.md) remains unchanged and un-installable.
Artwork approval alone does not accept application rendering or authorize the remaining 104
illustrations. Subsequent user-reported browser evidence is recorded in the
[acceptance checklist](cover_acceptance_v0_3_2.md#user-reported-browser-evidence).

## Prepared Images and Limits

The user-approved pilot delivery profile is static WebP, 2:3, 600 × 900 main image and 240 × 360
thumbnail. Review legibility, detail, proportions and file weight at mobile and desktop sizes
before expanding that profile to the remaining 104 covers. The user confirmed mobile pilot
presentation in Zen; desktop and other detailed checks remain in the acceptance checklist.

The initial validation contract is:

- Accept static JPEG, PNG and WebP only, with a decoder capable of validating each enabled format.
  WebP support requires a small reviewed decoding dependency; merely trusting its header is
  insufficient. Dependency selection belongs to implementation, and must preserve the current
  container build. No encoder is required because files arrive prepared.
- Reject animation, SVG, unsupported formats, corrupt or truncated data, zero dimensions, and
  mismatches between the declared format/extension and the decoded format. Derive the stored
  extension and MIME type from validated content; never trust the submitted MIME type alone.
- Limit each file to 5 MiB (5,242,880 bytes), each dimension to 4,096 pixels, and total pixels to
  12,000,000. Check dimensions before full decode, then validate the image by decoding it.
- Limit the whole multipart request to 12 MiB (12,582,912 bytes), including form overhead, and
  accept at most one main file and one optional thumbnail. Clean up multipart temporary files.
- Main images must be portrait. A thumbnail must use the same format and aspect ratio as its
  main image, with smaller dimensions. Exact pilot dimensions are a delivery-profile decision,
  not a universal constraint on externally hosted covers.
- Calculate MIME type, byte size, dimensions and SHA-256 from the accepted bytes separately for
  each delivered file. Do not silently transform files or synthesize missing metadata.

A thumbnail is an optional derivative of `front`, not another semantic variant such as `back`.
Cards use it when supplied and otherwise use the main file; details use the main file. Supplying
two files requires visual review that they depict the same artwork. The browser layout preserves
proportions without stretching. Source artwork is retained in the delivery bundle, outside the
public serving directory.

## References, Names and Persistence

Use `APP_COVERS_DIR` for the storage root and a persistent Docker volume for container runs.
Store only relative storage keys and public URLs in SQL, never host paths or image bytes.
Paired SQLite/PostgreSQL migration `000008` extends `covers` with an optional local storage key
and the derivative metadata needed for an optional thumbnail.
Existing main-image metadata columns remain authoritative for the main file.

An external cover has an HTTPS URL and no local storage key. A local cover has a validated storage
key and an application-relative URL under `/media/covers/`. URLs and keys for local files are
server-generated, not free-form URL input. A front cover has one active source at a time; external
URLs are never fetched or copied by the server.

Generate immutable flat keys as `<lowercase-sha256>.<jpg|png|webp>` from validated file bytes.
Do not use submitted filenames or book slugs as paths. Reject arbitrary paths, separators,
traversal and symlink escapes in storage and serving operations. Create temporary files inside
the same storage filesystem, then publish complete files atomically without overwriting another
file. Reusing an existing key requires verifying that its bytes match the expected checksum.

Serve only keys allowed by this contract, without directory listings or access to temporary,
source or manifest files. Use the verified MIME type and the existing security-header policy.
Immutable URLs may be cached; a replacement receives a new URL when its bytes change. Cover
assets are public even when displayed in a private library; serving must expose no membership
or account data. Volume configuration and permissions must survive container recreation.

### Implemented Storage Boundary (Step 2)

`internal/storage/covers` creates/opens the configured root and exposes `Put`, `Read` and `Close`.
`Put` accepts a context, input stream, original filename and optional declared MIME type. It
validates prepared portrait images and returns the generated key, public URL and measured
metadata. The original filename is only an extension hint, never a destination path. An omitted
MIME type is detected; a supplied mismatch is rejected. PNG animation chunks and WebP animation
flags/chunks are rejected explicitly in addition to full decoding. The WebP decoder is
`golang.org/x/image/webp` v0.47.0, a Go implementation requiring no CGO; its module also upgrades
the existing indirect `golang.org/x/sys` requirement to v0.49.0.

Files are synced before publication by a no-overwrite hard link inside an `os.Root`; the directory
is synced before success. The storage filesystem must support hard links and directory syncing.
Existing content must match byte-for-byte before a key can be reused. Regular files only are
read; symlinks and keys outside the flat naming contract are rejected. A crash can leave an
unreferenced `.upload-*` temporary file, which is never publicly served. No deletion API is exposed.
The configured directory and its OS permissions are controlled by the environment owner;
the volume must not be writable by untrusted local processes.

The process opens the store at startup and closes it on shutdown. `APP_COVERS_DIR` defaults to
`./data/covers`; an explicitly blank value is a configuration error. The default directory is
ignored by Git. Compose sets a separate persistent `/app/covers` volume per environment; see
[delivery and recovery](cover_operations.md) before upgrading a retained container.

`GET` and `HEAD /media/covers/{key}` work without session or flash loading. The handler verifies
the file's checksum and fully validates its image content on each request before serving; this
bounded initial implementation trades decode work for detecting damaged or incorrectly supplied
files. It returns the image MIME type, SHA-256 ETag and a one-year immutable public cache policy.
Conditional requests and byte ranges are supported. Missing, invalid or damaged files return
`404`; unexpected storage failures return a generic `500` and server-side diagnostics. Error
responses use `no-store`; other methods return `405`. All routes retain global security headers.

## Attachment, Replacement and Failure Rules

Step 3 provides `StoredCoverFile`/`LocalFrontCover` and transactional `SetLocalFrontCover` on
both repositories, with complete local/thumbnail metadata returned by operator and detail reads.
Replacing an external URL clears local fields; unchanged URLs preserve them. See
[database semantics and guarded downgrade](database.md#local-cover-references-28-step-3).
Step 4 wires file verification/publication and administrator authorization into the browser
form/service. Direct repository callers still own those prerequisites and the audit transaction.

An administrator explicitly chooses keep, upload/replace, set external HTTPS URL, or remove.
An ordinary metadata edit defaults to keep. Reject contradictory actions and files before writing.
Replacing a main file without a new thumbnail clears the old thumbnail association, so it cannot
display previous artwork. Setting an external URL clears local keys and derivative metadata.
Removing an association deletes only `front`; `back` and other variants remain untouched.

Validate files and publish them durably before the SQL transaction references them. Save the
front reference, verified metadata and existing catalog audit entry in one transaction. A
validation or storage failure changes neither the database nor the old file. A confirmed SQL
rollback leaves the previous association intact; a failed operation has no success flash.

SQL and the filesystem have no shared transaction. A crash after file publication can leave an
unreferenced file. If the commit result is uncertain, retain published files until the database
state is checked. Never delete a possibly committed or shared file as automatic compensation.
Concurrent replacements use immutable files and the existing catalog write policy; this step
does not introduce optimistic locking for the editor.

Replaced and detached files remain on disk initially. Cleanup is a separate operator maintenance
procedure with a reference check, a dry-run inventory and stopped cover writers; it must consider
both main and thumbnail references across all books. No automatic garbage collection or physical
deletion is part of the user-facing remove action.

### Operator Workflow (Step 4)

Apply migrations through `000008`, configure `APP_COVERS_DIR` with a writable local directory and
use an appointed administrator. Open `/admin/catalog` to find a book and its edit page, or
`/admin/catalog/books/new` to create one. The existing authorization, no-store, same-origin POST
protection and one-request success flashes apply to both URL-encoded and multipart requests.

The multipart form offers these actions without JavaScript:

| Action | Required input | Result |
|--------|----------------|--------|
| Keep current cover (default) | No URL or files | Save book text/relationships without changing the cover |
| Upload or replace local cover | Main image; optional thumbnail | Verify both images and their relationship, publish files, then attach metadata |
| Set external HTTPS URL | HTTPS URL; no files | Save the URL and clear old local/derivative metadata without fetching it |
| Remove front cover | No URL or files | Remove only the `front` database association; retain other variants and files |

The edit form links to the current image and leaves replacement URL/file fields empty. An omitted
action means keep; a blank URL no longer implicitly removes the cover. Old forms supplying a URL
without an explicit external action must be reopened. Unknown actions, duplicate actions/files,
unknown file fields, thumbnail-only uploads and contradictory URL/file input return `422`.

The whole multipart body is limited to 12 MiB, text to 128 KiB, each image to 5 MiB; the format and
dimension limits above also apply. File names and MIME hints must match decoded content. Both files
and thumbnail proportions/format are checked before publishing either file. File streams are
closed and multipart temporary files removed on success and failure. Text/selections/action survive
validation errors; browsers cannot refill file inputs, so select files again before retrying.

The service checks administrator capability and the target book before publication outside the
SQL transaction, then checks them again inside the write transaction. A no-op upload still adds
one audit entry with an empty changed-group summary. Replacing/removing just the thumbnail at the
same main URL is a `front_cover` change. Input file names, image bytes and URL/text values are not
copied into audit summaries or request logs. SQL commit/audit failure retains published
files, including on uncertain commit results; no automatic deletion or retry is attempted.

For later cleanup, stop all cover writers and resolve uncertain commits against the actual DB.
Preserve a matching SQL/file snapshot using the [paired backup procedure](cover_operations.md#paired-backup).
Export every referenced local key, including thumbnails and all books:

```sql
SELECT storage_key
FROM covers
WHERE storage_key IS NOT NULL
UNION
SELECT thumbnail_storage_key
FROM covers
WHERE thumbnail_storage_key IS NOT NULL;
```

Compare this set with an inventory of regular files directly under the storage root matching the
canonical key pattern. Produce and review a dry-run list of unreferenced files; missing referenced
files require restoration, not cleanup. Leave unknown names, directories and `.upload-*` remnants
for separate inspection. While writers remain stopped, recheck each candidate against the latest
references and move only explicitly reviewed files to quarantine outside the public serving root.
Retain inventory/checksums so moves can be reversed, never overwrite quarantine files, and do not
use wildcard/recursive deletion. Verify all SQL references before restarting writers. Permanent
deletion needs a separate operator decision. This maintenance procedure is documented but not yet
exercised as operational acceptance; saving a book never invokes it.

Automated SQLite/PostgreSQL HTTP checks cover upload/create, replacement, keep, URL switch,
removal, refusals, cleanup, storage failure and audit rollback. Manual browser acceptance, the five
approved artworks and container lifecycle/restore acceptance
remain later work. For a local manual check, use the create/edit routes above and the saved book's
`/books/{slug}` detail page and the card contexts below.

## Shared Rendering and Fallback (Step 5)

Catalog (including filtered HTMX fragments), home and author cards now receive cover metadata from
one batch read per book list. Owner-scoped library lists reuse that batch read for their selected
books. No per-card SQL, filesystem checks or external fetches are performed while rendering HTML.
Only `front` is displayed: cards prefer its supplied thumbnail, otherwise its main image; details
always use the main image. Other variants remain in the database but are not shown as front artwork.
External HTTPS main URLs remain supported. No schema change or dependency was added.

All contexts use `book_cover.tmpl`, retaining the title-based CSS placeholder underneath the image.
The stable 2:3 frame, optional measured width/height attributes and `object-fit: contain` preserve
layout and image proportions without stretching/cropping; images use lazy loading and async decode.
Cover links retain a title-based accessible label, and the detail frame has an accessible image role.
The artwork is decorative inside those labelled containers to avoid duplicate announcements.

The external `cover-images.js` script hides failed images to reveal the existing placeholder,
without inline handlers or changes to CSP. It captures non-bubbling image errors for both initial
and newly inserted images and checks already-completed failures at DOM readiness and `htmx:load`.
This unavailable-image fallback requires JavaScript; absent metadata renders the placeholder with
or without JavaScript. A failed thumbnail goes directly to the placeholder, not a second image
request; a missing thumbnail association uses the main URL from the start.

SQLite HTTP tests cover each context with a local pair, main-only and external covers, an absent
cover and missing main/thumbnail files (`404` assets with usable `200` HTML). The same scenarios are
available as opt-in PostgreSQL tests. JavaScript DOM doubles cover initial/cached failures, dynamic
errors and HTMX insertion; they do not establish browser layout or visual acceptance.

For manual verification outside the sandbox, attach permitted prepared images to one known book,
then open `/books`, its filtered results, `/books/{slug}`, `/authors/{author-slug}` and `/me/library`
after adding it. Open `/` when it is in the deterministic featured selection (the pilot includes
A Christmas Carol). Check mobile/desktop proportions, keyboard links, missing-file and blocked
external-image fallback, and filtering without a full reload. Do not remove production files to
test failure; use a disposable local setup or browser request blocking. Pilot assets/profile review
and operational acceptance remain unclaimed.

## Delivery and Acceptance Gates

Step 6 provides `cmd/coverctl` and `internal/coverops`: an explicit idempotent five-book installer
with pinned manifest/file checksums and whole-bundle/slug validation before publication, one
reference/audit transaction, and a post-restore verifier for every local reference. No artwork is
shipped or accepted by this tool. It preserves unrelated catalog data and cover variants and never
reruns migrations/seed. A fresh disposable setup applies migrations and seed separately, then
installs matching reviewed pilot files/references with an existing administrator for the audit.
See [the operator procedure](cover_operations.md) for manifest format, commands, old-storage
transition warnings, per-environment volumes, paired backup/restore and pending acceptance gates.

Document backup and restore of both SQL and the cover volume with cover writers stopped, bundle
version/checksum verification, and post-restore reference checks. The five books must work on
details, catalog and their author pages and when added to the private library. Check the home
card for A Christmas Carol using the current seeded selection; no home-selection change is needed.

Acceptance requires focused storage/service/HTTP tests, equivalent SQLite/PostgreSQL semantics,
missing/unavailable-image fallback including HTMX, and a manual visual review of the proposed
delivery profile. Restart, container recreation/redeployment, fresh installation and restoration
must be verified locally outside the Codex sandbox. Record actual commands and environment as
evidence. Completing this contract does not accept those later gates or close #28 or v0.3.2.
