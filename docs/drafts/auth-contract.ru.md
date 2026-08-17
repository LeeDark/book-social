# Черновик контракта MPA-аутентификации v0.2.5–v0.2.6

> **Статус: принятый implementation draft; поведение ещё не реализовано.**
>
> Этот документ — точная переносимая копия двух контрактных блоков из прежней private-заметки:
> нового «Шаг 0–7» и прежнего «Scope и account → Definition of Done v0.2.5».
> Он предназначен для review и последующей чистки. Текущее реализованное поведение описывают
> `docs/routes.md`, `docs/domain.md`, `docs/database.md` и `docs/development.md`.

## Принятые решения

### Шаг 0 — подготовка входных данных

Перед реализацией сверены текущие границы `book-social` v0.2.5–v0.2.6 с кодом и документацией:

- MPA остаётся основным интерфейсом; актуальная HTTP-цепочка зафиксирована в `docs/routes.md` и
  `internal/app/app.go`/`internal/app/routes.go`.
- `v0.2.4` HTTP Foundation закрыта: `/healthz`, `/static/*`, recovery, request ID, trusted real IP,
  request logging, security headers и route-scoped timeout не относятся к auth/session scope.
- Источник истины для auth scope — существующие `users`/`roles` schema foundations, roadmap v0.2.5 и
  решения ниже; Greenlight/Snippetbox examples используются только как study reference.
- Первый slice ограничен registration, login, logout и `GET /me`; не добавляются profile,
  activation, recovery, API auth, roles/RBAC routes, CORS, rate limiting, JWT или OpenAPI.
- Для каждого asset и actor заранее определены границы: password/hash, user record, session token и
  store row, CSRF token, flash message и private resource. Secrets и opaque session IDs не попадают
  в response, page model или structured log.

**Результат шага 0:** входной scope и существующие boundaries подтверждены до изменения schema,
dependency list или application code. Открытые decisions не скрываются выбором library; следующие
подразделы фиксируют их как contract.

### Шаг 1 — actors и authorization model

Для первого slice не вводится generic RBAC. Нужны два фактических actor-а и один явный default-deny
boundary:

| Actor                    | Может                                                                                                        | Не может / outcome                                                                                                                  |
|--------------------------|--------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------|
| `anonymous`              | Читать public MPA pages, открыть registration/login forms, отправить auth form с валидным CSRF.              | Открыть `GET /me` или любую будущую private action; получает redirect на `/login` или другой принятый MPA refusal.                  |
| `authenticated user`     | Открыть `GET /me`, видеть только собственную minimal identity и завершить свою session через `POST /logout`. | Передать identity через query/header, получить admin access или чужой private resource; при отсутствии owner rule получает refusal. |
| `admin` / elevated actor | Не входит в v0.2.5–v0.2.6.                                                                                   | `is_admin` не открывает routes и не меняет authorization model этого release.                                                       |

#### Что нужно сделать

- [x] Записать actor × action matrix для всех принятых routes: `/register`, `/login`, `/logout`,
  `/me`, включая method, authentication requirement, CSRF requirement и expected outcome.
- [x] Явно разделить authentication и authorization: current-user middleware устанавливает minimal
  identity, но решение о праве на private action принимает use case/authorization boundary.
- [x] Зафиксировать, что navigation links не являются security control: скрытие `/me` или Logout в
  template не заменяет route guard и server-side identity check.
- [x] Для `GET /me` записать минимальный data boundary: handler получает только identity текущего
  user; password/hash, session token, role internals и private future-library data не входят в page
  model.
- [x] Зафиксировать default-deny behavior для missing, invalid и expired session: anonymous state,
  предсказуемый MPA redirect, отсутствие 500 и отсутствие disclosure существования private resource.
- [x] Не добавлять owner/non-owner branch в v0.2.5–v0.2.6 там, где private resource ещё не входит в
  release. Для будущей private action отдельно записать owner source, allowed operation и refusal
  outcome до её реализации.
- [x] Не считать `is_admin` или database role автоматическим разрешением: обычная роль `user`
  назначается server-side, а отсутствие default role в corrupted/manual DB даёт internal error.
