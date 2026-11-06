$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$archive = Join-Path $root "SAST\joern-cli-windows-x86_64.zip"
$runtimeRoot = Join-Path $root "SAST\.runtime"
$joernHome = Join-Path $runtimeRoot "joern-cli"
$marker = Join-Path $joernHome "lib\io.joern.joern-cli-4.0.674.jar"

if (Test-Path -LiteralPath $marker) {
    Write-Host "Joern 4.0.674 runtime is ready: $joernHome"
    exit 0
}
if (-not (Test-Path -LiteralPath $archive)) {
    Write-Host "找不到 Joern 发行包：$archive" -ForegroundColor Red
    Write-Host "请下载 joern-cli-windows-x86_64.zip 并放入 SAST 目录。" -ForegroundColor Yellow
    exit 1
}

New-Item -ItemType Directory -Force -Path $runtimeRoot | Out-Null
$staging = Join-Path $runtimeRoot (".extract-" + [Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Force -Path $staging | Out-Null
try {
    if (Get-Command tar.exe -ErrorAction SilentlyContinue) {
        & tar.exe -xf $archive -C $staging
        if ($LASTEXITCODE -ne 0) {
            throw "Joern ZIP 解压失败"
        }
    } else {
        Expand-Archive -LiteralPath $archive -DestinationPath $staging -Force
    }
    $stagedHome = Join-Path $staging "joern-cli"
    $stagedMarker = Join-Path $stagedHome "lib\io.joern.joern-cli-4.0.674.jar"
    if (-not (Test-Path -LiteralPath $stagedMarker)) {
        throw "Joern 发行包版本不是预期的 4.0.674"
    }
    if (Test-Path -LiteralPath $joernHome) {
        $resolvedRuntime = (Resolve-Path -LiteralPath $runtimeRoot).Path
        $resolvedJoern = (Resolve-Path -LiteralPath $joernHome).Path
        if (-not $resolvedJoern.StartsWith($resolvedRuntime + [IO.Path]::DirectorySeparatorChar)) {
            throw "拒绝替换工作区外目录：$resolvedJoern"
        }
        Remove-Item -LiteralPath $joernHome -Recurse -Force
    }
    Move-Item -LiteralPath $stagedHome -Destination $joernHome
    Write-Host "Joern 4.0.674 runtime prepared at $joernHome"
} finally {
    if (Test-Path -LiteralPath $staging) {
        Remove-Item -LiteralPath $staging -Recurse -Force
    }
}
