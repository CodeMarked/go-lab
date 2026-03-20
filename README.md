# go-lab

Platform API baseline for Marble and other products: **versioned JSON API**, **JWT access tokens**, **MySQL + golang-migrate**, **Docker Compose**, and **CI smoke tests**.

- Angular SPA in [`client/`](client/)
- Go API in [`go_CRUD_api/`](go_CRUD_api/)
- SQL migrations in [`migrations/`](migrations/)

## Architecture

| Component   | Role |
|------------|------|
| `frontend` | Built Angular app on `http://localhost:4200` |
| `backend`  | Go API on `http://localhost:5000` |
| `mysql`    | MySQL 8.4 on `localhost:3306` |
| `migrate`  | One-shot job: applies [`migrations/`](migrations/) |

The API **does not** create databases or tables at runtime; only the migrate job applies schema.

## Requirements

- Docker Desktop (or Docker Engine + Compose)
- Git
- (Optional) Go 1.23+ for local `go test` without Docker

## Quick start

```powershell
Copy-Item .env.example .env
# Edit .env: set JWT_SECRET (≥32 chars), PLATFORM_CLIENT_ID / PLATFORM_CLIENT_SECRET, MIGRATION_EXPECTED_VERSION
docker compose up -d --build
docker compose run --rm migrate
```

Open:

- SPA: `http://localhost:4200`
- Health: `http://localhost:5000/healthz`
- Readiness (DB + optional migration version): `http://localhost:5000/readyz`

## Configuration (environment only)

All backend settings come from the environment (Compose loads [`.env.example`](.env.example) → `.env`).

| Variable | Purpose |
|----------|---------|
| `DB_*` | MySQL connection |
| `JWT_SECRET` | HS256 signing key (**≥32 characters**, rotate in production) |
| `JWT_ISSUER`, `JWT_AUDIENCE` | JWT validation claims |
| `JWT_ACCESS_TTL_SECONDS` | Access token lifetime (60–86400) |
| `PLATFORM_CLIENT_ID`, `PLATFORM_CLIENT_SECRET` | Bootstrap `client_credentials` for `POST /api/v1/auth/token` |
| `MIGRATION_EXPECTED_VERSION` | If set, `/readyz` requires `schema_migrations.version` ≥ this and `dirty = 0` |
| `CORS_ALLOWED_ORIGINS` | Comma-separated browser origins (no wildcard) |
| `GIN_MODE` | `release` (default) or `debug` |

## API surface (`/api/v1`)

### Health

- `GET /healthz` — process up
- `GET /readyz` — DB reachable; optional migration version check

### Auth

- `POST /api/v1/auth/token` — body: `grant_type`, `client_id`, `client_secret` (`client_credentials`). Rate-limited per IP.

### Users (requires `Authorization: Bearer <access_token>`)

- `GET /api/v1/users`
- `GET /api/v1/users/search?name=`
- `GET /api/v1/users/:id`
- `POST /api/v1/users`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id` — **204** empty body

### JSON envelopes

Success (except `204`):

```json
{ "data": <T>, "meta": { "request_id": "..." } }
```

Error:

```json
{ "error": { "code": "...", "message": "...", "details": {} }, "meta": { "request_id": "..." } }
```

Unversioned `/api/...` (except `/api/v1`) returns **410 Gone** with the error envelope.

### Stable error codes (examples)

`VALIDATION_ERROR`, `NOT_FOUND`, `UNAUTHORIZED`, `INTERNAL_ERROR`, `RATE_LIMITED`

## Local Angular + API

[`client/src/environments/environment.ts`](client/src/environments/environment.ts) does not store platform credentials. The SPA obtains a bootstrap token from `POST /api/v1/auth/bootstrap` and then uses Bearer auth for API calls.

## Daily commands

```bash
docker compose up -d --build
docker compose run --rm migrate
docker compose down
```

## Testing

Backend unit tests:

```powershell
cd go_CRUD_api
go test ./...
```

Smoke tests (requires running stack + migrations):

```powershell
./scripts/test.ps1
```

Optional: `./scripts/migrate.ps1`

## CI

Workflow: [`.github/workflows/ci.yml`](.github/workflows/ci.yml)

- `go test ./...` in `go_CRUD_api`
- Docker Compose up, `migrate up`, wait for `/readyz`, `./scripts/test.ps1`

## Operations checklist (self-host)

1. Generate a strong `JWT_SECRET` (≥32 random bytes, store in secrets manager).
2. Set unique `PLATFORM_CLIENT_*` values; rotate if leaked.
3. Terminate TLS at a reverse proxy; do not expose MySQL publicly.
4. Set `MIGRATION_EXPECTED_VERSION` to the latest migration after each deploy.
5. Run `migrate up` before rolling new app instances.
6. Restrict `CORS_ALLOWED_ORIGINS` to real front-end origins.
7. Use structured logs from the API (JSON via `slog`) in your log aggregator.

## Migrations

See [`docs/migrations.md`](docs/migrations.md).

## Troubleshooting

- Docker unavailable on Windows (`open //./pipe/docker_engine`): start Docker Desktop.
- Backend errors: `docker compose logs backend --tail 100`
- `readyz` not ready: run migrations; check `MIGRATION_EXPECTED_VERSION` vs applied version; inspect `schema_migrations` for `dirty=1`.