- [x] Составить test triples для каждого protected action: anonymous, authenticated owner и
  authenticated non-owner; admin добавляется только после отдельного product decision.

#### Принятые решения для текущего release

| Route            | Anonymous                                                                                      | Authenticated user                            | Admin/elevated actor       |
|------------------|------------------------------------------------------------------------------------------------|-----------------------------------------------|----------------------------|
| `GET /register`  | `200` form                                                                                     | `303 /me`                                     | Не выделяется отдельно     |
| `POST /register` | CSRF + validation; success creates user/session and `303 /me`                                  | Не является обычным path; `303 /me`           | Не выделяется отдельно     |
| `GET /login`     | `200` form                                                                                     | `303 /me`                                     | Не выделяется отдельно     |
| `POST /login`    | CSRF + credentials; success creates session and `303 /me`; invalid credentials → neutral `422` | Не является обычным path; `303 /me`           | Не выделяется отдельно     |
| `POST /logout`   | CSRF policy выполняется; stale/missing session очищается и `303 /`                             | Invalidate own session, clear cookie, `303 /` | Не выделяется отдельно     |
| `GET /me`        | `303 /login`; no resource existence disclosure                                                 | `200`, minimal current-user identity only     | Не имеет отдельного bypass |

`GET /register` и `GET /login` не создают session state. `POST /register`, `POST /login` и
`POST /logout` являются browser state-changing routes и защищаются CSRF. Ни один route не принимает
identity из query или header, а navigation visibility не считается authorization.

- Первый protected route — `GET /me`; его authorization rule: authenticated current user only.
- Registration, login и logout — authentication actions, а не generic resource authorization.
- `/healthz` и `/static/*` публичны и не получают actor/session behavior без отдельной причины.
- `is_admin` остаётся данными для будущего scope и не создаёт admin route, shortcut или bypass.
- Первый release не имеет private library resource, поэтому owner/non-owner implementation
  откладывается; contract сохраняет test template, чтобы следующая private action не появилась без
  явного rule.

**Результат шага 1:** reviewable actor × action matrix, отдельное правило current-user identity для
`GET /me`, explicit default-deny/refusal behavior и зафиксированная граница «roles/RBAC и ownership
private resources пока deferred». Ни один route не полагается только на template navigation.

### Шаг 2 — routes, forms и outcomes

Первый auth slice использует только эти routes; параметр `next` сознательно не принимается, поэтому
после успешной registration или login target всегда `/me`.

| Method/path      | Actor и CSRF                                                                          | Success outcome                                                              | Failure outcome / side effect                                                                                                                      |
|------------------|---------------------------------------------------------------------------------------|------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|
| `GET /register`  | Anonymous; CSRF token выдаётся в form.                                                | `200`, accessible registration form.                                         | Authenticated user получает `303 /me`; session state не создаётся.                                                                                 |
| `POST /register` | Anonymous form submit; CSRF required.                                                 | Validate fields, atomically create user + default role + session, `303 /me`. | Invalid/duplicate input → `422` с safe field errors; password не repopulate. Store/internal failure → generic `500`, redacted log.                 |
| `GET /login`     | Anonymous; CSRF token выдаётся в form.                                                | `200`, login form.                                                           | Authenticated user получает `303 /me`; session state не создаётся.                                                                                 |
| `POST /login`    | Anonymous form submit; CSRF required.                                                 | Verify credentials, create a new session, `303 /me`.                         | Invalid credentials → `422` и neutral `Invalid login or password`; CSRF/store failure — safe error outcome, technical detail only in redacted log. |
| `POST /logout`   | Authenticated or stale/anonymous browser state; CSRF required by browser POST policy. | Invalidate own session if present, clear cookie, `303 /`.                    | Missing/stale session is idempotent: cookie is still cleared and response is `303 /`; no `GET /logout`.                                            |
| `GET /me`        | Authenticated current user only; no state change.                                     | `200`, minimal current-user identity.                                        | Missing/invalid/expired session → `303 /login`; no private-resource disclosure and no `500`.                                                       |

