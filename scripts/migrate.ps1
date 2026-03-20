# Applies SQL migrations via the migrate service in docker-compose.yml
$ErrorActionPreference = "Stop"
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $repoRoot
docker compose run --rm migrate
