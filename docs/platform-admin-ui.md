# Platform admin UI (Angular in go-lab)

## Role

The Angular app under [`client/`](../client/) is an optional **operator console**: login, dashboard, user directory CRUD, and **permission-gated** nav (Cases, Players, Characters, Economy, DataOps, Security, Audit). **Cases**, Players, Characters, **Economy**, DataOps, Security, and Audit call gated `/api/v1/*` routes; nav items are **hidden unless** `GET /api/v1/security/me` reports the matching permission (e.g. `cases.read`). It is **not** the full suite control-plane UX — positioning vs TaskStack/Marble: [MASTER_PLAN.md](MASTER_PLAN.md), [data-ownership.md](data-ownership.md), [platform-api-consumer-brief.md](platform-api-consumer-brief.md).

## Configuration

See repo [README](../README.md) § Platform admin: `useBootstrapAuth`, **`apiBaseUrl: ''`** + same-origin proxy to `/api`, CORS vs `CORS_ALLOWED_ORIGINS`.

## Navigation (this SPA)

- **Cases:** list, create, and detail (notes, sanctions, recovery, appeals) via `/api/v1/cases/*` (`cases.read`, `cases.write`, `sanctions.write`, `recovery.write`, `appeals.resolve`); see [`cases/`](../client/src/app/cases/).
- **Players / Characters:** read-only JSON stubs; `GET` via [`platform.service.ts`](../client/src/app/platform.service.ts). **Economy** uses `GET /api/v1/economy/ledger` (`economy.read`).
- **DataOps:** restore workflow UI — status + list + create/approve/reject/fulfill/cancel via `/api/v1/backups/*` (`backups.read`, `backups.restore.*`); **`X-Platform-Action-Reason`** (≥ 10 chars) on mutations. Physical restore execution is out of band ([split-host-operations.md](split-host-operations.md)).
- **Security:** `GET /api/v1/security/me`; support ack **`POST /api/v1/support/ack`** with header **`X-Platform-Action-Reason`** (min length enforced server-side; UI requires ≥ 10 chars before submit).
- **Audit:** `GET /api/v1/audit/admin-events`.

**Grant platform roles in SQL** — [platform-operator-roles.md](platform-operator-roles.md).

## Session behavior

- After login, or when **`GET /api/v1/auth/csrf`** succeeds on reload, the app timers **`POST /api/v1/auth/refresh`** on `sessionRefreshIntervalMs` (cookie mode; skipped when `useBootstrapAuth` is true). Keep the interval safely under **`SESSION_IDLE_TTL_SECONDS`**.
- **`UnauthorizedInterceptor`:** **401** on protected API calls clears auth and navigates to **`/login`** (with `session=expired` query param). **401** on login/register/bootstrap/token and on the initial **`GET /api/v1/auth/csrf`** probe is ignored.

## Related

- [platform-control-plane.md](platform-control-plane.md) — RBAC matrix, route ↔ permission.
- [openapi.yaml](openapi.yaml) — contract.
