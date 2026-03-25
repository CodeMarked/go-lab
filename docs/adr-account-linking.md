# ADR: Account linking (password ↔ external IdP)

**Status:** Accepted (policy). Implementation may lag; behavior must match this doc once linking ships.

## Context

- Local users exist with **email + password** (`users` + `auth_sessions`).
- External logins are keyed by **`(issuer, sub)`** — `sub` is not globally unique ([oidc-auth0.md](oidc-auth0.md)).
- Linking by **email alone** risks takeover if IdP email is wrong, stale, or attacker-controlled.

## Decision

1. **No automatic link** on first OIDC login when `(issuer, sub)` is unknown, even if `email` matches an existing row.
2. **Allowed linking paths** (pick per product; all require audit events):
   - **Explicit “Connect IdP”** after authenticated password session (or magic link) proving control of the existing account.
   - **Admin-only** link in operator tools.
   - **Migration window:** one-time password sign-in (or dedicated endpoint) that attaches `user_identities` to the current `users.id`.
3. **Data model:** one `users` row; one or more `user_identities` rows `(user_id, issuer, subject)`. No duplicate `users` for the same person.

## Consequences

- Some users may have **two accounts** until they complete a linking flow; communicate in UX/support.
- OIDC JIT provisioning continues to create **new** users when no identity row matches `(issuer, sub)` ([oidc-auth0.md](oidc-auth0.md)).

## Related

- [README.md](README.md) — documentation index.
- [oidc-auth0.md](oidc-auth0.md) — `aud`, M2M, refresh separation, gateway vs Redis.
- [auth-session.md](auth-session.md) — cookies, CSRF, rate limits.
