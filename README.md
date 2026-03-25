# go-lab

Go API, Angular SPA, MySQL 8.4, and golang-migrate, orchestrated with Docker Compose. Sources: [`go_CRUD_api/`](go_CRUD_api/), [`client/`](client/), [`migrations/`](migrations/).

**Prerequisites:** Docker with Compose and Git.

**Go on your machine is optional.** Image builds only run `go build` (not tests). CI runs `go test ./...` on GitHub. Install Go 1.23+ if you want local tests or `go run` without Docker.

## First-time setup

1. Copy [`.env.example`](.env.example) to `.env` and set secrets (e.g. long `JWT_SECRET`; `PLATFORM_CLIENT_ID` / `PLATFORM_CLIENT_SECRET` for auth).

   ```bash
   cp .env.example .env
   ```

   PowerShell: `Copy-Item .env.example .env`

2. Start the stack:

   ```bash
   docker compose up -d --build
   ```

3. Apply schema (every fresh DB, and after new migration files):

   ```bash
   docker compose run --rm migrate
   ```

4. **App** [http://localhost:4200](http://localhost:4200) · **API** [http://localhost:5000](http://localhost:5000) · **health** [http://localhost:5000/healthz](http://localhost:5000/healthz) · **readiness** [http://localhost:5000/readyz](http://localhost:5000/readyz) · MySQL `localhost:3306`

Tables are not created by the app at startup—only the `migrate` job applies [`migrations/`](migrations/).

### Platform admin (Angular)

- Default auth is **cookie login** (`client/src/environments/environment.ts`: `useBootstrapAuth: false`). Open [http://localhost:4200](http://localhost:4200) → **Register** (first user) → **Sign in**. Mutating API calls send **CSRF** via `X-CSRF-Token` (read from `gl_csrf` cookie); all API calls use **`withCredentials`**. While signed in, the SPA periodically calls **`POST /auth/refresh`** (`sessionRefreshIntervalMs`, keep below `SESSION_IDLE_TTL_SECONDS`). A **401** on protected API calls clears local state and sends the user to sign-in again.
- Dev-only **bootstrap JWT** bridge: set `useBootstrapAuth: true` to call `POST /api/v1/auth/bootstrap` on startup (no register/login UI needed). Production-style deploys should keep it **false** ([`docs/bootstrap-sunset.md`](docs/bootstrap-sunset.md)).
- **Same-origin API (required for cookie sessions):** Browsers treat `localhost:4200` and `localhost:5000` as **different sites**, so `SameSite=Lax` session cookies are **not** sent on XHR to port 5000. The app defaults to **`apiBaseUrl: ''`** and calls **`/api/v1/...`** on the **same host as the SPA**. Docker: [`docker/frontend.nginx.conf`](docker/frontend.nginx.conf) proxies **`/api/`**, **`/healthz`**, **`/readyz`** to the `backend` service. Local: `ng serve` uses [`client/proxy.conf.json`](client/proxy.conf.json). Only set a full `apiBaseUrl` (e.g. `http://localhost:5000`) if you intentionally use **bootstrap Bearer** auth, not cookie login.
- **Docker UI (`frontend` service, port 4200):** open [http://localhost:4200/](http://localhost:4200/) — API traffic stays on that origin via nginx. You can still hit the backend directly at [http://localhost:5000](http://localhost:5000) for **`/healthz`** / **`/readyz`** / debugging.

## Daily use

- **Start (rebuild images):** `docker compose up -d --build`
- **Start (reuse images):** `docker compose up -d`
- **Migrations:** `docker compose run --rm migrate` (on full stack, `backend` waits for the `migrate` service to finish successfully so the API does not start on an empty schema).
- **Register returns 500:** Check backend logs for `auth_register_db_error`. Usually **migrations not applied** (`Unknown column 'email'` / missing `auth_*` tables) or **`MIGRATION_EXPECTED_VERSION` unset** so `/readyz` skipped the migration check while the DB was still on an old revision—run migrate and set `MIGRATION_EXPECTED_VERSION` in `.env` (see [`.env.example`](.env.example)).
- **`migrate up` prints `no change` but `Unknown column 'email'`:** The `schema_migrations` row can be ahead of the real schema (e.g. `force` or a bad copy). Check with `SHOW COLUMNS FROM users;` vs `SELECT version, dirty FROM schema_migrations;`. If version is `2` but `email` is missing, repair with: `docker compose run --rm migrate force 1` then `docker compose run --rm migrate up` (review data impact before doing this on non-dev DBs).
- **Stop:** `docker compose down` · **wipe DB volume:** `docker compose down -v`
- **Logs:** `docker compose logs backend --tail 100`

## Configuration

Everything is env-driven; the full list is in [`.env.example`](.env.example). Commonly touched:

- `DB_*` — MySQL (DSN includes `parseTime=true&loc=UTC` so session `TIMESTAMP` columns scan correctly; without this, login can succeed then **`GET /users` → 401**)
- `JWT_SECRET` — signing key (use **≥32** chars in production); optional `JWT_SECRET_PREVIOUS` during rotation ([docs/jwt-rotation.md](docs/jwt-rotation.md))
- `CSRF_*` — double-submit token names ([docs/auth-session.md](docs/auth-session.md))
- `PLATFORM_CLIENT_ID` / `PLATFORM_CLIENT_SECRET` — `POST /api/v1/auth/token` and bootstrap issuance
- `SESSION_*`, `AUTH_BOOTSTRAP_ENABLED` — cookie sessions and bootstrap bridge (see [docs/auth-session.md](docs/auth-session.md)). **`SESSION_COOKIE_SECURE=false`** on plain HTTP (default in [`.env.example`](.env.example)); if unset, the API defaults **Secure=true** and browsers **drop** session cookies—correct password login then **401** on `/users`. Docker Compose also sets `SESSION_COOKIE_SECURE=false` on the `backend` service.
- `MIGRATION_EXPECTED_VERSION` — if set, `/readyz` enforces migration version
- `CORS_ALLOWED_ORIGINS` — comma-separated origins (no `*`)
- `GIN_MODE` — `release` (default) or `debug`

## API (`/api/v1`)

| Area | Endpoints |
|------|-----------|
| Health | `GET /healthz`, `GET /readyz` |
| Auth | `POST /api/v1/auth/register`, `POST /api/v1/auth/login` (sets HttpOnly session cookie), `POST /api/v1/auth/logout`, `POST /api/v1/auth/refresh` |
| Auth (service / bridge) | `POST /api/v1/auth/token` (client credentials), `POST /api/v1/auth/bootstrap` (temporary; Origin allowlist + deprecation metadata in `data.bootstrap`) |
| Users | `GET`/`POST /api/v1/users`, `GET /api/v1/users/search?name=`, `GET`/`PUT`/`DELETE /api/v1/users/:id` — **Bearer JWT or session cookie** (`DELETE` → 204) |

JSON envelope: success `{ data, meta.request_id }`; errors `{ error: { code, message, details }, meta }`. Examples: `VALIDATION_ERROR`, `NOT_FOUND`, `UNAUTHORIZED`, `INTERNAL_ERROR`, `RATE_LIMITED`, `CONFLICT`. Unversioned `/api/...` (outside `/api/v1`) → **410 Gone**. SPA base URL/config: [`client/src/environments/environment.ts`](client/src/environments/environment.ts).

Auth: [docs/auth-session.md](docs/auth-session.md) · Admin UI scope: [docs/platform-admin-ui.md](docs/platform-admin-ui.md) · Bootstrap sunset: [docs/bootstrap-sunset.md](docs/bootstrap-sunset.md) · Desktop: [docs/desktop-auth-bridge.md](docs/desktop-auth-bridge.md) · JWT rotation: [docs/jwt-rotation.md](docs/jwt-rotation.md).

## Tests and CI

- **Unit (needs Go):** `cd go_CRUD_api && go test ./...`
- **Smoke (stack + migrate running):** `./scripts/test.ps1` · helper: `./scripts/migrate.ps1`

Workflow: [`.github/workflows/ci.yml`](.github/workflows/ci.yml) — unit tests and Compose smoke in parallel; BuildKit + [GHA cache](https://docs.docker.com/build/cache/backends/gha/) via [`docker-compose.ci.yml`](docker-compose.ci.yml) on push / same-repo PRs.

## Production

Strong `JWT_SECRET`; rotate `PLATFORM_CLIENT_*` if leaked; TLS in front; **`SESSION_COOKIE_SECURE=true`** behind HTTPS; keep MySQL private; bump `MIGRATION_EXPECTED_VERSION` after deploys; run migrations before new app replicas; tight `CORS_ALLOWED_ORIGINS`; aggregate logs; disable bootstrap (`AUTH_BOOTSTRAP_ENABLED=false`) when all clients use cookie login.

**Multiple API replicas:** backend rate limits and login lockout are **per-process memory** only—use an edge proxy or Redis-backed limits if you scale the API horizontally ([`docs/auth-session.md`](docs/auth-session.md)).

## Migrations and troubleshooting

- Guide: [`docs/migrations.md`](docs/migrations.md)
- Windows: if Docker errors on the engine socket, start Docker Desktop.
- `/readyz` stuck: run migrations; match `MIGRATION_EXPECTED_VERSION` to the DB; check `schema_migrations` for `dirty = 1`.
