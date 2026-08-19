# Контракт MPA-аутентификации v0.2.5–v0.2.6

> **Статус: принятый versioned contract.**
>
> Foundation v0.2.5 находится на финальном implementation review; user-facing workflow v0.2.6 ещё
> не реализован. Текущее принятое поведение описывают `docs/routes.md`, `docs/domain.md`,
> `docs/database.md` и `docs/development.md`. Пошаговые планы разделены на
> `docs/private/plan-v0_2_5.ru.md` и `docs/private/plan-v0_2_6.ru.md`.

## 1. Цель, граница и существующая основа

Цель диапазона — минимальный MPA-путь:

```text
анонимный посетитель
  → регистрация или вход
  → аутентифицированная сессия
  → GET /me
  → выход
```

Ответственность между версиями разделена и не используется взаимозаменяемо:

| Версия                           | Входит                                                                                                                                                                       | Не является её release gate                                                                        |
|----------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------|
| v0.2.5 Auth Foundation           | Contract, password/validation policy, migrations, user/session service и repositories, cookie/current-user/guard foundation, global `CrossOriginProtection` и focused tests. | Registration/login/logout pages, production `/me`, auth navigation, flashes и полный browser flow. |
| v0.2.6 Registration/Login/Logout | Routes, handlers, forms/templates, application wiring, navigation/flashes и полный register/login/`/me`/logout flow.                                                         | Изменение session strategy, API auth, RBAC, account lifecycle или private-library ownership.       |

MPA и `html/template` остаются основным UI, buffered renderer и layered boundaries сохраняются.
v0.2.6 начинается только после закрытого v0.2.5 foundation; отсутствие v0.2.6 routes не блокирует
закрытие v0.2.5.

v0.2.4 HTTP-основа уже закрыта. Существующие `/healthz`, `/static/*`, recovery, request ID, trusted
real IP, request logging, security headers и route-scoped timeout не расширяются auth scope. Auth
применяется только к dynamic MPA routes.

Используются существующие `users` и `roles`. Общий диапазон v0.2.5–v0.2.6 включает только foundation
и последующий registration/login/logout/`GET /me` workflow; не включает profile, activation,
recovery, password change/reset, API auth, RBAC/admin routes, CORS, rate limiting, JWT, OpenAPI или
private library v0.3.

Защищаемые активы: password/hash, user record, session token/store row, flash message и private
resource. Секреты и непрозрачные идентификаторы не попадают в response, page model, structured log,
metrics или test fixtures.

## 2. Участники, identity и авторизация

Матрица ниже задаёт наблюдаемое поведение будущих routes v0.2.6. В v0.2.5 эти actors нужны для
проектирования identity/context/guard boundaries, но публичные auth routes ещё отсутствуют.

| Участник                            | Разрешено                                                                                                 | Запрещено / результат                                                                     |
|-------------------------------------|-----------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------|
| Анонимный посетитель                | Публичные MPA-страницы и формы регистрации/входа; unsafe browser requests проходят cross-origin проверку. | `GET /me` и будущие private actions: MPA redirect/refusal.                                |
| Аутентифицированный пользователь    | `GET /me` с собственной минимальной identity; `POST /logout` для своей сессии.                            | Передавать identity через query/header, получать admin access или чужой private resource. |
| Администратор / повышенный участник | Не входит в этот release.                                                                                 | `is_admin` и database role не создают route, shortcut или bypass.                         |

Аутентификация устанавливает minimal current-user identity через typed request context. Авторизация
проверяет право на private action в use case или другой явной server-side boundary. Navigation,
hidden form fields и database role не являются security control.

Единственная planned protected boundary v0.2.6 — `GET /me`: authenticated current user only.
Missing, invalid или expired session означает anonymous state, redirect на `/login`, отсутствие
`500` и отсутствие disclosure private resource. Handler получает только identity текущего
пользователя; password/hash, session token, role internals и future-library data в page model не
входят. v0.2.5 должна предоставить testable typed context и guard, но не обязана регистрировать
`/me` в production.

Ownership library data намеренно отложен до v0.3. Для каждой будущей private action заранее
обязателен contract: anonymous → refusal; owner → allowed operation; non-owner → refusal without
disclosure. Admin case появляется только после отдельного product decision.

## 3. Маршруты, формы и CSRF — target v0.2.6

Таблица ниже является route contract v0.2.6, а не описанием текущего v0.2.5 behavior. `next` в
первом user-facing slice отсутствует: успешные registration и login перенаправляют на `/me`. Если
return-to понадобится позднее, допустимы только validated local paths с fallback `/me`.

