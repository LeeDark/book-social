# Technical Backlog

This document collects technical ideas, experiments, cleanup opportunities, and possible future
tasks that are not committed release work.

`docs/roadmap.md` remains the source of truth for the current priority, release sequence, and
accepted scope. An item in this backlog is not a promise and should not interrupt the active
roadmap stage. Review backlog items when planning a release, after learning from an experiment, or
when their revisit conditions become true. Promote an item to the roadmap only with a concrete
user or operator outcome, scope, verification plan, and definition of done.

## Candidate format

Each item should record:

- status and current decision;
- why it is not roadmap work now;
- evidence or historical notes;
- concrete conditions that justify reconsidering it.

Remove resolved items or keep their final decision in the relevant architecture/task-history
document. Do not use this file as an unbounded duplicate of the roadmap.

## Current planning boundary

Active roadmap work is v0.2.6 Registration/Login/Logout. No backlog item is promoted by this
document; defer unrelated API, rendering, and infrastructure ideas until the active release is
closed or a concrete revisit condition is met.

## Public catalog JSON API

Status: deferred; no active implementation.

Current decision: the catalog remains MPA-first. A public, read-only `/api/v1/books` slice is not
roadmap work until a concrete consumer exists.

Revisit only when a named same-origin interactive feature, partner integration, or other external
consumer needs machine-readable public catalog data. Define the exact read contract, pagination,
error model, compatibility policy, tests, and OpenAPI scope then. Do not add CORS, client tokens,
or rate limiting in anticipation of that use case.

## Typed server-side rendering: Templ or gomponents

Status: deferred; no active implementation.

Current decision: keep `html/template` as the only application rendering path. The completed Templ
and gomponents spikes were removed from executable code during v0.2.3 cleanup so that catalog
changes do not require maintaining three parallel `BookCard` models and routes.

What the experiments showed:

- Templ gives typed component contracts and HTML-like source, but adds code generation, generated
  files, tooling, and adapter types.
- gomponents uses normal Go code and needs no generator, but larger HTML structures are less easy
  to scan and still require adapter types at the package boundary.
- `html/template` remains the simplest fit for the current MPA, where reusable UI is still small
  and existing partials are adequate.

Historical evidence:

- [Templ BookCard spike](ai/templ-spike-book-card.md)
- [Templ and gomponents comparison](ai/frontend-rendering-spike-book-card.md)

Revisit only if at least one of these becomes true:

- several non-trivial components are repeated across pages and template partial contracts become
  difficult to maintain;
- template-field mistakes or refactors cause recurring defects that typed components would catch;
- a planned UI change has enough scope to measure one typed approach against `html/template`.

If revisited, select one concrete component and one candidate technology. Define the expected
benefit and removal criteria before adding a dependency; do not restore both experimental stacks
or public spike routes by default.
