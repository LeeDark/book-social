# Cover File Storage and Five-Book Pilot (#28)

Status: implementation contract prepared on 2026-10-25. Storage, uploads, image delivery and
the pilot are not implemented or accepted. This document extends the original
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

## Prepared Images and Limits

The proposed pilot delivery profile is static WebP, 2:3, 600 × 900 main image and 240 × 360
thumbnail. Review legibility, detail, proportions and file weight at mobile and desktop sizes
before freezing that profile or generating the remaining 104 covers. No format or visual
acceptance is claimed by this document.

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
Extend `covers` with an optional local storage key and the derivative metadata needed for an
optional thumbnail; choose the minimal paired SQLite/PostgreSQL migration during implementation.
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

## Attachment, Replacement and Failure Rules

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

## Delivery and Acceptance Gates

Later steps must provide an idempotent pilot installer that validates the entire bundle and
resolves slugs before attachment. It must preserve unrelated catalog data and cover variants, and
must not rerun the development seed to update an existing database. A fresh disposable setup
applies migrations and seed, then installs the matching pilot files and references.

Document backup and restore of both SQL and the cover volume with cover writers stopped, bundle
version/checksum verification, and post-restore reference checks. The five books must work on
details, catalog and their author pages and when added to the private library. Check the home
card for A Christmas Carol using the current seeded selection; no home-selection change is needed.

Acceptance requires focused storage/service/HTTP tests, equivalent SQLite/PostgreSQL semantics,
missing/unavailable-image fallback including HTMX, and a manual visual review of the proposed
delivery profile. Restart, container recreation/redeployment, fresh installation and restoration
must be verified locally outside the Codex sandbox. Record actual commands and environment as
evidence. Completing this contract does not accept those later gates or close #28 or v0.3.2.
