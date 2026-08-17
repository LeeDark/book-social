# Контракт MPA-аутентификации v0.2.5–v0.2.6

> **Статус: принятый черновик реализации; поведение ещё не реализовано.**
>
> Документ фиксирует решения для review и реализации. Текущее реализованное поведение описывают
> `docs/routes.md`, `docs/domain.md`, `docs/database.md` и `docs/development.md`.

## 1. Цель, граница и существующая основа

Цель диапазона — минимальный MPA-путь:

```text
анонимный посетитель
  → регистрация или вход
  → аутентифицированная сессия
  → GET /me
  → выход
```

v0.2.5 создаёт внутренний контракт identity, паролей, сессий, CSRF и validation; v0.2.6 добавит
страницы и формы. MPA и `html/template` остаются основным UI, buffered renderer и layered boundaries
сохраняются.

v0.2.4 HTTP-основа уже закрыта. Существующие `/healthz`, `/static/*`, recovery, request ID, trusted
real IP, request logging, security headers и route-scoped timeout не расширяются auth scope. Auth
применяется только к dynamic MPA routes.

Используются существующие `users` и `roles`. Первый этап включает только registration, login, logout
и `GET /me`; не включает profile, activation, recovery, password change/reset, API auth, RBAC/admin
routes, CORS, rate limiting, JWT, OpenAPI или private library v0.3.

Защищаемые активы: password/hash, user record, session token/store row, CSRF token, flash message и
private resource. Секреты и непрозрачные идентификаторы не попадают в response, page model,
structured log, metrics или test fixtures.

## 2. Участники, identity и авторизация

| Участник                            | Разрешено                                                                      | Запрещено / результат                                                                     |
|-------------------------------------|--------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------|
| Анонимный посетитель                | Публичные MPA-страницы и формы регистрации/входа с CSRF.                       | `GET /me` и будущие private actions: MPA redirect/refusal.                                |
| Аутентифицированный пользователь    | `GET /me` с собственной минимальной identity; `POST /logout` для своей сессии. | Передавать identity через query/header, получать admin access или чужой private resource. |
| Администратор / повышенный участник | Не входит в этот release.                                                      | `is_admin` и database role не создают route, shortcut или bypass.                         |

Аутентификация устанавливает minimal current-user identity через typed request context. Авторизация
проверяет право на private action в use case или другой явной server-side boundary. Navigation,
hidden form fields и database role не являются security control.

Текущая protected boundary — только `GET /me`: authenticated current user only. Missing, invalid или
expired session означает anonymous state, redirect на `/login`, отсутствие `500` и отсутствие
disclosure private resource. Handler получает только identity текущего пользователя; password/hash,
session token, role internals и future-library data в page model не входят.

Ownership library data намеренно отложен до v0.3. Для каждой будущей private action заранее
обязателен contract: anonymous → refusal; owner → allowed operation; non-owner → refusal without
disclosure. Admin case появляется только после отдельного product decision.

## 3. Маршруты, формы и CSRF

`next` в первом этапе отсутствует: успешные registration и login всегда перенаправляют на `/me`.
Если return-to понадобится позднее, допустимы только validated local paths с fallback `/me`.

