# Private Library v0.3 Contract

This document records the implemented v0.3.0 private-library behavior and the accepted v0.3.1
lifecycle contract. Current routes, domain behavior, and schema are also summarized in
[routes.md](routes.md), [domain.md](domain.md), and [database.md](database.md).

## Accepted Baseline

v0.3.0 builds on the closed v0.2.6 authentication flow:

- the current user is loaded from an opaque DB-backed session;
- protected routes use the existing authentication guard;
- handlers receive only the typed current-user identity, never credentials or session secrets;
- unsafe cross-origin browser requests are rejected by `http.CrossOriginProtection`;
- dynamic user-specific responses use `Cache-Control: no-store`.

The legacy `library`, `shelves`, and `tags` tables are demo data and are not part of this contract.

## v0.3.0 Use Cases

### Add a catalog book

An authenticated user adds one existing catalog book to their private library. The request uses the
book slug as the public identifier; the handler always takes the owner ID from the authenticated
request context. A client-supplied user ID is never accepted.

The service:

1. validates the user ID and normalized book slug;
2. resolves the book through the catalog boundary;
3. records one library item with the service clock's current UTC time;
4. returns an explicit conflict when the same user already owns an item for that book.

The unique `user_id + book_id` database constraint is authoritative. Repeated submissions are not
treated as successful idempotent writes.

### List the current user's library

An authenticated user lists only their own items. The repository query is always scoped by user ID
and returns enough catalog data to build the existing server-rendered book presentation without
exposing the stored owner ID to templates.

Items are ordered by `added_at DESC`, then library-item ID descending for deterministic ties. An
empty library is a successful result, not a not-found error.

## Module Boundary

The `library` module owns its domain model, service, errors, and repository contract. It may depend
on the existing `books` module as the upstream catalog boundary; the `books` module must not depend
on `library`.

The consuming service defines these minimum ports and values:

```go
type BookFinder interface {
	GetBookBySlug(ctx context.Context, slug string) (books.Book, error)
}

type Repository interface {
	Add(ctx context.Context, params AddItemParams) error
	ListByUserID(ctx context.Context, userID int) ([]Item, error)
}

type AddItemParams struct {
	UserID  int
	BookID  int
	AddedAt time.Time
}

type Item struct {
	ID      int
	Book    books.Book
	AddedAt time.Time
}
```

The service exposes add and list use cases. It owns input normalization, catalog-error translation,
the operation clock, and orchestration. Repositories own SQL and database-error translation.
Handlers own HTTP parsing, redirects, status codes, flash messages, and page/view models.

No service-level transaction is required in v0.3.0: add performs one authoritative insert after a
catalog read, and list is read-only. Foreign keys handle a catalog row disappearing between lookup
and insert. If a later use case needs multiple writes, the service defines the transaction boundary
and receives a transaction-capable repository abstraction; handlers never manage transactions.

## Error Contract

The library module uses errors compatible with `errors.Is`:

| Application outcome                   | Library error          | HTTP behavior                       |
|---------------------------------------|------------------------|-------------------------------------|
| Missing/invalid user ID or blank slug | `ErrInvalidInput`      | `422 Unprocessable Entity`          |
| Catalog slug does not identify a book | `ErrBookNotFound`      | `404 Not Found`                     |
| User already owns the book            | `ErrItemAlreadyExists` | `409 Conflict`                      |
| Requested library item does not exist | `ErrItemNotFound`      | `404 Not Found`                     |
| Authenticated user is not permitted   | `ErrForbidden`         | `403 Forbidden`                     |
| Unexpected repository/catalog failure | `ErrInternal`          | generic `500 Internal Server Error` |

Missing or invalid authentication remains `users.ErrUnauthenticated` at the existing HTTP auth
boundary. Protected MPA routes redirect anonymous users to `/login` with `303 See Other`; the
library service does not accept anonymous calls as a normal use case.

Catalog `books.ErrBookNotFound` is translated to `library.ErrBookNotFound`. Database uniqueness is
translated to `ErrItemAlreadyExists`. Unexpected details are wrapped for server logs but never sent
to the client.

## Implemented HTTP Contract for v0.3.0

```text
GET  /me/library   list the authenticated user's private library
POST /me/library   add the book identified by form field book_slug
```

- Both routes use the existing authentication guard and no-store response policy.
- A successful add uses Post/Redirect/Get with `303 See Other` to `/me/library` and a success flash.
- Invalid input returns `422`; unknown books return `404`; duplicates return `409`.
- Forms remain ordinary MPA forms. Existing cross-origin protection is retained without adding a
  separate CSRF-token mechanism.
- Catalog and book-detail pages render the same add form only for authenticated users; neither can
  select an owner. The library page presents every v0.3.0 item as `Want to read` and includes an
  accessible empty state with a Browse catalog action.

There is no route for another user's library. Future item mutation routes must include the current
user ID in every repository lookup or mutation instead of loading an item globally and checking it
afterward.

## Ownership and Privacy Rules

- Every item has exactly one user owner and one catalog book.
- Only the authenticated owner may list or mutate an item.
- Repository methods that read or mutate a specific item must be owner-scoped.
- User IDs, role internals, passwords, session tokens, and token hashes do not enter page models,
  form fields, flash messages, logs, or test diagnostics.
- Cross-user probes must not reveal private item existence. Owner-scoped missing items use
  `ErrItemNotFound`; `ErrForbidden` is reserved for operations where policy is checked before a
  private item lookup.

## Reading-State Rules for v0.3.1

v0.3.0 has no status mutation. Its items are presented as want-to-read entries, and its minimal
schema contains ownership, book identity, uniqueness, and `added_at`. v0.3.1 will persist the
explicit statuses `want_to_read`, `reading`, and `read` and add `started_at` and `finished_at`.

All three status values may transition to either of the other values. Repeating the current status
is an idempotent no-op. Timestamp invariants are:

| Target status  | `started_at`                                                 | `finished_at`          |
|----------------|--------------------------------------------------------------|------------------------|
| `want_to_read` | cleared                                                      | cleared                |
| `reading`      | preserve an existing value, otherwise set to transition time | cleared                |
| `read`         | preserve an existing value, including `nil`                  | set to transition time |

This permits marking a book as read when its start date is unknown. Moving from `read` back to
`reading` preserves a known start and clears the finish. v0.3.1 will define the conflict-safe update
mechanism, status forms, explicit-confirmation removal, and their HTTP routes before implementation.

## Deferred Work

The following are not part of v0.3.0: persisted status changes, removal, ratings, notes, reading
progress, custom shelves/tags, public libraries, external catalog import, and social features.
