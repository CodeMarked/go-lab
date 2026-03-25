# Documentation

## What each kind of doc is for

| Kind | Role | Required? |
|------|------|-----------|
| **[MASTER_PLAN.md](MASTER_PLAN.md)** | Single place for roadmap, **§6 decisions**, shipped matrix, **§9 backlog**, open questions. | **Yes** — canonical plan. |
| **Topic guides** (auth-session, oidc-auth0, migrations, …) | Accurate behavior and ops detail for **one area**; keep them tight and link from MASTER_PLAN. | **Yes** for anything you ship. |
| **[PROGRAM.md](PROGRAM.md)** | Short **dated window** (outcomes + assumptions + optional PR lines); **not** architecture essays. Optional narrative layer. | **No** — use when you want a quarter/month story in-repo. |
| **[ci.md](ci.md)** | Human-readable summary of what CI enforces; keep in sync with `.github/workflows`. | **Recommended** if CI is non-obvious. |
| **[../CHAT_TODOS.md](../CHAT_TODOS.md)** | Scratch + in-flight notes; **prune** when merged; do not duplicate §9 long-term. | **No** — convenience only. |

**Architect “design spiel”** belongs in **MASTER_PLAN** (summary + §6) and **topic docs** (depth). **PROGRAM** is only “what we said we’d achieve by date X.”

---

| Doc | Topics |
|-----|--------|
| [MASTER_PLAN.md](MASTER_PLAN.md) | Suite + go-lab roadmap, decisions, backlog |
| [PROGRAM.md](PROGRAM.md) | Optional rolling near-term outcomes window |
| [ci.md](ci.md) | GitHub Actions CI jobs and local parity |
| [auth-session.md](auth-session.md) | Sessions, CSRF, limits, Redis |
| [oidc-auth0.md](oidc-auth0.md) | OIDC Bearer, `aud`, identities, M2M |
| [adr-account-linking.md](adr-account-linking.md) | Password ↔ IdP linking policy |
| [jwt-rotation.md](jwt-rotation.md) | HS256 rotation |
| [bootstrap-sunset.md](bootstrap-sunset.md) | Disabling bootstrap |
| [desktop-auth-bridge.md](desktop-auth-bridge.md) | Desktop / automation |
| [platform-admin-ui.md](platform-admin-ui.md) | Admin SPA |
| [migrations.md](migrations.md) | SQL migrations, `/readyz` |

[Repo README](../README.md) · [go_CRUD_api README](../go_CRUD_api/README.md)
