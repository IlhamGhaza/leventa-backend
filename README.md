# Leventa

Backend API for **Leventa**, a role-based property management platform. Internal system: guests (no account) browse the public catalog and contact Marketing via WhatsApp. Marketing (sales) users create and manage listings. Admins manage co-listing, ownership, deals and commission.

**Stack:** Go · Fiber v3 · PostgreSQL · GORM · JWT · golang-migrate · OpenAPI/Swagger. It is a modular monolith. It needs no Redis, no Docker and no other external service.

## Features

- Email/password auth with JWT access and refresh tokens (refresh rotation, revocation, login lockout, rate limiting). There is no role selector at login.
- Roles: `sales` (shown as _Marketing_) and `admin`. Customers are guests without an account.
- Property listings with a primary marketing owner and up to two co-listing marketing users.
- Private owner data and exact addresses are returned only to authorized actors.
- Deals and cancellations recorded by Admin (`POST /properties/{id}/deal`, `/cancel`) and commissions.
- Image uploads (local filesystem).

## Prerequisites

| Tool       | Version                                  |
| ---------- | ---------------------------------------- |
| Go         | 1.26 or newer (see `go.mod`)             |
| PostgreSQL | 14 or newer, running locally or remotely |

## Quick start

```bash
# 1. Clone (or fork, then clone your fork)
git clone https://github.com/IlhamGhaza/leventa-backend
cd leventa-backend

# 2. Create a database
createdb propertydb
# or: psql -U postgres -c "CREATE DATABASE propertydb;"

# 3. Configure environment
cp .env.example .env          # Windows PowerShell: Copy-Item .env.example .env
# Edit .env: set DATABASE_URL (user/password) and a new JWT_SECRET

# 4. Apply migrations
go run ./cmd/migrate -up

# 5. Start the server and load demo data (first run only)
go run ./cmd/server -seed

# Next runs
go run ./cmd/server
```

Check that it works:

```bash
curl http://localhost:8080/health
```

- Swagger UI: <http://localhost:8080/swagger>
- OpenAPI spec: <http://localhost:8080/openapi.yaml>

> `-seed` is refused when `APP_ENV=production`. It skips tables that already have data, so re-running is safe. To start over, run `go run ./cmd/migrate -down`, then `-up`, then `-seed`.

## Demo accounts

Created by `-seed`. **For local development only.**

| Role      | Email                   | Password      |
| --------- | ----------------------- | ------------- |
| Admin     | `admin@leventa.com`     | `admin123`    |
| Marketing | `sales@leventa.com`     | `sales123`    |
| Marketing | `sales2@leventa.com`    | `sales123`    |
| Marketing | `sales3@leventa.com`    | `sales123`    |

Login:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"sales@leventa.com","password":"sales123"}'
```

Send the returned access token as `Authorization: Bearer <token>`. All routes live under `/api/v1`.

## Configuration

All settings come from environment variables. A `.env` file in the working directory is loaded automatically.

| Variable                                 | Default in `.env.example`                                                | Notes                                                            |
| ---------------------------------------- | ------------------------------------------------------------------------ | ---------------------------------------------------------------- |
| `APP_ENV`                                | `local`                                                                  | `production` enforces `JWT_SECRET` ≥ 32 chars and blocks `-seed` |
| `APP_HOST`                               | empty                                                                    | Bind address; empty = all interfaces (0.0.0.0)                   |
| `APP_PORT`                               | `8080`                                                                   | HTTP port                                                        |
| `DATABASE_URL`                           | `postgres://postgres:password@localhost:5432/propertydb?sslmode=disable` | **Required**                                                     |
| `JWT_SECRET`                             | placeholder                                                              | **Required.** Change it                                          |
| `JWT_ACCESS_EXPIRATION`                  | `900`                                                                    | Seconds                                                          |
| `JWT_REFRESH_EXPIRATION`                 | `604800`                                                                 | Seconds                                                          |
| `LOG_FILE`                               | empty                                                                    | Optional JSON log file in addition to stdout                     |
| `DOCUMENT_PATH`                          | `./documents`                                                            | Private ownership documents; not served statically               |
| `UPLOAD_PATH`                            | `./uploads`                                                              | Local upload directory                                           |
| `UPLOAD_MAX_SIZE`                        | `5242880`                                                                | Bytes (5 MB)                                                     |
| `API_RATE_LIMIT_MAX` / `API_RATE_LIMIT_WINDOW` | `300` / `60` | Every `/api/v1` request, per IP |
| `UPLOAD_RATE_LIMIT_MAX` / `UPLOAD_RATE_LIMIT_WINDOW` | `30` / `60` | Image and document uploads, per IP |
| `RATE_LIMIT_MAX` / `RATE_LIMIT_WINDOW`   | `5` / `60`                                                               | Auth rate limit (requests per seconds)                           |
| `CORS_ALLOWED_ORIGINS`                   | empty                                                                    | Comma-separated origins. Empty = no cross-origin browser access  |
| `COMMISSION_DEFAULT_PERCENTAGE`          | `2.5`                                                                    | 0–100                                                            |
| `MAX_LOGIN_ATTEMPTS` / `LOCKOUT_MINUTES` | `5` / `15`                                                               | Brute-force protection                                           |
| `MAX_ACTIVE_SESSIONS`                    | `5`                                                                      | Per user                                                         |
| `TRUSTED_PROXIES`                        | empty                                                                    | Reverse-proxy IPs/CIDRs                                          |

