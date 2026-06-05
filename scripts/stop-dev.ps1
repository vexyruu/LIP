# scripts/stop-dev.ps1
# Stop MLIP dev app processes (ports) and optionally Docker infra.
param(
    [switch]$Docker
)

$ErrorActionPreference = "SilentlyContinue"

Write-Host "== MLIP stop-dev ==" -ForegroundColor Cyan

$ports = @(3000, 3001, 8080, 8081, 50051)
$stopped = 0

foreach ($port in $ports) {
    $conns = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue
    foreach ($conn in $conns) {
        $proc = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue
        if ($proc) {
            Write-Host "Stopping $($proc.ProcessName) on port $port (PID $($proc.Id))..." -ForegroundColor Yellow
            Stop-Process -Id $proc.Id -Force
            $stopped++
        }
    }
}

if ($stopped -eq 0) {
    Write-Host "No MLIP app processes found on ports $($ports -join ', ')." -ForegroundColor Green
} else {
    Write-Host "Stopped $stopped process(es)." -ForegroundColor Green
}

Write-Host ""
Write-Host "Close any remaining 'MLIP *' PowerShell windows manually if still open." -ForegroundColor Yellow

if ($Docker) {
    $Root = Split-Path $PSScriptRoot -Parent
    Set-Location $Root
    Write-Host "Stopping Docker Compose stack..." -ForegroundColor Yellow
    docker compose down
    Write-Host "Docker stopped." -ForegroundColor Green
} else {
    Write-Host "Docker left running. Use: .\scripts\stop-dev.ps1 -Docker" -ForegroundColor Cyan
}
