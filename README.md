# LedgerLine Backend

A Go backend template built on **Clean Architecture** (repository + service pattern), Gin, PostgreSQL (GORM), JWT, and Swagger.

- Go 1.25 · Gin · GORM · golang-migrate · JWT (HS256) · slog · swaggo
- Every cross-layer dependency goes through an **interface** declared in `internal/domain`

## Folder structure

```
.
├── cmd/api/                     # entrypoint: read config, wire dependencies, start the server
├── internal/
│   ├── config/                  # load & validate environment variables
│   ├── database/                # PostgreSQL connection & pool setup (GORM)
│   ├── domain/                  # entities + interfaces (ports) + typed errors + error codes  <- the core, no framework dependencies
│   ├── repository/postgres/     # repository implementation (GORM) + tx manager + error mapping
│   │   └── model/               # GORM persistence models: xxx_model.go -> XxxModel + mappers
│   ├── service/                 # business logic
│   ├── delivery/http/
│   │   ├── apierr/              # the single error -> HTTP response writer (status + envelope)
│   │   ├── dto/                 # request/response DTOs (XxxDTO) + mappers to the domain
│   │   ├── handler/             # HTTP adapters + swagger annotations
│   │   ├── middleware/          # request id, logger, recovery, cors, auth, rate limit, timeout
│   │   └── router/              # route registration
│   ├── server/                  # http.Server + graceful shutdown
│   └── mocks/                   # domain interface mocks for unit tests
├── pkg/                         # reusable utilities: logger, response, validator, jwt, hash, mailer, pagination
├── db/migrations/               # SQL migrations (golang-migrate)
├── docs/                        # generated swagger output (do not edit by hand)
└── .github/workflows/           # CI: vet, test, lint
```

### Dependency direction rules

```
handler  ->  service  ->  repository
    \           |            /
     ----->  domain  <-------      (domain depends on nobody)
```

- `domain` holds only entities + interfaces. It must never import gin, GORM, or anything external — table mapping is handled by the persistence models in the repository layer.
- `service` receives a `domain.UserRepository` (an interface), not a Postgres struct — so it is easy to mock in tests and the database can be swapped without touching business logic.
- `handler` never touches SQL; `repository` never knows about HTTP status codes.

## Running

```bash
cp .env.example .env          # then change JWT_SECRET
make setup                    # install swag, migrate, golangci-lint, air
docker compose up -d postgres # or use a local Postgres
make migrate-up
make run                      # or: make dev (hot reload)
```

The whole stack at once:

```bash
make docker-up
```

## Main commands

| Command | Purpose |
|---|---|
| `make run` / `make dev` | run the app (plain / hot reload) |
| `make test` | unit tests + race detector |
| `make test-cover` | coverage report to `coverage.html` |
| `make lint` / `make fmt` | golangci-lint / gofmt + go vet |
| `make swag` | regenerate the Swagger documentation |
| `make migrate-up` / `make migrate-down` | migrate up / roll back 1 version |
| `make migrate-create name=create_xxx_table` | create a new migration file |

## Swagger

Once the app is running: <http://localhost:8080/swagger/index.html> (automatically disabled when `APP_ENV=production`).

Regenerate it every time a handler annotation changes: `make swag`.

For endpoints marked `@Security BearerAuth`, click **Authorize** and enter `Bearer <access_token>`.

## Built-in endpoints

| Method | Path | Access |
|---|---|---|
| GET | `/health`, `/health/ready` | public |
| POST | `/api/v1/auth/register` | public |
| POST | `/api/v1/auth/login` | public |
| POST | `/api/v1/auth/refresh` | public |
| POST | `/api/v1/auth/forgot-password`, `/verify-otp`, `/reset-password` | public |
| GET | `/api/v1/auth/me` | authenticated |
| GET | `/api/v1/users`, `/api/v1/users/:id` | authenticated |
| POST/PATCH/DELETE | `/api/v1/users`, `/api/v1/users/:id` | authenticated (see the TODO in `router.go`: admin-only is not wired up yet) |
| CRUD | `/api/v1/categories` | authenticated (per-user data) |
| CRUD | `/api/v1/roles`, `/api/v1/roles/:id` | authenticated |
| GET | `/api/v1/audit-logs`, `/overview`, `/options`, `/export` | authenticated |

`categories` is a **complete example module** worth copying for new features — see [PANDUAN.md](PANDUAN.md).

## Response format

Success:

```json
{ "success": true, "message": "success", "data": { }, "meta": { "page": 1, "per_page": 10, "total_items": 42, "total_pages": 5 } }
```

Failure:

```json
{ "success": false, "message": "the submitted data is invalid", "code": "VALIDATION_ERROR",
  "errors": [{ "field": "email", "message": "email must be a valid email address" }],
  "request_id": "9b2c1f2e-6f0a-4a1e-9c7d-2f8b0a1c3d4e" }
```

`code` is the contract the frontend switches on; `message` is for humans and must never be parsed. The full catalogue is in [ERROR_CODES.md](ERROR_CODES.md).

A service returns a domain error and says nothing about HTTP:

```go
return nil, domain.Conflict(domain.CodeUserEmailTaken, "email is already registered").WithField("email")
```

[internal/delivery/http/apierr](internal/delivery/http/apierr/apierr.go) is the single place an error becomes a response — handlers and middleware both go through it, so every failure leaves the same envelope. `Kind` picks the status, `Code` and `Message` go to the client, and the wrapped cause goes to the log only.

## Logging

Uses `log/slog` (stdlib), JSON format in production, switchable via `LOG_FORMAT=text` during development.

- The `RequestID` middleware creates an `X-Request-ID` per request; every error body repeats it as `request_id`.
- The `Logger` middleware records method, path, status, latency, IP, and injects a `request_id`-tagged logger into the `context`.
- From any layer: `logger.FromContext(ctx).Info("message", slog.String("key", "value"))` — the log automatically carries the same `request_id`.

## Testing

```bash
make test        # all unit tests
make test-cover  # + coverage
```

Patterns in use:

- **Table-driven tests** + `testify` (`require` for fatal, `assert` to continue).
- Domain interface mocks live in `internal/mocks` — services are tested without a database at all.
- Handlers are tested with `httptest` + mock services (see [user_handler_test.go](internal/delivery/http/handler/user_handler_test.go)).
- `t.Parallel()` in every test, with `-race` enabled in both `make test` and CI.

## Security measures already in place

- bcrypt passwords; a failed login always answers `ErrInvalidCredentials` (anti user enumeration).
- Short-lived access token (15m) + refresh token; a refresh token cannot be used as an access token.
- Per-IP rate limiting, security headers, configured CORS, request timeout.
- Soft delete + partial unique index; every category query is filtered by `user_id`.
- Queries go through GORM (automatic parameter binding) — no SQL string concatenation.
