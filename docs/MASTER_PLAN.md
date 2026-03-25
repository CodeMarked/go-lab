# Master plan — Product suite + go-lab (platform API)

**Single planning doc:** specs (summary), decisions, shipped matrix, phases, prioritized backlog. Edit when priorities change; deep detail stays in `docs/*.md`.

**Related:** [Documentation index](README.md) · [Repo README](../README.md)

---

## Table of contents

1. [Executive snapshot](#1-executive-snapshot)
2. [Product Suite Architecture Spec v0.1 (condensed)](#2-product-suite-architecture-spec-v01-condensed)
3. [Go-lab platform API — architecture spec v0.1](#3-go-lab-platform-api--architecture-spec-v01)
4. [Go-lab repo — implementation reality](#4-go-lab-repo--implementation-reality)
5. [Version stance (suite vs repo)](#5-version-stance-suite-vs-repo)
6. [Architecture decisions (consensus)](#6-architecture-decisions-consensus)
7. [Shipped vs still open](#7-shipped-vs-still-open)
8. [Roadmap phases (operator)](#8-roadmap-phases-operator)
9. [Prioritized backlog](#9-prioritized-backlog)
10. [Open questions](#10-open-questions)
11. [Constraints & ground rules](#11-constraints--ground-rules)
12. [Agent handoff & priming](#12-agent-handoff--priming)
13. [Maintenance notes](#13-maintenance-notes)

---

## 1. Executive snapshot

| Item | State |
|------|--------|
| **Products** | **Marble** (game sim), **TaskStack** (control-plane UX), **Go-lab** (this repo — platform API + admin SPA + migrations) |
| **Suite spec maturity** | **v0.1** (foundation; not full prod hardening of every subsystem) |
| **Go-lab vs suite Phase 1** | Auth/session **ahead** of thin “skeleton” wording; **OpenAPI**, **migration snapshot tests**, **TLS/runbook** still gap vs “quality gates” |
| **Auth model** | Hybrid: **cookie + CSRF** admin SPA; **HS256** Bearer (`/auth/token`); optional **OIDC RS256** Bearer (`OIDC_ISSUER_URL` + `OIDC_AUDIENCE`); **`user_identities`** for `(issuer, sub)` |
| **Scale-out** | Optional **Redis** (`REDIS_URL`) for shared rate limits + email lockout; else in-memory |
| **Next focus (priority)** | **Phase 4:** OpenAPI, negative tests, migration snapshots, TLS/runbook → then **Phase 5** Marble/TaskStack integration |

---

## 2. Product Suite Architecture Spec v0.1 (condensed)

**Products:** **Marble** — simulation + netcode authority. **TaskStack** — accounts / orchestration UX (consumes platform API). **Go-lab** — this repo: platform API, admin SPA, migrations.

**Principles:** Gameplay can stay offline; platform is additive. **Multi-repo** with explicit contracts. **Self-host Docker Compose** first; split-host / cloud later without boundary rewrites.

**MySQL:** schema = **migrations only**. **API:** JSON `/api/v1`, additive changes preferred; **OpenAPI** is the formal contract gap (go-lab Phase 4). **Trust:** clients trust API; game server trusts platform tokens later; platform never trusts client gameplay state.

**Deploy:** (A) single-host Compose — primary. (B) split host. (C) managed cloud.

**Security / observability / quality:** env config, validation, CORS allowlist, envelopes, no secrets in repo; structured logs; health + smoke; CI green; release = migrations + upgrade path.

**Non-goals (v0.1):** microservices split, enterprise SSO, global matchmaking, full streaming stack.

**Suite roadmap (spec):** (1) **Stabilize core** — API conventions, auth, smoke, docs/runbook. (2) **Control plane** — TaskStack auth flows, entitlements sketch, OpenAPI, contract tests. (3) **Marble bridge** — handshake, token validation, heartbeat, split-host guidance.

**Versioning:** spec `vMAJOR.MINOR`; breaking boundaries → changelog + migration note + compatibility statement.

*Full prose lives in your canonical Product Suite doc if you maintain one separately; this section is the working summary.*

---

## 3. Go-lab platform API (this repo)

Self-hostable **Go API**: users, auth/session, tokens, health/readiness. **Owns:** user data, sessions (evolving), future entitlements. **Does not own:** game sim, rendering, physics.

**Ops:** Compose; env-only config; **no runtime DDL**; `go test ./...` + smoke + CI.

**Exit (original v0.1 module):** clean migrate + tests + runbook — **extended** in practice by full cookie auth, OIDC, Redis (see §7).

---

## 4. Go-lab repo — implementation reality

| Area | Location / notes |
|------|------------------|
| **Backend** | `go_CRUD_api/` — Gin, `middleware/`, `auth/`, `authstore/`, `myhandlers/` |
| **Frontend** | `client/` — Angular admin SPA |
| **Migrations** | `migrations/` — `000002_*` auth/session/users; `000003_*` `user_identities` |
| **Compose** | `docker-compose.yml` — mysql, migrate, backend, frontend; optional **`redis`** profile |
| **Docs** | `docs/*.md` — auth-session, oidc-auth0, jwt-rotation, bootstrap-sunset, desktop bridge, platform-admin-ui, migrations; [PROGRAM.md](PROGRAM.md) (optional dated outcomes); [ci.md](ci.md) (CI contract); roles: [README.md](README.md) |
| **Key env** | See [`.env.example`](../.env.example): `JWT_*`, `SESSION_*`, `OIDC_*`, `REDIS_URL`, `MIGRATION_EXPECTED_VERSION`, CSRF, platform client creds |

---

## 5. Version stance (suite vs repo)

- **Product Suite Architecture Spec** remains **v0.1** until you bump it deliberately.
- **Go-lab** can be described as **“v0.1 + auth/OIDC foundation”**: real sessions, CSRF, abuse limits, OIDC API path, optional Redis — while **OpenAPI**, **snapshot migration tests**, and **full TLS/runbook** remain the main **spec-shaped** gaps (aligned with Phase 4 below).

---

## 6. Architecture decisions (consensus)

| Topic | Decision |
|-------|----------|
| **Identity source of truth** | **Hybrid:** first-party email/password + opaque session cookie for admin SPA; local `users` + `user_identities` for `(issuer, sub)`. Optional OIDC access JWT as Bearer when `OIDC_*` set. Platform owns authz/tenancy/audit in DB. |
| **Auth0-only vs API-only** | Prefer hybrid. Pure Auth0-only only if API is a thin BFF with almost no local user model. |
| **TaskStack / Marble** | TaskStack: **consumer** of platform APIs. Marble: gameplay authority; join/session trust **later** (Phase 5). |
| **Canonical `aud` (OIDC)** | One API identifier: **`OIDC_AUDIENCE`**. Do not conflate with **`JWT_AUDIENCE`** unless deliberately standardized and documented. See [oidc-auth0.md](oidc-auth0.md). |
| **Multiple `aud` / multi-API** | Prefer gateway as single `aud`, or separate Auth0 APIs/tokens, or multi-aud with explicit validation only. |
| **`sub`** | Identity = **`(issuer, sub)`** only; `user_identities` enforces. |
| **Account linking** | No naive email auto-link. Safe: verified email + explicit action, admin linking, or migration “password once to attach.” JIT new user on first OIDC `(iss, sub)` today. |
| **Refresh tokens** | Auth0 refresh = client/Auth0/BFF; do not stuff into `auth_refresh_tokens` without naming flow ownership. |
| **Cookie vs Bearer** | Shipped default: HttpOnly cookie + CSRF for admin. Bearer for HS256 and OIDC. Bearer-in-SPA = XSS tradeoff; BFF/cookie preferred for IdP-in-browser if avoiding tokens in JS. |
| **M2M** | `…@clients` → `client:<id>`; not in `users`. Human-only route guards = follow-up. |
| **Redis** | Optional; fail **open** on errors for limits/lockout. Fail closed only for controls that need strong consistency. |
| **Gateway vs app limits** | Edge: coarse IP/TLS/WAF. App: email lockout, route semantics. Avoid duplicate same numeric cap unless intentional. |
| **Clock / gameplay** | JWT strictness affects **API** calls, not game loop. NTP on API hosts; optional leeway = polish. |
| **Cross-platform play** | Login UX can vary; matchmaking/connectivity = game + netcode + join tokens, orthogonal to OIDC skew. |
| **IdP portability** | OIDC primitives in config/code; avoid vendor SDKs scattered through handlers. |
| **Env / config catalog** | **Authoritative variable list:** [`.env.example`](../.env.example). New or changed settings belong there and are cross-referenced from topic docs (`auth-session`, `oidc-auth0`, etc.). |

Deep dives: [oidc-auth0.md](oidc-auth0.md), [auth-session.md](auth-session.md), [adr-account-linking.md](adr-account-linking.md), [jwt-rotation.md](jwt-rotation.md), [migrations.md](migrations.md), [platform-admin-ui.md](platform-admin-ui.md). Full index: [README.md](README.md).

---

## 7. Shipped vs still open

| Area | Status |
|------|--------|
| Browser register/login/logout/refresh, session cookie, `/users` Bearer or cookie | **Shipped** |
| `000002_*` users auth cols, sessions, refresh shell, audit | **Shipped** |
| Argon2id | **Shipped** |
| Idle + absolute session, logout | **Shipped** |
| Bootstrap bridge + sunset docs | **Shipped** ([bootstrap-sunset.md](bootstrap-sunset.md)) |
| JWT rotation | **Incremental** — `JWT_SECRET_PREVIOUS` ([jwt-rotation.md](jwt-rotation.md)) |
| Change-password + revoke all sessions | **Shipped** |
| Desktop bridge | **Doc** ([desktop-auth-bridge.md](desktop-auth-bridge.md)) |
| CSRF | **Shipped** |
| Rate limits + email lockout | **Shipped** (memory default; [Redis optional](auth-session.md)) |
| Admin Angular UX | **Shipped** ([platform-admin-ui.md](platform-admin-ui.md)) |
| OIDC + `user_identities` | **Shipped** when `OIDC_*` set ([oidc-auth0.md](oidc-auth0.md)) |

---

## 8. Roadmap phases (operator)

Phases are **go-lab–centric** and align with suite direction; numbering is not identical to §2.17 naming.

### Phase 1 — Platform core auth — **shipped**

Cookie + CSRF admin SPA, same-origin API proxy, session TTL, in-memory limits (default), `000002_*`, HS256 + `JWT_SECRET_PREVIOUS`, bootstrap sunset docs.

### Phase 2 — External IdP (OIDC) — **shipped (env-gated)** + follow-ups

- **Shipped:** `OIDC_ISSUER_URL` + `OIDC_AUDIENCE`; RS256 via discovery/JWKS; `000003_*`; `BearerOrSession` HS256 then OIDC; M2M → `client:<id>`; safe reject logs.
- **Follow-ups:** optional `OIDC_JWKS_URL`; clock leeway; richer reject metadata; Connect Auth0 / linking; human-only middleware.

### Phase 3 — Scale-out & edge — **mostly shipped (optional path)**

Redis + gateway/app split documented; Compose `redis` profile. Triggers: multi-replica LB.

### Phase 4 — Contracts & hardening (P1 lift) — **next**

OpenAPI; negative auth tests; migration-from-snapshot; TLS/runbook; TaskStack vs go-lab admin positioning doc updates.

### Phase 5 — Marble / TaskStack integration — **later**

**In this repo:** join-token contract (API/docs), desktop bridge implementation, OpenAPI for consumer surfaces. **Outside this repo:** game handshake semantics, heartbeat, full split-host playbooks beyond platform config. TaskStack as consumer (M2M + user) spans both.

---

## 9. Prioritized backlog

Order is **default execution priority** for platform work unless you reprioritize.

### P0 — Now / next session

1. **Phase 4 kickoff:** OpenAPI starter (`/api/v1` auth + user surfaces + security schemes: cookie vs bearer).
2. **Tests:** auth negative paths; migration-from-snapshot (or documented manual gate until automated).
3. **Runbook / TLS:** reverse-proxy HTTPS baseline, `Secure` cookie + HSTS notes.

### P1 — Soon after

4. **Positioning:** expand [platform-admin-ui.md](platform-admin-ui.md) vs future TaskStack.
5. **Incident / ops:** secret rotation checklist (cross-ref [jwt-rotation.md](jwt-rotation.md) for HS256; IdP-side rotation for OIDC); auth audit event taxonomy.
6. **Phase 2 follow-ups** (pick as needed): human-only routes; OIDC leeway; linking endpoint spec + implementation.

### P2 — In-repo program narrative

7. When you run a **time-boxed plan** (month/quarter), update **[PROGRAM.md](PROGRAM.md)** — outcomes, assumptions, optional in-flight PR lines; **closed windows** append to its history table. Skip or leave placeholders if you only use §9 + issues. If an external suite doc has a parallel narrative, **summarize or link it in PROGRAM** so intent is not only outside the repo.

### P3 — Phase 5 (blocked on contract + trust model)

**Go-lab–owned when scoped:** join-token **API contract** documented in-repo; **desktop bridge implementation** per [desktop-auth-bridge.md](desktop-auth-bridge.md); consumer-facing surfaces captured in **OpenAPI** once Phase 4 starts.

**Suite / game–owned (not specified here):** Marble handshake semantics, in-game heartbeat behavior, split-host playbooks beyond API/env documentation.

### Explicit non-goals (unless you change §6)

- Rebuild shipped cookie admin without explicit ask.
- SPA Auth0 login UI before explicit ask (cookie path remains default for admin).

---

## 10. Open questions

- **Bootstrap disable milestone:** *TBD* — choose **release tag** or **calendar date** per [bootstrap-sunset.md](bootstrap-sunset.md); record the choice **here** when set (and in release notes / tags as appropriate).
- **Desktop user auth shape:** exchange code vs token-in-body / separate endpoint — [desktop-auth-bridge.md](desktop-auth-bridge.md); decide before implementation work.
- JWT clock **leeway**: implement vs strict + NTP only?
- Which routes **reject `client:*`** first?
- **`OIDC_JWKS_URL`** override for locked-down networks?
- **Argon2 `m,t,p` changes:** if parameters change, require a **re-hash-on-login** (or equivalent) strategy — document in [auth-session.md](auth-session.md) or a short ADR when triggered.
- PM tooling: optional until multiple assignees need dates/queues; use this file + `docs/adr/*.md`.

---

## 11. Constraints & ground rules

Migration-only schema; **`/api/v1`** additive where possible; no secrets in frontend; Compose baseline; security-first; **new deps** need a one-line justification. Bump **`MIGRATION_EXPECTED_VERSION`** with schema ([migrations.md](migrations.md)). **CI contract:** [.github/workflows/ci.yml](../.github/workflows/ci.yml); human summary in [ci.md](ci.md) — update both when workflow meaningfully changes.

---

## 12. Agent handoff & priming

**Smoke / health:** `/healthz`, `/readyz`; `scripts/test.ps1` when Docker is up. **CI:** [ci.md](ci.md).

**Architecture-only agent — paste as system or first message:**

```text
You are the architecture steward for the go-lab monorepo (platform API) within Marble + TaskStack (Product Suite Architecture Spec v0.1).

Sources of truth (read in order):
1) docs/MASTER_PLAN.md (this repo)
2) docs/README.md — full doc index; then topic files as needed (oidc-auth0, auth-session, desktop-auth-bridge, bootstrap-sunset, jwt-rotation, migrations, adr-account-linking, platform-admin-ui, PROGRAM, ci)
3) Any attached suite spec deltas

Your job: shipped vs planned; flag contradictions; propose next 3 milestones with risks; do not expand scope silently. Do not implement code unless asked.
```

---

## 13. Maintenance notes

Update **§1 + §9** each sprint; **§7** on ship; **§6** on principle changes; **[PROGRAM.md](PROGRAM.md)** when you **roll** a dated planning window (optional doc — see [README.md](README.md) “What each kind of doc is for”). **ADRs:** [adr-account-linking.md](adr-account-linking.md) pattern. Commit this file; secrets only in `.env`.

---

*Last consolidated: 2026-03-21 — adjust date when you make substantive edits.*
