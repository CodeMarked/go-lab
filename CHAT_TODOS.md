# Go-Lab Chat Todos (operator scratch)

> **Canonical plan:** [`docs/MASTER_PLAN.md`](docs/MASTER_PLAN.md) — decisions, shipped matrix, phases, **P0/P1 backlog**. Update that file when priorities or shipped state change. **This file** is for scratch notes and in-flight tracking, not the authoritative backlog.

## Next focus (one line)

Phase 4 (OpenAPI, tests, TLS/runbook) → Phase 5 (Marble/TaskStack). **Human-only** routes when M2M matters. Don’t rebuild cookie admin / OIDC / Redis paths without an explicit ask.

## Ground rules

- **Prune completed scratch:** Remove or collapse items once the work is **merged** (PR in `main` or your release branch). **Git history** is the audit trail for what used to be here — you do not need to keep every line forever.
- **Until merged:** It is fine to list an item with a **PR link** or branch name; delete that line when the PR lands and [`MASTER_PLAN.md`](docs/MASTER_PLAN.md) §7/§9 already reflect the outcome.
- **Do not** mirror the full prioritized backlog from MASTER_PLAN §9 here long-term — that duplicates the source of truth.
- Security-first; minimal deps; no destructive workflows; Compose baseline; migration-only schema.

## Where everything else lives

| What | Where |
|------|--------|
| Suite + go-lab summary, version stance | MASTER_PLAN §1–§5 |
| Architecture decisions | MASTER_PLAN §6 |
| Shipped vs open | MASTER_PLAN §7 |
| Phases 1–5 | MASTER_PLAN §8 |
| Prioritized backlog | MASTER_PLAN §9 |
| Open questions | MASTER_PLAN §10 |
| Rolling program narrative | [docs/PROGRAM.md](docs/PROGRAM.md) |
| CI contract | [docs/ci.md](docs/ci.md) |
| Agent priming (short) | MASTER_PLAN §12 |

## Original P0 core-auth — satisfied

Register/login/logout/refresh, sessions, migrations, Argon2, CSRF, limits, bootstrap scaffolding, tests, docs — all landed. Detail: MASTER_PLAN §7–§8 Phase 1.

## Completed (historical — this effort)

- Envelope API + auth middleware; no platform creds in SPA; bootstrap endpoint; migration hygiene; auth error visibility.
- **P0 auth:** `000002_*`, sessions, Argon2id, bootstrap hardening, TTL, audit, [`docs/auth-session.md`](docs/auth-session.md), tests.
- **Hardening:** CSRF, rate limits, lockout, change-password + revoke all, `JWT_SECRET_PREVIOUS`, [`docs/bootstrap-sunset.md`](docs/bootstrap-sunset.md) et al., tests.
- **SPA:** refresh timer, 401 → login, same-origin proxy, ops fixes.
- **Later:** OIDC + `user_identities`, optional Redis — [`docs/oidc-auth0.md`](docs/oidc-auth0.md), [`docs/auth-session.md`](docs/auth-session.md).

## Quick handoff

[`docs/README.md`](docs/README.md) · `/healthz` + `/readyz` + `./scripts/test.ps1` · new deps need one-line rationale · `MIGRATION_EXPECTED_VERSION` after migrations.

## Architecture agent (minimal priming)

Read **`docs/MASTER_PLAN.md`**, then `oidc-auth0.md` + `auth-session.md`. Task: shipped vs planned, contradictions, next 3 milestones — no code unless asked. Full paste block: MASTER_PLAN §12.

## Commit note

Commit **`docs/MASTER_PLAN.md`**, **`docs/PROGRAM.md`** (when you roll a planning window), and code. Commit this file when scratch/PR tracking here is useful to the team; **remove** stale lines after merge so the file stays small.
