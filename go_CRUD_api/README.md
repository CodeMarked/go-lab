# platform-api (go_CRUD_api)

Go HTTP API (Gin) with JSON envelopes, MySQL, JWT + cookie session auth.

## Run / test

```bash
# from repo root: load .env with DB_* and secrets, then:
cd go_CRUD_api && go test ./...
```

Docker Compose builds this module from the repo root; see [../README.md](../README.md).

## Auth notes

- Passwords: Argon2id (`auth/password.go`). Session persistence: `authstore` + migration `000002_*`.
- CSRF double-submit for cookie mutating routes (`middleware/csrf.go`); per-IP rate limits + per-email login lockout (`middleware/ratelimit.go`, `myhandlers/login_throttle.go`).
- Env vars: [`.env.example`](../.env.example); guides: [auth-session.md](../docs/auth-session.md), [jwt-rotation.md](../docs/jwt-rotation.md).