| Method/path      | Участник и защита от CSRF                                                                        | Успех                                                                        | Отказ / побочный эффект                                                                                                    |
|------------------|--------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------|
| `GET /register`  | Анонимный; token форме не требуется.                                                             | `200`, accessible form.                                                      | Authenticated user → `303 /me`; session state не создаётся.                                                                |
| `POST /register` | Анонимная form submit; unsafe browser request проверяется middleware.                            | Validate fields; atomically create user + default role + session; `303 /me`. | Invalid/duplicate → `422` safe field errors; password не repopulate. Cross-origin browser request → safe `403`.            |
| `GET /login`     | Анонимный; token форме не требуется.                                                             | `200`, login form.                                                           | Authenticated user → `303 /me`; session state не создаётся.                                                                |
| `POST /login`    | Анонимная form submit; unsafe browser request проверяется middleware.                            | Verify credentials; create/rotate new session; `303 /me`.                    | Invalid credentials → neutral `422`; cross-origin browser request → safe `403`; store failure → redacted `500`.            |
| `POST /logout`   | Authenticated, stale или anonymous browser state; unsafe browser request проверяется middleware. | Invalidate own session if present, clear cookie, `303 /`.                    | Missing/stale session idempotent: cookie is still cleared; no `GET /logout`; cross-origin browser request не меняет state. |
| `GET /me`        | Authenticated current user only; no state change.                                                | `200`, minimal current-user identity.                                        | Missing/invalid/expired session → `303 /login`; no disclosure and no `500`.                                                |

GET/HEAD не меняют state. v0.2.5 подключает единый `http.CrossOriginProtection` к application router
и проверяет его на test-only handler; v0.2.6 обязана провести через него все реальные unsafe browser
routes:
middleware отклоняет cross-origin request по Fetch Metadata (`Sec-Fetch-Site`) либо `Origin` с safe
`403`; forms не получают CSRF token. Запрос без обоих этих заголовков допускается как non-browser
или unknown request, поэтому production использует HTTPS, а `SameSite=Lax` остаётся независимым
defense-in-depth барьером. Middleware включается для application router после `Recoverer`; не
использовать `AddInsecureBypassPattern`, а trusted origin добавлять только при отдельной будущей
потребности.

## 4. Учётная запись, validation и domain errors

Foundation v0.2.5 определяет registration input и service validation для `first_name`, `login`,
`email`, `password` и confirmation; `second_name` и `sur_name` не входят ни в input, ни в будущую
форму. Обычная роль `user` гарантируется migration/seed, назначается только server-side; отсутствие
role в corrupted/manual DB — internal error. v0.2.6 отображает эти правила через form/handler, не
создавая второй набор validation rules.

`first_name`, `login` и `email` trim-ятся; login/email сравниваются в canonical форме. Максимумы:
`first_name` 100, `login` 64, `email` 254. Password — 12–128 символов, confirmation совпадает;
complexity rules не вводятся и password никогда не повторно отображается.

Domain boundary v0.2.5 определяет `ValidationError`, `ErrLoginTaken`, `ErrEmailTaken`,
`ErrInvalidCredentials`, `ErrUnauthenticated` и internal error. Handler v0.2.6 сопоставляет их с
`422`, `303` или generic `500`; он не сравнивает database-driver error strings.

Duplicate registration получает safe field errors. Unknown login/email и wrong password всегда дают
одинаковые external `422` и `Invalid login or password`; причина — только в redacted log.

## 5. Password policy

v0.2.5 использует bcrypt как maintained adaptive hash; cost находится в одном auth package и
выбирается по допустимой пользовательской задержке и нагрузке, а не копируется как непроверенная
константа. Хранится только hash: custom hashing, reversible encryption, plaintext comparison и
plaintext storage запрещены. Plaintext существует только в form input и narrow hash/verify path;
password/hash не попадают в DTO, template, response, error, log, metric или fixture.

## 6. Сессии и cookie

Session — DB-backed opaque session. Cookie содержит cryptographically random raw token, database —
только cryptographic hash, `user_id`, `created_at`, `expires_at`. Raw token не хранится в DB/logs, а
`Set-Cookie` отправляется только после successful authentication и успешной session persistence.

Сессия имеет absolute lifetime 7 days без sliding renewal. v0.2.5 предоставляет create/load/delete,
expiry semantics, token hashing, cookie manager и current-user context как foundation, но не обязана
вызывать их из отсутствующих login/logout handlers. v0.2.6 после успешной registration/login создаёт
или rotate новый identifier; request ищет hash, проверяет expiry и кладёт minimal identity в
context. Missing/invalid/expired token — anonymous, не internal error; допустима lazy cleanup при
load/delete, без worker. Logout v0.2.6 delete/invalidate сессию и clear `book_social_session`, даже
если DB row уже отсутствует; операция idempotent и old-token reuse получает refusal.

