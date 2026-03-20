# Database Migrations

Go-Lab uses SQL migrations in [`migrations/`](../migrations/) with `golang-migrate`.

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

- `000001_*`: creates/drops `users`.
- `000002_*`: pending/no-op marker (kept for version continuity).

## Readiness check

Set `MIGRATION_EXPECTED_VERSION` in `.env` to latest version number.

Example: `MIGRATION_EXPECTED_VERSION=2` for `000002_*`.

`/readyz` requires:

- DB ping succeeds
- `schema_migrations` is not dirty
- `version >= MIGRATION_EXPECTED_VERSION`
