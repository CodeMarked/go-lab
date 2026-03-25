# Continuous integration (go-lab)

**Source of truth for automation:** [`.github/workflows/ci.yml`](../.github/workflows/ci.yml).

## Triggers

- **Push** to any branch
- **Pull requests** (any base)

## Jobs

### `backend-tests`

- **Runner:** `ubuntu-latest`
- **Steps:** Checkout → setup Go from `go_CRUD_api/go.mod` → `go test ./...` in `go_CRUD_api/`

### `compose-smoke`

- **Runner:** `ubuntu-latest`
- **Timeout:** 20 minutes
- **Compose files:** `docker-compose.yml` plus `docker-compose.ci.yml` when the workflow can use the actions cache (same-repo pushes and PRs from the same repo); otherwise `docker-compose.yml` only
- **Flow:**
  1. Checkout, Docker Buildx
  2. `cp .env.example .env`
  3. `docker compose build --parallel`
  4. `docker compose up -d --no-build`
  5. `docker compose run --rm migrate`
  6. Wait until `http://localhost:5000/readyz` succeeds (up to ~60s)
  7. `./scripts/test.ps1` (PowerShell smoke)
  8. On failure: dump `docker compose ps` and tail logs for backend, frontend, mysql
  9. **Always:** `docker compose down -v`

## Local parity

- Backend: `cd go_CRUD_api && go test ./...`
- Stack + smoke: Docker up, migrate, then `./scripts/test.ps1` (see repo [README](../README.md))

## Related

[MASTER_PLAN.md](MASTER_PLAN.md) · [migrations.md](migrations.md) · [README.md](README.md)
