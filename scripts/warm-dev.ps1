# scripts/warm-dev.ps1
# Run after every reboot before start-dev.ps1 — recreates ephemeral Pub/Sub resources.
$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent
Set-Location $Root

Write-Host "== MLIP warm-dev ==" -ForegroundColor Cyan

function Wait-PubSubEmulator {
    param([int]$MaxAttempts = 30)
    Write-Host "Waiting for Pub/Sub emulator..." -ForegroundColor Yellow
    $probe = "http://localhost:8085/v1/projects/local-dev/topics/_init_probe"
    for ($i = 1; $i -le $MaxAttempts; $i++) {
        try {
            Invoke-RestMethod -Method Put -Uri $probe -ErrorAction Stop | Out-Null
            return
        } catch {
            Start-Sleep -Seconds 1
        }
    }
    throw "Pub/Sub emulator did not become ready in $MaxAttempts seconds"
}

docker compose up -d | Out-Null
Wait-PubSubEmulator

Write-Host "Recreating Pub/Sub topic and subscription..." -ForegroundColor Yellow
$base = "http://localhost:8085/v1/projects/local-dev"
Invoke-RestMethod -Method Put -Uri "$base/topics/listing-created" | Out-Null
Invoke-RestMethod -Method Put -Uri "$base/subscriptions/listing-created-sub" `
    -ContentType "application/json" `
    -Body '{"topic":"projects/local-dev/topics/listing-created"}' | Out-Null

Write-Host "Warm complete. Run .\scripts\start-dev.ps1 next." -ForegroundColor Green
