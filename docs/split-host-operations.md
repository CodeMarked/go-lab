# Split-host operations, observability, suite checklist

**Audience:** operators and integrators (TaskStack, Marble). **Companion:** [platform-control-plane.md](platform-control-plane.md) (RBAC + routes), [tls-reverse-proxy.md](tls-reverse-proxy.md), [platform-api-consumer-brief.md](platform-api-consumer-brief.md).

---

## 1. What this repo covers here

- **Restore governance:** `backup_restore_requests` table (`000007_*`), API under `/api/v1/backups/*`, Angular **DataOps** page. Two **distinct** human approvers (neither the requester); physical restore execution stays **out of band** (mysqldump, vendor backup, runbook).
- **Readiness signal:** when `MIGRATION_EXPECTED_VERSION` is set, successful **`GET /readyz`** includes `migration_version` and `migration_expected_min` for orchestration and dashboards.
- **Structured logs (API):** restore workflow mutations emit `slog` lines with keys such as `backup_restore_request_created`, `backup_restore_request_approved`, `backup_restore_request_fulfilled`, plus `request_db_id`, `request_id` (HTTP request correlation), and actor user id where applicable.

---

## 2. Split-host / production hardening (checklist)

| Area | Guidance |
|------|-----------|
| **TLS** | Terminate TLS at the edge; set `SESSION_COOKIE_SECURE=true` and HSTS per [tls-reverse-proxy.md](tls-reverse-proxy.md). |
| **CORS** | `CORS_ALLOWED_ORIGINS` must list every browser origin that calls the API (TaskStack web, admin SPA). No wildcards in production. |
| **Cookies** | Cross-site cookie setups need explicit SameSite policy and proxy path alignment ([platform-admin-ui.md](platform-admin-ui.md)). |
| **Readiness gates** | Use `/readyz` (not only `/healthz`) before traffic shift; align `MIGRATION_EXPECTED_VERSION` with migrated schema ([migrations.md](migrations.md)). |
| **Secrets** | Platform client secret, JWT secret, DB credentials — rotation playbook [ops-secret-rotation.md](ops-secret-rotation.md). |

---

## 3. Suite integration checklist (TaskStack / Marble)

Cross-repo work remains **suite-owned**; use this as a punch list:

| Item | Owner | Notes |
|------|--------|--------|
| Desktop exchange → Bearer → **join-token** | Suite clients | [desktop-auth-bridge.md](desktop-auth-bridge.md), [openapi.yaml](openapi.yaml). |
| Verify **`token_use=join`** JWTs in game | Marble | Keys/TTL from platform contract; do not trust client gameplay claims. |
| **Heartbeat** / session semantics | Marble / infra | Not defined in go-lab; document in suite runbooks. |
| TaskStack **server-side** platform calls | TaskStack | Never expose `PLATFORM_CLIENT_SECRET` to browsers ([data-ownership.md](data-ownership.md)). |
| Contract tests / generated clients | Suite | Optional; OpenAPI in this repo is the contract ([platform-api-consumer-brief.md](platform-api-consumer-brief.md)). |

---

## 4. Observability suggestions

- **Probe JSON:** parse `/readyz` for `migration_version` vs deploy expectation; alert on `not_ready`.
- **Logs:** ship stdout JSON logs; filter on `backup_restore_*` for DataOps audits.
- **Traces:** if you add tracing later, propagate `X-Request-Id` (or platform `meta.request_id`) across TaskStack → platform → Marble.

---

*Last updated: 2026-03-24 — restore workflow + readiness fields + this runbook.*