#### Что нужно сделать

- [x] Зафиксировать для каждого route method, path, purpose, actor, CSRF requirement и
  state-changing side effect в таблице выше.
- [x] Зафиксировать registration/login fields and validation: `first_name`, `login`, `email`,
  `password`, confirmation; safe normalization and accepted length limits находятся в password/
  validation contract.
- [x] Для каждого success path записать exact status/redirect, session creation/rotation и
  отсутствие session work для GET forms.
- [x] Для каждого failure path записать client-visible status/message/re-render: malformed or
  invalid input, duplicate identity, invalid credentials, CSRF failure, anonymous protected access и
  store failure.
- [x] Зафиксировать, что `next` отсутствует; если продукт позже потребует return-to, принимать
  только validated local paths с safe `/me` fallback, никогда arbitrary external URL.
- [x] Зафиксировать, что GET/HEAD не меняют state, а logout остаётся только CSRF-protected `POST`.

**Результат шага 2:** route/form/outcome table пригодна для реализации handlers и `httptest` без
угадывания status, redirect, CSRF boundary или session side effect; authenticated/anonymous outcomes
не смешаны с будущим API contract.

### Шаг 3 — ownership и protected-resource rules

В v0.2.5–v0.2.6 нет private library resource: `GET /me` защищает только identity текущего
пользователя. Поэтому ownership rule для library не реализуется искусственно, но boundary для
следующей private action фиксируется заранее.

| Protected boundary             | Anonymous                             | Authenticated current user                      | Non-owner / elevated actor                                                                                          |
|--------------------------------|---------------------------------------|-------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| `GET /me`                      | `303 /login`, без disclosure resource | `200`, только собственная minimal identity      | Не применимо: route не принимает target user/resource и не имеет admin bypass                                       |
| Будущая private library action | Предсказуемый MPA refusal             | Только owner, если operation разрешена contract | Non-owner получает generic refusal без disclosure; elevated actor появится только после отдельного product decision |

#### Что нужно сделать

- [x] Зафиксировать единственный текущий protected boundary: `GET /me` → current authenticated user
  only; missing/invalid/expired session → anonymous redirect, не `500`.
- [x] Разделить authentication и authorization: session middleware/current-user context доказывает
  identity, но future private action должна проверять owner/permission в use case или другом ясном
  server-side boundary.
- [x] Зафиксировать, что navigation visibility, hidden form fields, `is_admin` и database role не
  являются authorization control или bypass.
- [x] Не добавлять owner/non-owner implementation до появления `library_items` или другой private
  resource; это v0.3+ scope, а не placeholder feature v0.2.6.
- [x] Для будущей private action подготовить обязательную test triple: anonymous → refusal, owner →
  allowed operation, non-owner → refusal without private existence/detail disclosure. Admin/elevated
  case добавляется только отдельным product decision.
- [x] Зафиксировать MPA outcome convention: anonymous redirect на `/login`; authenticated non-owner
  получает generic refusal/forbidden page согласно будущему route contract, без API `401`/`403`
  payload semantics в этом release.

**Результат шага 3:** current-user rule для `/me` принят, owner/non-owner behavior для будущего
private resource явно deferred с test template и security boundary, а generic RBAC/admin bypass не
появляется без product scope.

### Шаг 4 — session и cookie lifecycle

Для v0.2.5 принят DB-backed opaque session. Cookie не является хранилищем identity или session
payload: она содержит только cryptographically random raw token, а database — только его
cryptographic hash и минимальные lifecycle fields.

| Состояние/событие                | Session state                                                                                      | Cookie/outcome                                                                        |
|----------------------------------|----------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------|
| Anonymous GET form/static/health | No authenticated session; no unnecessary store write.                                              | No auth cookie required; `/static/*` и `/healthz` остаются вне session middleware.    |
| Successful registration/login    | Create a new session; authentication always creates/rotates the ID.                                | `Set-Cookie` only after success; cookie contains raw opaque token, never user data.   |
| Authenticated request            | Load by token hash, verify not expired, attach minimal current user through typed request context. | Existing valid cookie remains; raw token не попадает в logs.                          |
| Missing/invalid/expired token    | Treat as anonymous, not as internal error.                                                         | Protected `/me` redirects to `/login`; invalid/expired session may be removed lazily. |
| Logout                           | Delete/invalidate own DB session; operation is idempotent for stale/missing session.               | Expire/clear `book_social_session` even when DB row is already absent; `303 /`.       |
| Expired records                  | Lazy cleanup on load/delete is sufficient for this release.                                        | No cleanup worker or background session task.                                         |

