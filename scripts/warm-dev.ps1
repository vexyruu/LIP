# scripts/warm-dev.ps1
# Run after every reboot before start-dev.ps1 — recreates ephemeral Pub/Sub resources.
$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent
Set-Location $Root

Write-Host "== MLIP warm-dev ==" -ForegroundColor Cyan

function Put-PubSubResource {
    param(
        [string]$Uri,
        [string]$Body = $null
    )
    try {
        if ($Body) {
            Invoke-RestMethod -Method Put -Uri $Uri -ContentType "application/json" -Body $Body | Out-Null
        } else {
            Invoke-RestMethod -Method Put -Uri $Uri | Out-Null
        }
    } catch {
        if ($_.Exception.Response.StatusCode.value__ -eq 409) { return }
        throw
    }
}

function Wait-PubSubEmulator {
    param([int]$MaxAttempts = 60)
    Write-Host "Waiting for Pub/Sub emulator..." -ForegroundColor Yellow
    $probe = "http://localhost:8085/v1/projects/local-dev/topics/init-probe"
    for ($i = 1; $i -le $MaxAttempts; $i++) {
        try {
            Invoke-RestMethod -Method Put -Uri $probe -ErrorAction Stop | Out-Null
            return
        } catch {
            # Any HTTP response means the emulator is listening (even 400 on bad names).
            if ($_.Exception.Response) { return }
            Start-Sleep -Seconds 1
        }
    }
    throw "Pub/Sub emulator did not become ready in $MaxAttempts seconds"
}

docker compose up -d | Out-Null
Wait-PubSubEmulator

Write-Host "Recreating Pub/Sub topic and subscription..." -ForegroundColor Yellow
$base = "http://localhost:8085/v1/projects/local-dev"
Put-PubSubResource -Uri "$base/topics/listing-created"
Put-PubSubResource -Uri "$base/subscriptions/listing-created-sub" `
    -Body '{"topic":"projects/local-dev/topics/listing-created"}'

Write-Host "Warm complete. Run .\scripts\start-dev.ps1 next." -ForegroundColor Green
