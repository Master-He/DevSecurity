$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
docker compose --file (Join-Path $root "docker-compose.yml") down
