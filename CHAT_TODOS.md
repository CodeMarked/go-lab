# Session notes (CHAT_TODOS)

**Canonical plan:** [docs/MASTER_PLAN.md](docs/MASTER_PLAN.md) §7–§9. Use this file for **short-lived session notes and in-flight PR reminders** only. Trim after merge; do not duplicate §9 here.

## Next focus

**Phase C (platform) shipped (v0):** `000007_*` `backup_restore_requests`; `GET/POST /api/v1/backups/*` restore workflow; Angular **DataOps**; permissions `backups.restore.*`; `/readyz` exposes `migration_version` when `MIGRATION_EXPECTED_VERSION` is set. Runbook: [docs/phase-c-split-host-operations.md](docs/phase-c-split-host-operations.md). Contract: [docs/openapi.yaml](docs/openapi.yaml).

**Next (suite / game–owned or cross-repo):** wire Marble/TaskStack clients to the exchange → desktop Bearer → join-token path; implement game-side validation of `token_use=join` JWTs; heartbeat / split-host playbooks. Backlog: [docs/MASTER_PLAN.md](docs/MASTER_PLAN.md) §9. Data model: [docs/data-ownership.md](docs/data-ownership.md).

**Fresh DB:** `docker compose down -v` → up → `migrate`; set **`MIGRATION_EXPECTED_VERSION=7`** in `.env`. Example defaults: [`.env.example`](.env.example).

## Rules

- Prune merged items. Git history is the audit trail.
- Security-first; minimal deps; migrations-only schema; Compose baseline.

**Handoff:** [docs/README.md](docs/README.md) · [docs/ci.md](docs/ci.md) · `./scripts/ci-local.ps1` (fast checks) · [docs/platform-api-consumer-brief.md](docs/platform-api-consumer-brief.md) (external integrators) · `/healthz`, `/readyz` · `./scripts/test.ps1` · bump `MIGRATION_EXPECTED_VERSION` when schema changes.