Cookie: `book_social_session`, `HttpOnly`, `Path=/`, `SameSite=Lax`; `Secure` для stage/prod,
development exception задаётся configuration и документируется. v0.2.5 фиксирует и тестирует эту
policy; v0.2.6 применяет её при реальном Set-Cookie. Session/current-user middleware в v0.2.6
работает только с dynamic routes, которым нужен state; static/health не создают store writes.
Allowlist session contents:
minimal `user_id`/identity и one-time flash data; password, password hash, raw session token и
private content запрещены.

## 7. Persistence, transactions и HTTP boundary

Следующая migration — `000003`; `000001` не переписывается. SQLite и PostgreSQL получают
эквивалентную `sessions`: `id`, FK `user_id`, unique `token_hash`, `created_at`,
`expires_at`, expiry index и одинаковые timestamp semantics/constraints. Нужен reversible down;
seed/reset/test bootstrap создают role `user` и проходят migration/seed smoke.

Service/use case владеет transaction, когда меняется несколько таблиц. v0.2.5 создаёт foundation для
server-owned default-role lookup и create user. Route contract v0.2.6 требует, чтобы successful
registration atomically создавала user и session; если foundation не может обеспечить эту
transaction boundary, v0.2.6 не начинается до исправления foundation или явного пересмотра success
behavior в contract. Successful login создаёт новую session; одиночный delete/invalidate session
может быть repository call.

Полный middleware order после wiring v0.2.6 сохраняет v0.2.4:
`SecurityHeaders → RequestID → TrustedRealIP → RequestLogger → Recoverer → session/current-user → CrossOriginProtection → route guard`.
В v0.2.5 production chain заканчивается глобальным `CrossOriginProtection`, а current-user/route
guard остаются testable foundation до появления auth dependencies и `/me`. Application timeout
динамических MPA routes сохраняется. Защищённые HTML-ответы v0.2.6 не получают публичный cache
policy: для них устанавливается `Cache-Control: no-store`.

## 8. Ошибки, журналирование и злоупотребления

Клиент получает stable safe outcome, server log — только request ID, route, operation, safe actor
state и typed error/event class. Denylist: password, password hash, raw session ID,
cookie/authorization headers, submitted credentials и private resource content.

v0.2.5 отвечает за typed domain errors, redacted service/repository outcomes и отсутствие secrets в
foundation responses/logs/test output. Полная client-status/event-class матрица применяется в
v0.2.6, когда появляются handlers и реальные auth events.

| Событие                          | Клиент                                             | Журнал                                                     |
|----------------------------------|----------------------------------------------------|------------------------------------------------------------|
| Invalid registration / duplicate | `422` safe field errors; password не возвращается. | validation/conflict class и имена полей без значений.      |
| Invalid credentials              | Neutral `422`, `Invalid login or password`.        | `invalid_credentials`; без login/email, password или hash. |
| Session / CSRF failure           | Safe refusal; `/me` → `303 /login`.                | anonymous/expired или `csrf_rejected`; без cookie values.  |
| Store/hash/internal failure      | Generic `500`, без DB/stack/credential detail.     | typed internal class и safe cause.                         |
| Success/logout/denied action     | Normal redirect/refusal.                           | outcome class; без secrets/private content.                |

Event classes включают registration validation/conflict, login success/failure, session
create/load/expire/invalidate, logout, denied protected action и unexpected error. CORS и API rate
limiting — Stage 7B; login throttling/account lockout требуют отдельного product/abuse decision.

## 9. Порядок реализации и проверка

v0.2.5 выполняется и закрывается отдельно:

1. `000003` migrations, seed и test bootstrap.
2. `internal/modules/users`: models, repository interfaces, password policy, service/use cases и
   unit tests без HTTP imports.
3. SQLite/PostgreSQL repositories и contract/parity tests.
4. HTTP cookie manager, current-user middleware, global CrossOrigin protection и testable
   protected-route guard.
5. Foundation verification, documentation и отдельный v0.2.5 release review.

Только после этого v0.2.6 добавляет handlers, templates, navigation, flashes, production auth wiring
и полный browser-style flow. У каждой версии собственный DoD и собственный closure; завершение
пунктов v0.2.6 не используется как условие закрытия v0.2.5.

Для обеих версий narrow tests предшествуют `GOCACHE=/tmp/book-social-go-cache make test`,
`GOCACHE=/tmp/book-social-go-cache go vet ./...`, `git diff --check` и
`git status --short --branch`. HTTP проверяется через `httptest`, не real listener; PostgreSQL —
отдельно на disposable DSN как environment-dependent evidence.

## 10. Review и release gates

Contract review был завершён до первоначального изменения schema, dependency list или application
code:

- Chapter 8 и Chapter 10 *Let's Go* использованы как source study, без переноса Snippetbox routes,
  MySQL schema, 12-hour lifetime или готового session manager.
- Contract определяет actors, authorization boundary, route/form outcomes, password policy,
  session/cookie lifecycle, CSRF, error/logging policy и verification; security-affecting `TBD` нет.
- Продуктовая граница сохранена: private library ownership, roles/RBAC, API security и account
  recovery отложены до своих trigger.
- v0.2.5 implementation проходит собственный финальный review; UI/forms остаются задачей v0.2.6 и не
  используются для преждевременного закрытия foundation.

## Definition of Done v0.2.5

v0.2.5 завершена, когда одновременно выполнено всё ниже:

- Контракт reviewable, без security-affecting TBD; foundation, planned v0.2.6 behavior и deferred
  work явно разделены.
- SQLite/PostgreSQL migration `000003`, reversible down, seed/reset/test bootstrap и migration/seed
  smoke создают ordinary `user` role и opaque sessions equivalently.
- Auth core создаёт user с bcrypt hash, validates input, distinguishes domain errors, verifies
  credentials и не хранит/не раскрывает plaintext, hash или password-derived details.
- DB-backed session repositories/services создают, загружают и удаляют hashed-token sessions,
  отклоняют missing/invalid/expired session и согласованы с absolute seven-day lifetime policy.
- Cookie manager генерирует opaque high-entropy token, применяет принятую cookie policy и позволяет
  consumer записать `Set-Cookie` только после успешной session persistence; raw token не попадает в
  DB, log или test failure output.
- Typed current-user middleware различает anonymous, valid session и internal store failure;
  route-guard foundation проверена на test-only handler, но production `/me` ещё не требуется.
- `http.CrossOriginProtection` подключён к production application chain после `Recoverer` без bypass
  patterns и проверен на test-only unsafe cross-origin refusal/same-origin success. Реальные auth
  POST routes ещё не требуются.
- Unit/repository/`httptest` tests покрывают foundation success/refusal paths: validation, duplicate
  identity mapping, neutral credential failure, session expiry/invalidation, current-user context,
  guard и CrossOrigin behavior; secrets не печатаются при failure.
- Документация фиксирует password/session policy, lifecycle, configuration и development-vs-HTTPS
  distinction; verification commands проходят без real listener. UI и navigation не заявляют, что
  registration/login уже доступны.

Отсутствие registration/login/logout handlers, production `/me`, auth navigation, flashes и полного
browser flow не является нарушением Definition of Done v0.2.5.

## Definition of Done v0.2.6

v0.2.6 завершена только после закрытой v0.2.5 и когда одновременно выполнено всё ниже:

- `GET/POST /register`, `GET/POST /login`, `POST /logout` и protected `GET /me` реализуют route
  contract выше; state-changing `GET` route отсутствует.
- Successful registration выполняет заранее принятый atomic user + default role + session outcome;
  successful login создаёт/rotate session только после credential verification и только затем
  отправляет cookie.
- Duplicate registration и invalid credentials имеют useful safe UI outcomes; unknown account и
  wrong password не раскрывают существование account сообщением, detail или практически наблюдаемым
  различием verification path.
- Current-user middleware подключён перед `CrossOriginProtection`, `/me` зарегистрирован через route
  guard, anonymous получает `303 /login`, authenticated user видит только свою minimal identity.
- Logout idempotently invalidates текущую session, очищает cookie даже для stale state и не
  допускает reuse старого token.
- Каждый реальный unsafe browser route проходит `CrossOriginProtection`: cross-origin request
  получает safe `403` без mutation, same-origin request достигает handler; GET/HEAD не меняют state.
- Templates, page models, flashes, responses, logs и test output не содержат password, password
  hash, raw session token, cookie headers или internal store/private details.
- Handler/router/template tests покрывают forms, validation, conflicts, neutral login refusal,
  cookie attributes, current-user context, anonymous/authenticated `/me`, logout/reuse и каждый
  CrossOrigin success/refusal path; manual browser smoke проверяет accessibility и navigation.
- `make test`, `go vet ./...`, lint и `git diff --check` проходят; PostgreSQL parity указана
  отдельно, а routes/domain/database/testing/roadmap/task documentation описывает фактический v0.2.6
  behavior.

## Правило перехода между версиями

v0.2.5 может быть закрыта без user-facing auth workflow, но прикладная реализация v0.2.6 не
начинается на непринятом foundation. Applied Stage 7A evidence для полного MPA flow добавляется
только после принятого v0.2.6 commit/tag; foundation evidence v0.2.5 обозначается отдельно и не
выдаётся за готовый registration/login/logout path.
