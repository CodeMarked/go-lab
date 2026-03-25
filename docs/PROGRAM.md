# Rolling program (go-lab)

## Why this file exists

**Problem it solves:** Someone asks “what did we commit to for *this* month/quarter?” — without that answer living only in Slack or an external suite deck.

**What it is:** A **short, time-boxed outcomes list** (dates + owner + a few bullets) plus **assumptions** and a **history table** of past windows. It is planning *narrative* for a window, not a second architecture spec.

**What it is not (put those elsewhere):**

| Not here | Put it in |
|----------|-----------|
| Long architecture rationale, principles, phased roadmap | [MASTER_PLAN.md](MASTER_PLAN.md) §1–§8 |
| Prioritized execution backlog (P0/P1…) | [MASTER_PLAN.md](MASTER_PLAN.md) §9 |
| Deep behavior of auth, OIDC, sessions | [auth-session.md](auth-session.md), [oidc-auth0.md](oidc-auth0.md), etc. |
| Formal one-topic decisions | [adr-account-linking.md](adr-account-linking.md) pattern or future `docs/adr/*` |
| “Design spiel” write-ups | Topic doc or MASTER_PLAN §6 row — not PROGRAM |

**Do you need it?** **Optional but useful** if you run explicit planning cycles. If you only ever track work in MASTER_PLAN §9 + issues/PRs, you can leave the template placeholders below until you want a dated narrative; nothing breaks.

**Canonical priorities** remain [MASTER_PLAN.md](MASTER_PLAN.md) §9. PROGRAM does not replace the backlog — it **summarizes intent** for humans and auditors for a calendar window.

---

## Current window

| Field | Value |
|-------|--------|
| **Window start** | *YYYY-MM-DD* |
| **Window end** | *YYYY-MM-DD* |
| **Owner** | *name or role* |

### Outcomes (this window)

1. *e.g. OpenAPI covers `/api/v1` auth + users with cookie + Bearer schemes.*
2. *e.g. CI includes or documents migration snapshot / negative auth tests.*
3. *…*

### Dependencies / assumptions

- *e.g. Admin SPA remains cookie-first; no Auth0-in-SPA without explicit decision (MASTER_PLAN non-goals).*

### In flight (optional — prune when merged)

- *e.g. OpenAPI starter — PR #123*

### Carry-over to next window

- *Items not finished; copy forward when you roll the dates.*

---

## History

**Closed windows:** append a row when you **roll** the dates (snapshot what shipped; link PRs or tags). That section is **append-only** so past windows stay readable. The **current window** section above is edited freely each cycle.

| Closed window | Summary |
|----------------|---------|
| *YYYY-MM — YYYY-MM* | *What shipped; link PRs or tags if useful.* |

---

## Related

[MASTER_PLAN.md](MASTER_PLAN.md) · [README.md](README.md) · [ci.md](ci.md)
