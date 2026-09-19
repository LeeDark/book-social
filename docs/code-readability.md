# Go and SQL Readability

This guide defines the project’s readable-default style for new or materially changed Go and SQL.
It supports review; it does not require bulk reformatting of untouched code or historical
migrations.

## How to Use This Guide

The **mandatory rules** below are project conventions. The **review questions** guide judgment when
several readable choices are valid. Preserve a nearby established pattern unless this guide makes a
clearer requirement.

Run `gofmt` for changed Go files and use the existing vet, lint, migration, and test checks that
match the task. Formatting alone does not prove behavior correct.

## Go

### Mandatory Rules

- Use `gofmt`; do not hand-align Go code against the formatter.
- Give packages, types, functions, and variables names based on their role. Use verbs for behavior
  (`Create`, `Find`, `List`) and `is`, `has`, or `can` for boolean values.
- Keep `context.Context` as the first parameter when an operation accepts one. Do not store it in a
  struct.
- Keep handlers responsible for HTTP, services for use-case orchestration, and repositories for SQL.
  Do not hide a layer transition behind a misleading helper name.
- Prefer early returns for invalid input and errors over deeply nested success paths.
- Preserve typed/sentinel errors needed by callers with `errors.Is`; add context only to unexpected
  errors and do not expose their internals to HTTP responses.
- Define a small interface in the package that consumes it. Export a constructor when callers need
  one, but return the concrete type unless a consumer needs an interface.
- Write comments for non-obvious intent, constraints, security decisions, or surprising behavior;
  do not narrate syntax.
- Keep table-driven tests focused on one observable behavior. Name cases by input and expected
  outcome rather than implementation detail.

### Visual Structure

- Separate distinct logical phases with a blank line; keep closely related statements together.
- Prefer a top-to-bottom sequence of domain or application steps over visually dense blocks or
  deeply nested control flow.
- Break a long composite literal or call across lines when it makes fields, arguments, or roles
  easier to scan. Do not wrap code mechanically when the one-line form is clearer.
- For orchestration functions such as `main`, make phases visible in order: configuration, loading,
  dependency initialization, transport setup, and execution.
- Optimize for a human reviewer reading the whole file, not for the fewest lines or the shortest
  function.

```go
func (s *CatalogService) BookDetailsPage(ctx context.Context, slug string) (BookDetailsPageData, error) {
	book, err := s.repo.GetBookBySlug(ctx, slug)
	if err != nil {
		return BookDetailsPageData{}, err
	}

	return mapBookDetailsPage(book), nil
}
```

The early return makes the successful path visible without another nesting level.

### Review Questions

- Can a reader understand the normal path without following several helpers?
- Is a new abstraction clearer than the repeated code it replaces?
- Does an error retain the information its caller needs while keeping unsafe detail out of responses
  and diagnostics?
- Does a test describe the user-visible contract rather than mirror private implementation steps?
- Are phase boundaries, long expressions, and composite literals easy to scan without mentally
  reformatting the code?

There is no line-count limit. Split a function when its responsibilities or control flow become
hard to explain, not merely to satisfy an arbitrary size target.

## SQL

### Mandatory Rules

- Write SQL keywords in uppercase and use one clause per line for multi-clause statements.
- Name columns explicitly in `SELECT` and `INSERT`; do not use `SELECT *`.
- Keep an `INSERT` column list and its `VALUES` list in the same order.
- Qualify selected or filtered columns when a query joins tables or could become ambiguous. Use short,
  stable aliases only when they improve readability.
- Put `JOIN`, `WHERE`, `ORDER BY`, and `LIMIT` clauses on separate lines. Put each independent
  predicate on its own line when that makes a compound condition easier to scan.
- Use the placeholder convention of the target storage package: `?` for SQLite and numbered `$N`
  placeholders for PostgreSQL. Do not make a query look cross-dialect when it is not.
- Name durable constraints and indexes with the existing prefixes: `uq_`, `fk_`, `ck_`, and `idx_`.
- Keep SQLite and PostgreSQL migration versions paired. Make intentional dialect differences
  explicit in the SQL or migration documentation.
- Comment only the reason for a non-obvious migration or database decision.

```sql
SELECT id, first_name, login, email, user_role_id
FROM users
WHERE login = ? OR email = ?
LIMIT 1;
```

```sql
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,

    CONSTRAINT fk_sessions_user
        FOREIGN KEY (user_id) REFERENCES users(id)
            ON UPDATE CASCADE
            ON DELETE CASCADE
);
```

The first query exposes the repository result shape. The second keeps a named constraint and its
referential action together.

### Review Questions

- Is the result shape, mutation target, and ordering obvious from the query?
- Are joins and predicates readable without relying on implicit column resolution?
- Does the migration preserve the documented SQLite/PostgreSQL application semantics?
- Does a comment explain a dialect or migration constraint that a future maintainer would otherwise
  miss?

## Readability-Only Changes

A readability-only change must not silently change observable behavior, public APIs, state
transitions, validation, error outcomes, migration semantics, or unresolved domain decisions. When
a review reveals an ambiguity, record it as a finding or create a scoped follow-up instead of
deciding it under the cover of cleanup.

## Review Checklist

Before approving a Go or SQL change, verify that it follows the mandatory rules, uses existing
project checks, and does not mix an unrelated readability cleanup into a behavior change. Record a
deliberate exception in the PR description when a nearby established pattern must be retained.