| Method/path      | Участник и CSRF                                                  | Успех                                                                        | Отказ / побочный эффект                                                                                            |
|------------------|------------------------------------------------------------------|------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------|
| `GET /register`  | Анонимный; выдаёт CSRF token форме.                              | `200`, accessible form.                                                      | Authenticated user → `303 /me`; session state не создаётся.                                                        |
| `POST /register` | Анонимная form submit; CSRF required.                            | Validate fields; atomically create user + default role + session; `303 /me`. | Invalid/duplicate → `422` safe field errors; password не repopulate. Store/internal → generic `500`, redacted log. |
| `GET /login`     | Анонимный; выдаёт CSRF token форме.                              | `200`, login form.                                                           | Authenticated user → `303 /me`; session state не создаётся.                                                        |
| `POST /login`    | Анонимная form submit; CSRF required.                            | Verify credentials; create/rotate new session; `303 /me`.                    | Invalid credentials → neutral `422`; CSRF/store failure → safe outcome, redacted log.                              |
| `POST /logout`   | Authenticated, stale или anonymous browser state; CSRF required. | Invalidate own session if present, clear cookie, `303 /`.                    | Missing/stale session idempotent: cookie is still cleared; no `GET /logout`.                                       |
| `GET /me`        | Authenticated current user only; no state change.                | `200`, minimal current-user identity.                                        | Missing/invalid/expired session → `303 /login`; no disclosure and no `500`.                                        |

GET/HEAD не меняют state. Все browser POST routes защищены единым server-side CSRF mechanism; token
передаётся только в forms и не попадает в logs.

## 4. Учётная запись, validation и domain errors

Registration принимает только `first_name`, `login`, `email`, `password` и confirmation;
`second_name` и `sur_name` в форму не входят. Обычная роль `user` гарантируется migration/seed,
назначается только server-side; отсутствие role в corrupted/manual DB — internal error.

`first_name`, `login` и `email` trim-ятся; login/email сравниваются в canonical форме. Максимумы:
`first_name` 100, `login` 64, `email` 254. Password — 12–128 символов, confirmation совпадает;
complexity rules не вводятся и password никогда не повторно отображается.

Domain boundary определяет `ValidationError`, `ErrLoginTaken`, `ErrEmailTaken`,
`ErrInvalidCredentials`, `ErrUnauthenticated` и internal error. Handler сопоставляет их с
`422`, `303` или generic `500`; не сравнивает database-driver error strings.

Duplicate registration получает safe field errors. Unknown login/email и wrong password всегда дают
одинаковые external `422` и `Invalid login or password`; причина — только в redacted log.

## 5. Password policy

Используется bcrypt как maintained adaptive hash; cost находится в одном auth package. Хранится
только hash: custom hashing, reversible encryption, plaintext comparison и plaintext storage
запрещены. Plaintext существует только в form input и narrow hash/verify path; password/hash не
попадают в DTO, template, response, error, log, metric или fixture.

## 6. Сессии и cookie

Session — DB-backed opaque session. Cookie содержит cryptographically random raw token, database —
только cryptographic hash, `user_id`, `created_at`, `expires_at`. Raw token не хранится в DB/logs, а
`Set-Cookie` отправляется только после successful authentication.

Сессия имеет absolute lifetime 7 days без sliding renewal. Registration/login создают или rotate
новый identifier; request ищет hash, проверяет expiry и кладёт minimal identity в context.
Missing/invalid/ expired token — anonymous, не internal error; допустима lazy cleanup при
load/delete, без worker. Logout delete/invalidate сессию и clear `book_social_session`, даже если DB
row уже отсутствует; операция idempotent и old-token reuse получает refusal.

Cookie: `book_social_session`, `HttpOnly`, `Path=/`, `SameSite=Lax`; `Secure` для stage/prod,
development exception задаётся configuration и документируется. Session middleware работает только с
dynamic routes, которым нужен state; static/health не создают store writes. Allowlist session
contents:
minimal `user_id`/identity и one-time flash data; password, password hash, raw CSRF, raw session
token и private content запрещены.

## 7. Persistence, transactions и HTTP boundary

Следующая migration — `000003`; `000001` не переписывается. SQLite и PostgreSQL получают
эквивалентную `sessions`: `id`, FK `user_id`, unique `token_hash`, `created_at`,
`expires_at`, expiry index и одинаковые timestamp semantics/constraints. Нужен reversible down;
seed/reset/test bootstrap создают role `user` и проходят migration/seed smoke.