#### Что нужно сделать

- [x] Выбрать server-side store and manager criteria: opaque high-entropy token in cookie, hash-only
  lookup in DB, expiry, explicit invalidation и session-ID renewal/rotation.
- [x] Записать allowlist session contents: minimal `user_id`/current identity and one-time flash
  data; password, password hash, raw CSRF value, raw session token и private content запрещены.
- [x] Зафиксировать lifecycle sequence: anonymous → successful registration/login → new
  authenticated session → protected request → logout invalidation → refusal on token reuse.
- [x] Принять absolute lifetime 7 days без sliding renewal; missing, invalid и expired token
  означает anonymous state; отдельный cleanup worker не нужен.
- [x] Зафиксировать cookie policy: `book_social_session`, `HttpOnly`, `Path=/`, `SameSite=Lax`,
  `Secure` для stage/prod, explicit configuration exception for local development и clear-cookie
  behavior on logout.
- [x] Ограничить session load/save middleware dynamic routes, которые действительно используют
  state;
  `/static/*` и `/healthz` не создают session writes без отдельной причины.

**Результат шага 4:** lifecycle и cookie-policy table приняты; login/registration rotation, request
load, expiry/anonymous fallback, logout invalidation, clear-cookie behavior и lazy cleanup можно
проверить отдельно. Значения не копируются из Snippetbox автоматически и не оставляют `TBD`, которое
меняет security behavior.

### Шаг 5 — password-storage policy

Password policy относится к одному auth/password package и не размазывается по handlers, templates
или repositories.

| Boundary          | Принятое правило                                                                                                                                                                     |
|-------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Algorithm         | Использовать bcrypt как поддерживаемый adaptive hash; cost хранится в одном auth package и не дублируется по call sites.                                                             |
| Input             | Password — 12–128 символов; confirmation обязана совпадать. Complexity theatre не вводится. `first_name`, `login`, `email` нормализуются и валидируются по schema-compatible limits. |
| Persistence       | Хранится только adaptive hash; plaintext password никогда не записывается. Reversible encryption и собственное хеширование запрещены.                                                |
| Runtime boundary  | Plaintext существует только в form input и узком hash/verify path; password и hash не входят в DTO, page model, response, error, log, metric или fixture.                            |
| Login failure     | Unknown login/email и wrong password дают одинаковый neutral `Invalid login or password` и одинаковый внешний `422`; техническая причина — только в redacted log.                    |
| Deferred features | Password change/reset, recovery и activation не входят в v0.2.5–v0.2.6 без отдельного scope decision.                                                                                |

#### Что нужно сделать

- [x] Выбрать maintained Go password-hashing algorithm/dependency and cost policy: bcrypt, один auth
  package, cost с обоснованным operational budget.
- [x] Зафиксировать one-way rule: persist only a hash; custom hashing, reversible encryption,
  plaintext comparison и plaintext storage запрещены.
- [x] Описать boundary: password только в form input и narrow hashing/verification path; он
  отсутствует в user response DTO/template data, errors, logs, metrics и test fixtures.
- [x] Принять generic invalid-credentials response, чтобы клиент не узнавал наличие login/email;
  internal detail остаётся redacted и не попадает в client response.
- [x] Отложить password-change/reset, recovery и activation до отдельного accepted scope; не
  оставлять half-designed endpoint в текущем contract.

**Результат шага 5:** password policy reviewable до кода; persistence, runtime, errors и logs не
имеют неявного plaintext/hash path, а deferred password features явно отделены от auth foundation.

### Шаг 6 — abuse, error и logging policy

