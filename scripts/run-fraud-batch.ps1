# Recompute PageRank / WCC fraud scores in Redis (same as init-dev tail).
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot

Write-Host "Running fraud-batch-job (PageRank -> Redis)..." -ForegroundColor Yellow
Push-Location "$Root\services\fraud-batch-job"
$env:NEO4J_URI = "bolt://localhost:7687"
$env:NEO4J_USER = "neo4j"
$env:NEO4J_PASSWORD = "changeme"
$env:REDIS_ADDR = "localhost:6379"
go run .
Pop-Location

Write-Host "Fraud batch complete." -ForegroundColor Green
