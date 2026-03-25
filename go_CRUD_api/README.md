# go_CRUD_api (platform API)

Gin + MySQL; JSON envelopes; JWT + cookie sessions.

```bash
cd go_CRUD_api && go test ./...
```

Compose builds from repo root — [../README.md](../README.md).

**Auth:** Argon2id (`auth/password.go`); sessions `authstore` + `000002_*`. CSRF on cookie mutations; per-IP + per-email limits; optional **`REDIS_URL`**. Env: [`.env.example`](../.env.example). Docs: [docs/README.md](../docs/README.md).