Auth errors разделяются на client-visible outcome и server diagnostic. Клиент получает стабильное
безопасное поведение; logger — только redacted operation context и request ID, без секретов.

| Failure/event                                     | Client-visible outcome                                             | Allowed server log fields                                                                             |
|---------------------------------------------------|--------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------|
| Invalid registration input                        | `422` с safe field errors; password не repopulate.                 | Request ID, route, validation/error class, field names without values.                                |
| Duplicate login/email                             | `422` с safe field error согласно registration contract.           | Request ID, operation, conflict class; no existing account details beyond required operational class. |
| Unknown login/email or wrong password             | Same neutral `422`, `Invalid login or password`.                   | Request ID, route, `invalid_credentials` class; never submitted login/email, password or hash.        |
| Missing/invalid/expired session                   | `/me` → `303 /login`; logout remains idempotent and clears cookie. | Request ID, route, anonymous/expired outcome class; no raw token or cookie header.                    |
| Missing/invalid CSRF                              | Generic safe MPA refusal/error; no token details.                  | Request ID, route, `csrf_rejected` class; never CSRF value or cookie contents.                        |
| Store/hash/unexpected internal failure            | Generic `500` body/page; no DB, stack or credential detail.        | Request ID, operation, typed internal error class and safe wrapped cause; no secrets.                 |
| Successful auth/logout or denied protected action | Normal redirect/refusal; no sensitive data in page model.          | Request ID, route, outcome class, safe actor state; no password, hash, token or private content.      |

#### Что нужно сделать

- [x] Разделить client message и server diagnostic для auth failures; client получает stable safe
  message, log — только redacted operation context, request ID и safe error class.
- [x] Создать logging denylist: passwords, password hashes, raw session IDs, CSRF tokens, cookie and
  authorization headers, submitted credentials и private resource content.
- [x] Определить допустимые event classes: registration validation/conflict, login success/failure
  class, session create/load/expire/invalidate, logout, denied protected action и unexpected server
  error — без sensitive values.
- [x] Назначить predictable outcome для store failure, hash failure и invalid/expired session;
  client не получает database, stack trace или credential detail.
- [x] Отметить controls, которые не включаются молча: API rate limiting и CORS относятся к 7B; login
  throttling/account lockout требуют отдельного product/abuse decision.

**Результат шага 6:** error-and-log matrix связывает каждую failure/event class с client outcome и
разрешёнными server fields; denylist и deferred abuse controls приняты до реализации handlers и
logging changes.

### Шаг 7 — собрать и review contract

Результаты шагов 0–6 собраны в этом документе и сверены перед schema, dependency и code changes.
Review проверяет именно согласованность решений, а не наличие уже реализованного auth behavior.

#### Собранный contract

- **Scope:** MPA registration, login, logout и protected `GET /me`; без profile, activation,
  recovery, API auth, RBAC routes, CORS, rate limiting, JWT/OpenAPI и v0.3 library resource.
- **Actors:** anonymous и authenticated user; admin/elevated actor deferred; `is_admin` не даёт
  bypass.
- **Routes/forms:** exact method/path/status/redirect/CSRF outcomes в route matrix; `next`
  отсутствует, GET/HEAD не меняют state.
- **Authorization:** `/me` доступен только current authenticated identity; future owner/non-owner
  action требует отдельного resource contract and test triple.
- **Session:** DB-backed opaque token, hash-only lookup, 7-day absolute lifetime, rotation after
  authentication, invalidation/clear cookie at logout, lazy expiry cleanup, no session work for
  static/health routes.
- **Password:** bcrypt, 12–128 characters, matching confirmation, adaptive hash only, no plaintext
  or hash disclosure, neutral invalid-credentials outcome.
- **Errors/logs:** stable safe client outcomes, generic internal errors, request-ID/error-class logs
  only, denylist for all credentials/tokens/private data; abuse controls outside accepted scope
  remain explicitly deferred.

#### Review checklist and outcome

- [x] Собрать actor/authorization matrix, route/form table, session/cookie policy, password policy,
  error/log matrix и test matrix в один reviewable contract.