Service/use case владеет transaction, когда меняется несколько таблиц: registration atomically
создаёт user, получает default role и создаёт session; successful login создаёт новую session.
Одиночный delete/invalidate session может быть repository call.

Middleware order сохраняет v0.2.4:
`SecurityHeaders → RequestID → TrustedRealIP → RequestLogger → Recoverer → session/current-user → CSRF → route guard`.
Application timeout динамических MPA routes сохраняется.

## 8. Ошибки, журналирование и злоупотребления

Клиент получает stable safe outcome, server log — только request ID, route, operation, safe actor
state и typed error/event class. Denylist: password, password hash, raw session ID, CSRF token,
cookie/authorization headers, submitted credentials и private resource content.

| Событие                          | Клиент                                             | Журнал                                                     |
|----------------------------------|----------------------------------------------------|------------------------------------------------------------|
| Invalid registration / duplicate | `422` safe field errors; password не возвращается. | validation/conflict class и имена полей без значений.      |
| Invalid credentials              | Neutral `422`, `Invalid login or password`.        | `invalid_credentials`; без login/email, password или hash. |
| Session / CSRF failure           | Safe refusal; `/me` → `303 /login`.                | anonymous/expired или `csrf_rejected`; без token/cookie.   |
| Store/hash/internal failure      | Generic `500`, без DB/stack/credential detail.     | typed internal class и safe cause.                         |
| Success/logout/denied action     | Normal redirect/refusal.                           | outcome class; без secrets/private content.                |

Event classes включают registration validation/conflict, login success/failure, session
create/load/expire/invalidate, logout, denied protected action и unexpected error. CORS и API rate
limiting — Stage 7B; login throttling/account lockout требуют отдельного product/abuse decision.

## 9. Порядок реализации и проверка

1. `000003` migrations, seed и test bootstrap.
2. `internal/modules/users`: models, repository interfaces, password policy, service/use cases и
   unit tests без HTTP imports.
3. SQLite/PostgreSQL repositories и contract/parity tests.
4. HTTP cookie manager, current-user middleware, CSRF и protected-route guard.
5. v0.2.6 handlers, templates, navigation и flash messages.
6. Narrow tests, затем `GOCACHE=/tmp/book-social-go-cache make test`,
   `GOCACHE=/tmp/book-social-go-cache go vet ./...`, `git diff --check` и
   `git status --short --branch`. HTTP проверяется через `httptest`, не real listener; PostgreSQL —
   отдельно на disposable DSN как environment-dependent evidence.

## Definition of Done v0.2.5

v0.2.5 завершена, когда одновременно выполнено всё ниже:

- Контракт выше reviewable, без security-affecting TBD; scope и deferred items отделены от
  implemented behavior.
- SQLite/PostgreSQL migration `000003`, reversible down, seed/reset/test bootstrap и migration/seed
  smoke создают ordinary `user` role и opaque sessions equivalently.
- Auth core создаёт user с bcrypt hash, validates input, distinguishes domain errors, verifies
  credentials и не хранит/не раскрывает plaintext, hash или password-derived details.
- DB-backed opaque sessions создаются/rotate после successful authentication, загружают current
  user, истекают через 7 days, очищаются lazy, invalidated на logout и не позволяют reuse старого
  token.
- Cookie policy, typed current-user context, route guard и CSRF защищают все browser POST routes;
  GET/HEAD не меняют state, anonymous `/me` получает `303 /login`, а navigation не заменяет
  authorization.
- Handler/service/repository/`httptest` tests покрывают success и refusal paths: validation,
  duplicate identity, neutral login failure, expiry/invalidation, current-user context, CSRF,
  anonymous redirect и no-secret response/logging behavior.
- Client/server error and logging policy работает: safe outcomes, generic internal errors, request
  ID/event class и denylist без credentials, tokens, headers или private content.
- Документация фиксирует password/session policy, lifecycle, configuration и development-vs-HTTPS
  distinction; verification commands проходят без real listener.
