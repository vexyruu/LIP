# scripts/start-dev.ps1
# Daily dev: set env vars and launch each service in its own terminal window.
$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent

Write-Host "== MLIP start-dev ==" -ForegroundColor Cyan
Write-Host "Warming Pub/Sub (required after reboot)..." -ForegroundColor Yellow
& "$PSScriptRoot\warm-dev.ps1"
Write-Host ""
Write-Host "Starting 6 services in separate windows..." -ForegroundColor Yellow
Write-Host "First time? Run .\scripts\init-dev.ps1 once." -ForegroundColor Yellow
Write-Host ""

function Start-MlipService {
    param(
        [string]$Title,
        [string]$WorkDir,
        [string]$Command
    )
    Start-Process powershell -ArgumentList @(
        "-NoExit", "-Command",
        "cd '$WorkDir'; `$Host.UI.RawUI.WindowTitle = '$Title'; $Command"
    )
}

# 1. ml-service (Python 3.12 venv, gRPC :50051)
Start-MlipService "MLIP ml-service" "$Root\services\ml-service" `
    "if (Test-Path '.\.venv\Scripts\Activate.ps1') { .\.venv\Scripts\Activate.ps1 } elseif (Test-Path '.\venv\Scripts\Activate.ps1') { .\venv\Scripts\Activate.ps1 } else { Write-Host 'No Python venv found. Run: python -m venv .venv; pip install -r requirements.txt' -ForegroundColor Red; exit 1 }; python server.py"

# 2. fraud-service :8081
Start-MlipService "MLIP fraud-service" "$Root\services\fraud-service" `
    "`$env:PORT='8081'; `$env:REDIS_ADDR='localhost:6379'; `$env:NEO4J_URI='bolt://localhost:7687'; `$env:NEO4J_USER='neo4j'; `$env:NEO4J_PASSWORD='changeme'; go run ."

# 3. listing-service :8080
Start-MlipService "MLIP listing-service" "$Root\services\listing-service" `
    "`$env:PORT='8080'; `$env:POSTGRES_CONN='postgres://postgres:postgres@localhost:5432/mlip'; `$env:GCP_PROJECT_ID='local-dev'; `$env:PUBSUB_EMULATOR_HOST='localhost:8085'; `$env:PUBSUB_TOPIC_LISTING_CREATED='listing-created'; `$env:FRAUD_SERVICE_URL='http://localhost:8081'; `$env:S3_ENDPOINT='http://localhost:9000'; `$env:S3_ACCESS_KEY='minioadmin'; `$env:S3_SECRET_KEY='minioadmin'; `$env:S3_BUCKET='mlip-listings'; `$env:S3_USE_SSL='false'; `$env:UPLOAD_PUBLIC_BASE_URL='http://localhost:9000/mlip-listings'; `$env:ALLOW_EXTERNAL_IMAGE_URLS='true'; go run ."

# 4. analysis-worker
Start-MlipService "MLIP analysis-worker" "$Root\services\analysis-worker" `
    "`$env:GCP_PROJECT_ID='local-dev'; `$env:PUBSUB_EMULATOR_HOST='localhost:8085'; `$env:POSTGRES_CONN='postgres://postgres:postgres@localhost:5432/mlip'; `$env:PUBSUB_SUBSCRIPTION='listing-created-sub'; `$env:FRAUD_SERVICE_URL='http://localhost:8081'; `$env:ML_SERVICE_ADDR='localhost:50051'; go run ."

# 5. moderation-dashboard :3000
Start-MlipService "MLIP dashboard" "$Root\services\moderation-dashboard" `
    "npm run dev"

# 6. listing-lab :3001 (dev-only seller simulator)
Start-MlipService "MLIP listing-lab" "$Root\tools\listing-lab" `
    "if (-not (Test-Path '.\node_modules')) { npm install }; npm run dev"

Write-Host ""
Write-Host "All services launched." -ForegroundColor Green
Write-Host "  Dashboard:   http://localhost:3000/login" -ForegroundColor Green
Write-Host "  Listing Lab: http://localhost:3001" -ForegroundColor Green
Write-Host "  Listing API: http://localhost:8080/v1/listings?status=UNDER_REVIEW" -ForegroundColor Green
Write-Host "  Fraud API (demo): http://localhost:8081/v1/risk/11111111-1111-1111-1111-111111111111" -ForegroundColor Green
Write-Host ""
Write-Host "Login: mod@mlip.dev" -ForegroundColor Cyan
Write-Host "Clean seller UUID: 00000000-0000-0000-0000-000000000002" -ForegroundColor Cyan