- [x] Проверить open questions: accepted decisions записаны; deferred items имеют scope trigger;
  security behavior не оставлен как unowned `TBD`.
- [x] Сверить с Chapter 8 notes: session ID opaque, renewed at authentication, invalidated at
  logout, loaded/saved only where state is needed; Snippetbox MySQL schema и 12-hour lifetime не
  копируются автоматически.
- [x] Провести focused review до изменения code/schema/dependencies и отделить accepted scope от
  unresolved product authority.

**Результат шага 7:** один reviewable auth-and-authorization contract согласован с текущим MPA
`book-social` scope; v0.2.5 foundation может переходить к persistence/service work, а v0.2.6 UI
остаётся следующим этапом. Contract не утверждает, что behavior уже реализован.

Следующий полный контракт утверждён для первого v0.2.5–v0.2.6 slice и является ориентиром до
реализации. Он не описывает уже реализованное поведение.

### Scope и account

- Сохраняются modular monolith, layered boundaries, MPA, `html/template` и existing buffered
  renderer. Это не API-auth, roles/RBAC release, profile, activation, recovery или v0.3 personal
  library.
- Auth добавляется только к dynamic MPA routes; `/healthz` и `/static/*` не получают auth/session
  behavior. Не добавляются CORS, rate limiting, bearer tokens/JWT или OpenAPI.
- Registration принимает только `first_name`, `login`, `email`, `password` и password confirmation.
  Существующие необязательные `second_name` и `sur_name` не входят в форму.
- Используется существующая таблица `users`. Обычная роль `user` гарантируется migration/seed и
  назначается сервером; HTML form не передаёт роль. `is_admin` не открывает новых routes в этом
  release. Отсутствующая default role в corrupted/manual database является internal error, а не
  причиной назначить другую роль или admin access.

### Password, validation и errors

- Password hashing: bcrypt; policy и cost находятся в одном auth package. Хранится только adaptive
  hash. Password и hash не попадают в page models, responses или logs.
- Password length — от 12 до 128 символов. Password-complexity rules не вводятся; confirmation
  должна совпадать с password.
- `first_name`, `login` и `email` нормализуются trim-ом внешних пробелов; login и email сравниваются
  в нормализованной canonical форме. Максимальные длины совместимы с PostgreSQL schema:
  `first_name` — 100, `login` — 64, `email` — 254 символа.
- Domain boundary определяет typed errors: `ValidationError`, `ErrLoginTaken`, `ErrEmailTaken`,
  `ErrInvalidCredentials`, `ErrUnauthenticated` и internal error. Handler, а не repository/service,
  сопоставляет их с HTTP outcome (`422`, `303` или generic `500`); handler не сравнивает строки
  ошибок database driver.
- Registration показывает безопасные field errors для validation/duplicate conflicts и сохраняет для
  повторного отображения только безопасные значения, никогда password.
- Неизвестный login/email и неверный password на login имеют одинаковые status/redirect и
  нейтральный client-visible text `Invalid login or password`; техническая причина допустима только
  в redacted server log.

### Session и cookie

- Session — DB-backed opaque session. Cookie содержит только cryptographically random raw token;
  database хранит только его cryptographic hash, `user_id`, `created_at` и `expires_at`. Raw token
  не хранится в database или logs. `Set-Cookie` происходит только после успешной authentication.
- Имя cookie — `book_social_session`; атрибуты: `HttpOnly`, `Path=/`, `SameSite=Lax`.
  `Secure` включён для `stage` и `prod`; development exception задаётся конфигурацией и
  документируется, а не является неявным ослаблением production policy.
- Сессия имеет абсолютный срок жизни 7 дней без sliding renewal. Успешная authentication всегда
  создаёт новую сессию; missing, invalid или expired token означает anonymous state.
- Logout удаляет/invalidate DB session и очищает cookie даже при уже отсутствующей сессии. Expired
  records можно удалять lazy при load/delete; отдельный cleanup worker не нужен.

### Persistence и migration boundary

