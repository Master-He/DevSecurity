$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot

function New-RandomSecret([int]$Bytes = 32) {
    $buffer = New-Object byte[] $Bytes
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($buffer)
    return [Convert]::ToBase64String($buffer).Replace("+", "-").Replace("/", "_").TrimEnd("=")
}

$envFile = Join-Path $root ".env"
if (-not (Test-Path -LiteralPath $envFile)) {
    $lines = @(
        "APP_ADMIN_USERNAME=admin",
        "APP_ADMIN_PASSWORD=$(New-RandomSecret 24)",
        "JWT_SECRET=$(New-RandomSecret 48)",
        "SCA_SERVICE_TOKEN=$(New-RandomSecret 32)",
        "POSTGRES_DB=devsecurity",
        "POSTGRES_USER=devsecurity",
        "POSTGRES_PASSWORD=$(New-RandomSecret 24)",
        "S3_ACCESS_KEY=devsecurity",
        "S3_SECRET_KEY=$(New-RandomSecret 24)",
        "S3_BUCKET=devsecurity",
        "MINIO_ROOT_USER=devsecurity",
        "MINIO_ROOT_PASSWORD=$(New-RandomSecret 24)",
        "GIT_ALLOWED_HOSTS=localhost,gitlab.com",
        "GRAFANA_ADMIN_PASSWORD=$(New-RandomSecret 24)"
    )
    Set-Content -LiteralPath $envFile -Value $lines -Encoding utf8NoBOM
    Write-Host "Created .env with random local secrets."
    Write-Host "Admin username: admin"
    Write-Host "Admin password: $((Get-Content -LiteralPath $envFile | Where-Object { $_ -like 'APP_ADMIN_PASSWORD=*' }) -replace '^APP_ADMIN_PASSWORD=', '')"
}

$scaTokenLine = Get-Content -LiteralPath $envFile | Where-Object { $_ -like 'SCA_SERVICE_TOKEN=*' }
if (-not $scaTokenLine) {
    Add-Content -LiteralPath $envFile -Value "SCA_SERVICE_TOKEN=$(New-RandomSecret 32)" -Encoding utf8NoBOM
    Write-Host "Added SCA_SERVICE_TOKEN to .env."
}

& docker version --format '{{.Server.Version}}' *> $null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Docker Engine 不可用。请先启动 Docker Desktop，并等待 Engine 显示为 Running。" -ForegroundColor Red
    exit 1
}

$composeFile = Join-Path $root "docker-compose.yml"
& docker compose --file $composeFile --env-file $envFile config --quiet
if ($LASTEXITCODE -ne 0) {
    Write-Host "Docker Compose 配置校验失败，请查看上面的配置错误。" -ForegroundColor Red
    exit $LASTEXITCODE
}

& docker compose --file $composeFile --env-file $envFile up --build
$composeExitCode = $LASTEXITCODE
if ($composeExitCode -ne 0) {
    Write-Host "" 
    Write-Host "Docker Compose 启动失败，退出码：$composeExitCode" -ForegroundColor Red
    Write-Host "如果错误包含 failed to resolve reference、TLS handshake 或 EOF，请检查：" -ForegroundColor Yellow
    Write-Host "  1. Docker Desktop > Settings > Docker Engine 中的 registry-mirrors"
    Write-Host "  2. Docker Desktop > Settings > Resources > Proxies 中的代理配置"
    Write-Host "  3. 尝试执行 docker pull postgres:16-alpine 验证镜像仓库连通性"
    $mirrors = & docker info --format '{{json .RegistryConfig.Mirrors}}' 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "当前 registry-mirrors: $mirrors"
    }
}
exit $composeExitCode
