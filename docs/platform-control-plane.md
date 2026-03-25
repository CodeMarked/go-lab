# Platform control plane — Phase A/B/C reference

**Audience:** operators and implementers. **Companion:** [data-ownership.md](data-ownership.md) (suite-wide ownership), [platform-operator-roles.md](platform-operator-roles.md) (SQL to grant roles), [phase-c-split-host-operations.md](phase-c-split-host-operations.md) (split-host + restore runbook).

This doc **closes Phase A planning gaps** and tracks **Phase B/C** slices: domain boundaries, **RBAC matrix** (must match [`api/platformrbac/permissions.go`](../api/platformrbac/permissions.go) — update both when adding roles or permissions), and **route ↔ permission** map.

---

## 1. Domain boundaries

| Domain | In go-lab DB / API today | Authoritative / next step |
|--------|---------------------------|---------------------------|
| **identity_*** | `users`, `user_identities`, auth session + desktop exchange tables | Platform; see migrations `000002`–`000004` |
| **player_*** | No player tables; `GET /api/v1/players` stub | Gameplay-linked profiles live in **Marble** / suite DBs keyed by `platform_user_id` ([data-ownership.md](data-ownership.md)) |
| **character_*** | No character tables; `GET /api/v1/characters` stub | **Marble**-authoritative |
| **session_*** (operator sense) | Login = `auth_sessions`; join/desktop = `000004_*` | Extended “session registry” UI/API = Phase B if needed |
| **backup_*** | `backup_restore_requests` + `GET/POST /api/v1/backups/*` restore workflow | **Governance in DB/API (Phase C v0)**; physical backup/restore execution remains **operator-owned** out of band ([phase-c-split-host-operations.md](phase-c-split-host-operations.md)) |
| **economy_*** (operator ledger) | `economy_ledger_events` + `GET /api/v1/economy/ledger` (read-only) | Append-only read model in platform DB; authoritative gameplay economy remains **Marble** ([data-ownership.md](data-ownership.md)); suite ingests rows out of band |
| **audit_*** (control plane) | `admin_audit_events` (immutable rows for privileged actions) | Includes support ack + backup-restore workflow; extend as Phase B routes ship |
| **audit_*** (auth) | `auth_audit_events` | Existing auth security trail |

---

## 2. RBAC matrix (roles × permissions)

Permissions are string constants in [`api/platformrbac/permissions.go`](../api/platformrbac/permissions.go). **`operator`** grants **`*`** (wildcard). **§3** lists routes that call `RequirePlatformPermission` today; `security.write` and `audit.write` are granted to `security_admin` but **no HTTP route checks them yet** (they may still appear in `GET /api/v1/security/me`).

| Permission | `operator` | `support` | `security_admin` |
|------------|:----------:|:---------:|:----------------:|
| `players.read` | yes | yes | yes |
| `characters.read` | yes | yes | no |
| `backups.read` | yes | yes | yes |
| `backups.restore.request` | yes | yes | no |
| `backups.restore.approve` | yes | no | yes |
| `backups.restore.fulfill` | yes | no | yes |
| `security.read` | yes | yes | yes |
| `security.write` | yes | no | yes |
| `audit.read` | yes | yes | yes |
| `audit.write` | yes | no | yes |
| `platform.support.ack` | yes | yes | no |
| `economy.read` | yes | yes | yes |

Unknown role names grant **no** permissions.

**Phase C split:** `support` may **request** restores; `security_admin` **approves** (two distinct humans) and **fulfills** after out-of-band restore. **`operator`** retains full access via `*`.

---

## 3. Routes ↔ required permission

All routes below use **`BearerOrSession` + `RequireHumanUser`** (subjects must be `user:<id>`; `client:*` cannot pass). Mutating backup-restore routes also use **`RequirePlatformActionReason`** (header **`X-Platform-Action-Reason`**, min length enforced server-side) and a dedicated rate limiter.

| Method | Path | Permission |
|--------|------|------------|
| GET | `/api/v1/players` | `players.read` |
| GET | `/api/v1/characters` | `characters.read` |
| GET | `/api/v1/backups/status` | `backups.read` |
| GET | `/api/v1/backups/restore-requests` | `backups.read` |
| POST | `/api/v1/backups/restore-requests` | `backups.restore.request` + reason header |
| POST | `/api/v1/backups/restore-requests/:id/approve` | `backups.restore.approve` + reason header |
| POST | `/api/v1/backups/restore-requests/:id/reject` | `backups.restore.approve` + reason header |
| POST | `/api/v1/backups/restore-requests/:id/fulfill` | `backups.restore.fulfill` + reason header |
| POST | `/api/v1/backups/restore-requests/:id/cancel` | `backups.restore.request` + reason header (requester only) |
| GET | `/api/v1/security/me` | `security.read` |
| GET | `/api/v1/audit/admin-events` | `audit.read` |
| GET | `/api/v1/economy/ledger` | `economy.read` |
| POST | `/api/v1/support/ack` | `platform.support.ack` + header `X-Platform-Action-Reason` (min length) |

Contract: [openapi.yaml](openapi.yaml) (`platform` tag).

---

## 4. Granting roles

See [platform-operator-roles.md](platform-operator-roles.md) for `INSERT` examples into `user_platform_roles`.

---

## Readiness (`/readyz`)

When **`MIGRATION_EXPECTED_VERSION`** is set (> 0), a successful ready response includes **`migration_version`** and **`migration_expected_min`**. See [migrations.md](migrations.md).

---

## Future (not yet shipped)

- Rich player/character **data** and **mutations** (sanctions, recovery, etc.).
- **Role assignment API/UI** (today: SQL only).
- Unified **audit taxonomy** across `auth_audit_events` and `admin_audit_events` ([MASTER_PLAN.md](MASTER_PLAN.md) §9 P2).
