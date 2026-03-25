# Auth and sessions (platform API)

## Passwords

**Argon2id** via `golang.org/x/crypto` (`auth/password.go`). Never logged or returned. Changing `m,t,p` later needs a re-hash-on-login strategy for old hashes.

## Browser sessions

- **Login:** `POST /api/v1/auth/login` → HttpOnly session cookie (`SESSION_COOKIE_NAME`, default `gl_session`).
- **Flags:** `Secure` ← `SESSION_COOKIE_SECURE` (default **true** if unset — use **`false`** on plain HTTP / Compose per `.env.example`). `SameSite` ← `SESSION_SAMESITE` (default `Lax`).
- **TTL:** Idle (`SESSION_IDLE_TTL_SECONDS`) + absolute cap (`SESSION_ABSOLUTE_TTL_SECONDS`); enforced in `authstore` + `POST /auth/refresh`.
- **Logout:** `POST /api/v1/auth/logout` revokes session, clears session + CSRF cookies.
- **Same site:** SPA must call API on the **same origin** as the page (proxy `/api`) — cross-port localhost won’t send `SameSite=Lax` cookies. See repo README.

**CORS:** List SPA origin in `CORS_ALLOWED_ORIGINS` when `Origin` is sent.

## CSRF (cookie mutating requests)

**POST/PUT/PATCH/DELETE** without `Authorization: Bearer`: double-submit — cookie `CSRF_COOKIE_NAME` (not HttpOnly) must match header `CSRF_HEADER_NAME`.

**Skipped:** Bearer present; safe methods; exempt auth paths (`/auth/register`, `/login`, `/token`, `/bootstrap`); logout with no session cookie.

**CSRF cookie:** Set on login + refresh; cleared on logout + password change. **`GET /api/v1/auth/csrf`** — refresh CSRF when session exists.

CORS allows the CSRF header name; credentialed responses use a concrete origin, never `*`.

## Abuse / rate limits

Per-IP **1-minute** windows (separate limiters):

| Route | ~ / min |
|-------|---------|
| register | 15 |
| login | 30 |
| logout | 60 |
| refresh | 120 |
| GET csrf | 60 |
| change-password | 10 |
| token, bootstrap | 30 |

**Lockout:** After **5** failed logins for a **known** email → **15 min** block. **`REDIS_URL` unset** = in-memory per process; **set** = shared across replicas. **Redis errors → fail open** + log.

**Edge vs app:** Coarse IP limits at **nginx/WAF**; email lockout + route limits in **app** (optional Redis). Don’t duplicate the same cap twice — see [oidc-auth0.md](oidc-auth0.md).

## OIDC Bearer

If **`OIDC_ISSUER_URL`** + **`OIDC_AUDIENCE`** set → RS256 JWT after HS256 fails; **`user_identities`**. [oidc-auth0.md](oidc-auth0.md). Linking policy: [adr-account-linking.md](adr-account-linking.md).

## Change password

`POST /api/v1/auth/change-password` — needs `user:<id>` session or Bearer; revokes **all** sessions.

## JWT / service clients

`POST /api/v1/auth/token` mints JWTs. `/users` accepts **Bearer or session**; Bearer wins if both sent. Rotation: [jwt-rotation.md](jwt-rotation.md).

## Bootstrap (temporary)

`POST /api/v1/auth/bootstrap` — CORS `Origin` required; `data.bootstrap` marks deprecation. [bootstrap-sunset.md](bootstrap-sunset.md).

## Desktop

[desktop-auth-bridge.md](desktop-auth-bridge.md).

## Schema

`000002_*`: `users` auth cols, `auth_sessions`, `auth_refresh_tokens` (reserved), `auth_audit_events`.

## Related

[Index](README.md) · optional **`JWT_ACTIVE_KEY_ID`** log — [jwt-rotation.md](jwt-rotation.md)