- Следующая numbered migration — `000003`; исходная v0.1 migration `000001` не переписывается.
- SQLite и PostgreSQL получают эквивалентную таблицу `sessions` с `id`, foreign key `user_id` на
  `users`, unique lookup по `token_hash`, `created_at`, `expires_at` и индексом `expires_at` для
  cleanup. Timestamp semantics и constraints одинаковы в обеих database implementations.
- Migration имеет обратимый `down`. Seed/reset и test bootstrap адаптируются так, чтобы ordinary
  role `user` была доступна и disposable database проходила migration/seed smoke.

### Transactions, identity и CSRF

- Service/use case владеет transaction, когда меняется более одной таблицы. Registration атомарно
  создаёт user, получает server-owned default role и создаёт session; successful login создаёт новую
  session. Одиночный delete/invalidate session может быть repository call.
- Current-user middleware загружает minimal identity через typed request context. Отсутствующая
  сессия — нормальное anonymous state, не 500; query/header identity не доверяется.
- Один server-side CSRF mechanism/token защищает все browser `POST` routes. GET/HEAD не меняют
  state; CSRF value передаётся только в forms и не попадает в logs.
- Auth middleware и CSRF добавляются с явным порядком относительно существующей цепочки
  `SecurityHeaders -> RequestID -> TrustedRealIP -> RequestLogger -> Recoverer`; dynamic MPA routes
  сохраняют application timeout.

### Route/form contract

| Route            | Contract                                                                                                                                                                               |
|------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `GET /register`  | Возвращает accessible form (`200`); authenticated user получает `303` на `/me`.                                                                                                        |
| `POST /register` | Проверяет CSRF и input. При успехе атомарно создаёт user и session, затем auto-login и `303` на `/me`; при validation/conflict возвращает `422` с безопасными field errors.            |
| `GET /login`     | Возвращает form (`200`); authenticated user получает `303` на `/me`.                                                                                                                   |
| `POST /login`    | Проверяет CSRF и credentials. При успехе создаёт новую session и возвращает `303` на `/me`; при invalid credentials возвращает `422` с единым нейтральным сообщением.                  |
| `POST /logout`   | Проверяет CSRF, invalidates session, очищает cookie и возвращает `303` на `/`. Нет state-changing `GET /logout`; anonymous/stale session также очищает cookie и получает `303` на `/`. |
| `GET /me`        | Protected minimal page с identity пользователя (`200`); anonymous получает `303` на `/login`.                                                                                          |

- Первый slice не принимает параметр `next`: successful registration и login всегда redirect-ят на
  `/me`.
- Navigation позднее показывает Login/Register anonymous visitor и `/me` с Logout form authenticated
  visitor; навигация не является authorization control.

### Порядок реализации и evidence

1. Сначала `000003` migrations, seed и test bootstrap.
2. Затем `internal/modules/users`: user/session models, repository interfaces, password policy и
   service/use cases с unit tests.
3. После этого — отдельные SQLite/PostgreSQL repositories и их contract/parity tests.
4. Затем HTTP boundary: cookie manager, current-user middleware, CSRF и protected-route guard.
5. Только после v0.2.5 core — v0.2.6 handlers, templates, navigation и flash messages.
6. Проверка идёт от narrow package tests к `GOCACHE=/tmp/book-social-go-cache make test`,
   `GOCACHE=/tmp/book-social-go-cache go vet ./...`, `git diff --check` и
   `git status --short --branch`. HTTP behavior проверяется через `httptest`, не real listener;
   PostgreSQL parity запускается отдельно на disposable DSN и сообщается как environment-dependent
   evidence.

### Definition of Done v0.2.5

- Auth core создаёт user с password hash, verifies credentials, создаёт/загружает/удаляет DB
  sessions и устанавливает current-user request context.
- Raw passwords, hashes, raw session IDs и CSRF values отсутствуют в responses и structured logs.
- SQLite и PostgreSQL имеют эквивалентную migration story; disposable database проходит
  migration/seed smoke.
- Focused unit, repository и `httptest` tests покрывают happy и refusal paths без real listener.
- Краткая documentation фиксирует password/session policy, lifecycle, configuration и difference
  между development и HTTPS stage/prod.
