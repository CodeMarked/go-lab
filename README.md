# go-lab

Go API, Angular SPA, MySQL 8.4, golang-migrate, Docker Compose. Code: [`go_CRUD_api/`](go_CRUD_api/), [`client/`](client/), [`migrations/`](migrations/).

**Need:** Docker + Compose + Git. **Go 1.24+** only for local `go test` / `go run` (CI and the backend image match `go_CRUD_api/go.mod`).

## First-time setup

1. `cp .env.example .env` (PowerShell: `Copy-Item .env.example .env`) — set long `JWT_SECRET`, `PLATFORM_CLIENT_*`.
2. `docker compose up -d --build`
3. `docker compose run --rm migrate` (any fresh DB or new migration)
4. **UI** [localhost:4200](http://localhost:4200) · **API** [localhost:5000](http://localhost:5000) · **health** `/healthz` · **ready** `/readyz` · MySQL `localhost:3306`

Schema comes **only** from [`migrations/`](migrations/), not app startup.

### Admin SPA (Angular)

- **Cookie auth (default):** `useBootstrapAuth: false` in `client/src/environments/environment.ts` → Register → Sign in. Mutations need **CSRF** header (`X-CSRF-Token` = `gl_csrf` cookie) and **`withCredentials`**. Periodic **`POST /api/v1/auth/refresh`**; **401** → sign-in again.
- **Bootstrap JWT (dev):** `useBootstrapAuth: true` calls `POST /api/v1/auth/bootstrap`. Turn off for prod-style deploys — [docs/bootstrap-sunset.md](docs/bootstrap-sunset.md).
- **Same origin for cookies:** different ports = different sites, so use **`apiBaseUrl: ''`** and proxy `/api` on the SPA host ([`docker/frontend.nginx.conf`](docker/frontend.nginx.conf), [`client/proxy.conf.json`](client/proxy.conf.json)). Full URL to `:5000` only if you mean to use **Bearer**, not cookie session.

## Daily use

- Up: `docker compose up -d` (add `--build` when images change)
- Migrate: `docker compose run --rm migrate` (backend waits for migrate on full stack)
- **500 on register:** logs `auth_register_db_error` — usually missing migrations or wrong **`MIGRATION_EXPECTED_VERSION`** ([`.env.example`](.env.example))
- **Schema out of sync** (`no change` but missing columns): compare `users` columns vs `schema_migrations`; repair with care — [docs/migrations.md](docs/migrations.md)
- Down: `docker compose down` · wipe DB: `docker compose down -v`
- Logs: `docker compose logs backend --tail 100`

## Configuration

Full list: [`.env.example`](.env.example). Common:

| Var | Note |
|-----|------|
| `DB_*` | DSN uses `parseTime=true` — required for sessions |
| `JWT_SECRET` | ≥32 chars prod; optional `JWT_SECRET_PREVIOUS` — [jwt-rotation.md](docs/jwt-rotation.md) |
| `SESSION_*` | HttpOnly session cookie; **`SESSION_COOKIE_SECURE=false`** on plain HTTP (see `.env.example` and Compose) or browsers drop cookies → 401 |
| `CSRF_*` | Double-submit for cookie mutations — [auth-session.md](docs/auth-session.md) |
| `PLATFORM_CLIENT_*` | `/auth/token` + bootstrap |
| `MIGRATION_EXPECTED_VERSION` | Optional; `/readyz` checks version |
| `OIDC_ISSUER_URL` + `OIDC_AUDIENCE` | Both or neither; RS256 Bearer — [oidc-auth0.md](docs/oidc-auth0.md) |
| `REDIS_URL` | Optional shared rate limits / lockout; `docker compose --profile redis` |
| `CORS_ALLOWED_ORIGINS` | No `*` |

## API (`/api/v1`)

| Area | Endpoints |
|------|-----------|
| Health | `GET /healthz`, `GET /readyz` |
| Auth | `POST .../auth/register`, `login`, `logout`, `refresh` (cookie session) |
| Service | `POST .../auth/token`, `.../auth/bootstrap` (temporary bridge) |
| Users | CRUD + `/users/search` — **session cookie or Bearer JWT** |

Responses: `{ data, meta }` or `{ error, meta }`. Old `/api/...` outside v1 → **410**. SPA config: [`client/src/environments/environment.ts`](client/src/environments/environment.ts).

**Docs:** [docs/README.md](docs/README.md) · **Roadmap / suite:** [docs/MASTER_PLAN.md](docs/MASTER_PLAN.md)

## Tests and CI

- `cd go_CRUD_api && go test ./...`
- Smoke: `./scripts/test.ps1` (stack up) · `./scripts/migrate.ps1`
- [`.github/workflows/ci.yml`](.github/workflows/ci.yml) — tests + Compose smoke; cache via [`docker-compose.ci.yml`](docker-compose.ci.yml)

## Production

Strong secrets; TLS in front; **`SESSION_COOKIE_SECURE=true`**; private MySQL; migrate before new replicas; tight CORS; **`AUTH_BOOTSTRAP_ENABLED=false`** when unused. **Multi-replica:** set **`REDIS_URL`** (or edge limits) so rate limits / lockout aren’t per-process only — [auth-session.md](docs/auth-session.md).

## More

- Migrations: [docs/migrations.md](docs/migrations.md)
- Docker Desktop must be running on Windows
