# Database Migrations

Go-Lab uses SQL migrations in [`migrations/`](../migrations/) with `golang-migrate`. Other guides: [README.md](README.md).

## What to run

From repo root:

```powershell
docker compose run --rm migrate
```

Or:

```powershell
./scripts/migrate.ps1
```

## Rules

- API never creates/alters schema at runtime.
- Only migration files change schema.
- Run migrations before smoke tests or app rollout.

## Current migration set

- `000001_*`: base `users` table (`id`, `name`, `pennies`).
- `000002_*`: auth/session schema — `users` email/password/timestamps, `auth_sessions`, `auth_refresh_tokens`, `auth_audit_events`.
- `000003_*`: `user_identities` — maps `(issuer, subject)` → local `users.id` for OIDC / external IdPs.

## Readiness check

Set `MIGRATION_EXPECTED_VERSION` in `.env` to the **latest applied** migration version number (integer prefix of the newest `NNNNNN_*.sql` file).

Example: `MIGRATION_EXPECTED_VERSION=3` after `000003_*` is applied.

`/readyz` requires:

- DB ping succeeds
- `schema_migrations` is not dirty
- `version >= MIGRATION_EXPECTED_VERSION`

If your local DB version is ahead or behind the migration files in this branch (e.g. after a branch switch or a removed migration in history), align the database with the repo’s migration chain (backup, `migrate` up/down as appropriate, or restore) before relying on automated `migrate up`.
