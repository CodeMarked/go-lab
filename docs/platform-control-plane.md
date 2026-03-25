# Platform control plane — RBAC and routes

**Audience:** operators and implementers. **Companion:** [data-ownership.md](data-ownership.md) (suite-wide ownership), [platform-operator-roles.md](platform-operator-roles.md) (SQL to grant roles), [split-host-operations.md](split-host-operations.md) (split-host + restore runbook).

Domain boundaries, **RBAC matrix** (must match [`api/platformrbac/permissions.go`](../api/platformrbac/permissions.go) — update both when adding roles or permissions), and **route ↔ permission** map.

---

## 1. Domain boundaries

| Domain | In go-lab DB / API today | Authoritative / next step |
|--------|---------------------------|---------------------------|
| **identity_*** | `users`, `user_identities`, auth session + desktop exchange tables | Platform; see migrations `000002`–`000004` |
| **player_*** | `GET /api/v1/players` stub | Gameplay-linked profiles live in **Marble** / suite DBs keyed by `platform_user_id` ([data-ownership.md](data-ownership.md)) |
| **character_*** | `GET /api/v1/characters` stub | **Marble**-authoritative |
| **operator_case_*** | `operator_cases`, `operator_case_notes`, `operator_case_actions` + `/api/v1/cases/*` | Platform governance + audit; Marble applies gameplay sanctions/recovery **out of band** |
| **session_*** (operator sense) | Login = `auth_sessions`; join/desktop = `000004_*` | Extended “session registry” UI/API if needed later |
| **backup_*** | `backup_restore_requests` + `GET/POST /api/v1/backups/*` restore workflow | **Governance in DB/API**; physical backup/restore execution remains **operator-owned** out of band ([split-host-operations.md](split-host-operations.md)) |
| **economy_*** (operator ledger) | `economy_ledger_events` + `GET /api/v1/economy/ledger` (read-only) | Append-only read model in platform DB; authoritative gameplay economy remains **Marble** ([data-ownership.md](data-ownership.md)); suite ingests rows out of band |
| **audit_*** (control plane) | `admin_audit_events` (immutable rows for privileged actions) | Includes support ack + backup-restore workflow + case lifecycle actions |
| **audit_*** (auth) | `auth_audit_events` | Existing auth security trail |

---

## 2. RBAC matrix (roles × permissions)

Permissions are string constants in [`api/platformrbac/permissions.go`](../api/platformrbac/permissions.go). **`operator`** grants **`*`** (wildcard). **§3** lists routes that call `RequirePlatformPermission` today; `security.write` and `audit.write` are granted to `security_admin` but **no HTTP route checks them yet** (they may still appear in `GET /api/v1/security/me`).

| Permission | `operator` | `support` | `security_admin` | `gm_liveops` |
|------------|:----------:|:---------:|:----------------:|:------------:|
| `players.read` | yes | yes | yes | yes |
| `characters.read` | yes | yes | no | yes |
| `backups.read` | yes | yes | yes | no |
| `backups.restore.request` | yes | yes | no | no |
| `backups.restore.approve` | yes | no | yes | no |
| `backups.restore.fulfill` | yes | no | yes | no |
| `security.read` | yes | yes | yes | yes |
| `security.write` | yes | no | yes | no |
| `audit.read` | yes | yes | yes | yes |
| `audit.write` | yes | no | yes | no |
| `platform.support.ack` | yes | yes | no | yes |
| `economy.read` | yes | yes | yes | yes |
| `cases.read` | yes | yes | yes | yes |
| `cases.write` | yes | yes | no | yes |
| `sanctions.write` | yes | no | yes | yes |
| `recovery.write` | yes | yes | no | yes |
| `appeals.resolve` | yes | yes | yes | yes |

Unknown role names grant **no** permissions.

**Restore workflow roles:** `support` may **request** restores; `security_admin` **approves** (two distinct humans) and **fulfills** after out-of-band restore. **`operator`** retains full access via `*`.

---

## 3. Routes ↔ required permission

All routes below use **`BearerOrSession` + `RequireHumanUser`** (subjects must be `user:<id>`; `client:*` cannot pass). Mutating backup-restore routes and case **mutations** also use **`RequirePlatformActionReason`** (header **`X-Platform-Action-Reason`**, min length enforced server-side) and dedicated rate limiters where noted.

| Method | Path | Permission |
|--------|------|------------|
| GET | `/api/v1/players` | `players.read` |
| GET | `/api/v1/characters` | `characters.read` |
| GET | `/api/v1/cases` | `cases.read` |
| POST | `/api/v1/cases` | `cases.write` + reason header |
| GET | `/api/v1/cases/:id` | `cases.read` |
| PATCH | `/api/v1/cases/:id` | `cases.write` + reason header |
| GET | `/api/v1/cases/:id/notes` | `cases.read` |
| POST | `/api/v1/cases/:id/notes` | `cases.write` + reason header |
| GET | `/api/v1/cases/:id/actions` | `cases.read` |
| POST | `/api/v1/cases/:id/sanctions` | `sanctions.write` + reason header |
| POST | `/api/v1/cases/:id/recovery-requests` | `recovery.write` + reason header |
| POST | `/api/v1/cases/:id/appeals/resolve` | `appeals.resolve` + reason header |
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

## 4. Admin audit taxonomy (case workflow)

Immutable `admin_audit_events.action` values for operator cases (extend consistently):

| Action | When |
|--------|------|
| `case.created` | POST `/cases` |
| `case.updated` | PATCH `/cases/:id` |
| `case.note_added` | POST `/cases/:id/notes` |
| `case.sanction_recorded` | POST `/cases/:id/sanctions` |
| `case.recovery_requested` | POST `/cases/:id/recovery-requests` |
| `case.appeal_resolved` | POST `/cases/:id/appeals/resolve` |

`resource_type` is `operator_case`; `resource_id` is the case id string.

---

## 5. Granting roles

See [platform-operator-roles.md](platform-operator-roles.md) for `INSERT` examples into `user_platform_roles`.

---

## 6. Break-glass and elevated access

There is **no automated JIT elevation** in the API today. Use this operational pattern when someone needs temporary full control (incident response, production recovery):

1. **Prefer narrow roles:** grant `gm_liveops` or add `support` / `security_admin` via SQL rather than sharing the `operator` account.
2. **Time-bound:** after the incident, **revoke** extra rows from `user_platform_roles` (or rotate the affected user’s password / sessions if the account was exposed).
3. **Audit:** every privileged mutation should already carry `X-Platform-Action-Reason`; for break-glass work, require ticket or incident ID in that field and review `admin_audit_events` afterward.
4. **Never** put `PLATFORM_CLIENT_SECRET` or long-lived M2M tokens in the admin SPA; human panel stays on cookie or user Bearer paths.

For secret and key rotation, see [ops-secret-rotation.md](ops-secret-rotation.md).

---

## Readiness (`/readyz`)

When **`MIGRATION_EXPECTED_VERSION`** is set (> 0), a successful ready response includes **`migration_version`** and **`migration_expected_min`**. See [migrations.md](migrations.md).

---

## Future (not yet shipped)

- **Role assignment API/UI** (today: SQL only).
- Unified **audit taxonomy** across `auth_audit_events` and `admin_audit_events` for non-case actions ([MASTER_PLAN.md](MASTER_PLAN.md) §8 P2).