Generate a secret: `openssl rand -hex 32`.

## Logs and troubleshooting

- Each request is logged with `method`, `path`, `status`, `latency`, `request_id` and, when logged in, `user_id` and `role`.
- Every response has an `X-Request-ID` header. When something fails, ask for that id and search the log for it: the 5xx line shows the real cause in the `error` field. Users only see "Internal server error".
- Set `LOG_FILE=./logs/app.log` in `.env` to also keep a JSON log file (it is git-ignored; rotate it with your OS tools).
- Business actions (listing changes, deals, commissions, document access, user changes) are in the `audit_logs` table; Admin reads them with `GET /api/v1/audit-logs`.
- Never share logs publicly without checking them first: they contain user IDs and IP addresses.

## Migrations

```bash
go run ./cmd/migrate -up              # apply all pending
go run ./cmd/migrate -steps 1         # apply one
go run ./cmd/migrate -steps -1        # roll back one
go run ./cmd/migrate -down            # roll back everything
```

Migration files are in [database/migrations](database/migrations). Create new ones as `NNNNNN_name.up.sql` and `NNNNNN_name.down.sql`.

## Tests

```bash
gofmt -l .
go vet ./...
go test ./...
```

Integration tests need a **separate, disposable** database with migrations applied. They are skipped when `TEST_DATABASE_URL` is not set.

```bash
# bash
TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/propertydb_test?sslmode=disable" \
  go test ./internal/usecase/...
```

```powershell
# PowerShell
$env:TEST_DATABASE_URL = "postgres://postgres:password@localhost:5432/propertydb_test?sslmode=disable"
go test ./internal/usecase/...
```

Do not point `TEST_DATABASE_URL` at a database with data you care about.

## API documentation and Postman

- [docs/openapi.yaml](docs/openapi.yaml): source of truth, served at `/openapi.yaml` and `/swagger`.
- [docs/postman_collection.json](docs/postman_collection.json) and [docs/postman_environment.json](docs/postman_environment.json): import into Postman.
- Regenerate Postman files after changing the spec: `pip install pyyaml && python scripts/gen_postman.py`.

## Project layout

```text
cmd/
  server/          HTTP server entrypoint (-seed flag)
  migrate/         Migration runner
database/
  migrations/      SQL migrations (golang-migrate)
  seeds/           Demo data
docs/              Specs, decisions and OpenAPI
internal/
  config/          Environment config
  delivery/http/   Handlers, middleware
  domain/          Entities and repository interfaces
  usecase/         Business rules and authorization
  repository/      PostgreSQL implementations
  routes/          Route wiring
  service/         JWT, notifications
  storage/         Upload storage (local filesystem)
pkg/               errs, logger, response, validator
scripts/           Tooling (Postman generator)
```

Request flow: `Handler → Use Case → Repository interface → PostgreSQL repository`.

## Design docs

Read these before changing behavior:

- [docs/PRODUCT_REQUIREMENTS.md](docs/PRODUCT_REQUIREMENTS.md)
- [docs/BUSINESS_RULES.md](docs/BUSINESS_RULES.md)
- [docs/AUTHORIZATION.md](docs/AUTHORIZATION.md)
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- [docs/DATABASE.md](docs/DATABASE.md)
- [docs/API_CONVENTIONS.md](docs/API_CONVENTIONS.md)
- [docs/DECISIONS.md](docs/DECISIONS.md)
- [docs/TESTING.md](docs/TESTING.md)

## Ground rules for contributors

- Actor identity and role come only from the verified JWT. Never trust `user_id`, role, owner ID or marketing ID from the request body.
- Enforce authorization in use cases and repositories, not only in route middleware.
- Owner contact data and exact address stay hidden from co-listing marketing, other marketing users and the public.
- Do not add Redis, Docker or microservices.
- When you change API behavior, update `docs/openapi.yaml` and add tests.

## Troubleshooting

| Problem                                | Fix                                                                        |
| -------------------------------------- | -------------------------------------------------------------------------- |
| `DATABASE_URL is required`             | Create `.env` from `.env.example`, or export the variable                  |
| `JWT_SECRET is required`               | Set `JWT_SECRET` in `.env`                                                 |
| Connection refused / auth failed       | Check PostgreSQL is running and `DATABASE_URL` credentials are correct     |
| `database "propertydb" does not exist` | Run `createdb propertydb`                                                  |
| Migration is marked dirty              | Fix the failing SQL, then reset the database or repair `schema_migrations` |
| Seed does nothing                      | Tables already have data. Reset as described above                         |
| Uploads fail                           | Make sure `UPLOAD_PATH` exists and is writable                             |
| Login returns 429                      | Auth rate limit hit. Wait `RATE_LIMIT_WINDOW` seconds                      |
| Any request returns 429                | Per-IP API limit hit (`Retry-After` header). Behind a proxy set `TRUSTED_PROXIES`, or raise `API_RATE_LIMIT_MAX` |

## License

Leventa is **source-available, not open source** in the OSI sense. It is licensed under the [PolyForm Noncommercial License 1.0.0](LICENSE).

- Allowed: read, fork, run, modify and share for noncommercial purposes. This covers personal study, research, hobby projects, and use by educational, charitable and government organizations.
- Not allowed: selling it, offering it as a paid product or service, or using it in a commercial business, without a separate written license from the author.
- When you share a copy, keep the `LICENSE` file and its `Required Notice:` line.

For commercial licensing, contact the author.
