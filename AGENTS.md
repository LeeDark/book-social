# AGENTS.md

## Project

This is a Go Book Social learning project.

Architecture:
- modular monolith
- layered architecture
- HTTP handlers
- services / use cases
- repositories
- MPA / server-side templates
- persistence details follow the current architecture and roadmap documents

Useful project docs:
- `README.md`
- `docs/architecture.md`
- `docs/development.md`
- `docs/routes.md`
- `docs/domain.md`
- `docs/database_v0_1.md`
- `docs/roadmap.md`
- `docs/ai/project-context.md`
- `docs/ai/task-history.md`
- `docs/ai/ai-augmented-development-workflow.md`

Use `docs/roadmap.md` as the source of truth for the current priority, active stage, and deferred
work. Read it when the task depends on project sequencing, but do not load it for unrelated,
well-scoped tasks.

## Stable technical direction

- Keep the project simple and educational.
- Prefer clear Go code over clever abstractions.
- Avoid large refactoring unless the task explicitly asks for it.
- Keep package boundaries clean.
- Do not introduce heavy dependencies without a strong reason.

## Operating contract

The user's latest explicit instruction takes precedence over the defaults in this file.

- Prefer small, reviewable steps.
- Inspect only the files and documents relevant to the requested task.
- Do not turn small tasks into large rewrites.
- Do not widen the scope, redesign architecture, or add dependencies without a task-specific reason.
- Use `docs/ai/ai-augmented-development-workflow.md` for the full workflow and mode explanations.
- When the prompt names a working mode, apply the matching contract below.

## Working modes

### Manager Mode

- Inspect the relevant code before editing.
- Make the smallest focused change that satisfies the request.
- Preserve the existing project style and architecture.
- Run the narrowest relevant checks.
- Report the result and remaining risks.

### Documentation Mode

- Change only the requested documentation files.
- Do not change implementation files.
- Preserve technical meaning and useful history.
- Separate current behavior from planned behavior.
- Keep raw AI prompt logs in `docs/archive/` until reviewed.
- Do not treat experimental spike notes as the current project direction.

### Review Mode

- Inspect the requested change without editing files.
- List findings first and order them by severity.
- Include file and line references when possible.
- Mention test gaps and residual risks.

### Planning Mode

- Inspect only the context needed to prepare the plan.
- Define scope, stages, verification, and the definition of done.
- Do not implement or edit files unless the user explicitly changes modes.

### Book or Source Study Mode

- Summarize concepts in original wording without copying large source fragments.
- Connect relevant concepts to the current project.
- Separate ideas into applicable now, deferred, or not relevant.
- Do not implement source material until the user explicitly changes modes.

### Tutor Mode

- Treat the request as one interactive learning turn, not a persistent goal.
- Explain the concept and provide one focused exercise.
- Do not write the final implementation first.
- Do not edit files unless the user explicitly changes modes.
- Finish the response after the exercise and wait for the user.
- Waiting for the user is not a blocker.
- Do not poll the repository while waiting.
- Do not create a persistent goal or token budget.

### Pair Programmer Mode

- Treat the request as an interactive session, not a persistent goal.
- Let the user implement the first version.
- Before implementation, provide only scoped guidance and acceptance criteria, then wait.
- Review the diff only after the user says it is ready.
- Suggest minimal targeted fixes instead of rewriting the solution.
- Waiting for the user is not a blocker.
- Do not poll the repository while waiting.
- Do not create a persistent goal or token budget.

## Persistent goals and token budgets

- Create a persistent Codex goal (`/goal`) only when the user explicitly requests one.
- Set a token budget only when the user explicitly specifies it.
- A persistent goal must describe one bounded result that the agent can complete without waiting for
  user work.
- Do not use persistent goals for Tutor Mode or Pair Programmer Mode.
- Do not use pause as a substitute for ending an interactive turn.
- If the task requires user implementation or input, finish the current turn and wait normally.

## Verification contract

- Run the narrowest relevant check first.
- Run the full test suite when the scope or risk justifies it.
- Report the exact commands run and their results.
- Separate pre-existing failures from failures caused by the current change.
- If sandbox or environment restrictions prevent a check, explain the limitation and provide the
  exact local verification step.

## Completion contract

After an implementation or documentation change, summarize:

- files changed;
- behavior or documentation changed;
- tests and checks run;
- failures, residual risks, or work intentionally left for later.

Do not commit changes unless the user explicitly requests it.

## Testing

- Use the standard Go testing package.
- Prefer table-driven tests.
- Use `httptest` for HTTP handlers.
- Use fake repositories/services for unit tests.
- Avoid database integration tests unless explicitly requested.
- Prefer `make test` for the full project test suite when the task scope or risk justifies it.

## UI

- This is an MPA project.
- Do not introduce frontend frameworks.
- Keep templates simple.
- Do not over-test HTML markup.

## Before changing code

1. Inspect the existing structure.
2. Explain briefly what you plan to change.
3. Make the smallest reasonable change.
4. Run or explain the relevant tests.

## Running the web server in Codex

Do not try to start the web server inside the Codex sandbox for verification.

Avoid commands like:

```bash
GOCACHE=/tmp/book-social-go-cache APP_HTTP_ADDR=:18080 go run ./cmd/web
curl -I http://localhost:18080/books
```

The Codex sandbox may not allow opening listening sockets or accessing `localhost`, so these checks can fail with environment errors such as:

```text
listen tcp :18080: socket: operation not permitted
curl: (7) Couldn't connect to server
```

These errors should be treated as sandbox/environment limitations, not as project failures.

For automated verification, prefer:

```bash
GOCACHE=/tmp/book-social-go-cache go test ./...
```

For HTTP behavior, use Go tests with `net/http/httptest` instead of starting a real server.

For visual/manual checks, report the exact routes the user should open locally, for example:

```text
/
/books
/books/{valid-slug}
/books/unknown-slug
```

If a task requires browser verification, stop and ask the user to run the app locally outside the Codex sandbox.
