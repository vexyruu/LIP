# scripts/init-dev.ps1
# One-time / after-reset setup: Docker infra, DB migrations, seeds, Pub/Sub, Neo4j, Redis scores.
$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent
Set-Location $Root

Write-Host "== MLIP init-dev ==" -ForegroundColor Cyan

function Wait-Postgres {
    param([int]$MaxAttempts = 30)
    Write-Host "Waiting for Postgres..." -ForegroundColor Yellow
    for ($i = 1; $i -le $MaxAttempts; $i++) {
        docker compose exec -T postgres pg_isready -U postgres 2>$null | Out-Null
        if ($LASTEXITCODE -eq 0) { return }
        Start-Sleep -Seconds 1
    }
    throw "Postgres did not become ready in $MaxAttempts seconds"
}

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

function Wait-Neo4j {
    param([int]$MaxAttempts = 60)
    Write-Host "Waiting for Neo4j..." -ForegroundColor Yellow
    for ($i = 1; $i -le $MaxAttempts; $i++) {
        docker compose exec -T neo4j cypher-shell -u neo4j -p changeme "RETURN 1" 2>$null | Out-Null
        if ($LASTEXITCODE -eq 0) { return }
        Start-Sleep -Seconds 2
    }
    throw "Neo4j did not become ready in $($MaxAttempts * 2) seconds"
}

function Wait-MinIO {
    param([int]$MaxAttempts = 30)
    Write-Host "Waiting for MinIO..." -ForegroundColor Yellow
    for ($i = 1; $i -le $MaxAttempts; $i++) {
        try {
            Invoke-WebRequest -Uri "http://localhost:9000/minio/health/live" -UseBasicParsing | Out-Null
            return
        } catch {
            Start-Sleep -Seconds 1
        }
    }
    throw "MinIO did not become ready in $MaxAttempts seconds"
}

function Initialize-MinIO {
    param(
        [string]$NetworkName
    )
    Write-Host "Configuring MinIO bucket..." -ForegroundColor Yellow

    docker run --rm --network $NetworkName --entrypoint /bin/sh minio/mc -c @'
mc alias set local http://minio:9000 minioadmin minioadmin &&
mc mb --ignore-existing local/mlip-listings &&
mc anonymous set download local/mlip-listings &&
cat > /tmp/minio-cors.json <<'EOF'
[
  {
    "AllowedOrigins": ["http://localhost:3000", "http://localhost:3001"],
    "AllowedMethods": ["GET", "PUT", "HEAD"],
    "AllowedHeaders": ["*"],
    "ExposeHeaders": ["ETag"],
    "MaxAgeSeconds": 3600
  }
]
EOF
mc cors set local/mlip-listings /tmp/minio-cors.json
'@ | Out-Null
}

# Start Docker
docker compose up -d
Wait-Postgres
Wait-PubSubEmulator
Wait-MinIO

$ComposeProject = (Split-Path $Root -Leaf).ToLower()
$DockerNetwork = "${ComposeProject}_default"
Initialize-MinIO -NetworkName $DockerNetwork

# Postgres migrations
$MigrationDir = Join-Path $Root "shared\go\migrations"
Get-ChildItem "$MigrationDir\*.up.sql" | Sort-Object Name | ForEach-Object {
    Write-Host "Applying $($_.Name)..." -ForegroundColor Yellow
    Get-Content $_.FullName | docker compose exec -T postgres psql -U postgres -d mlip
}

# Postgres seed
Write-Host "Seeding Postgres..." -ForegroundColor Yellow
Get-Content "$Root\scripts\seed-dev.sql" | docker compose exec -T postgres psql -U postgres -d mlip

# Pub/Sub topic + subscription
Write-Host "Creating Pub/Sub topic and subscription..." -ForegroundColor Yellow
$base = "http://localhost:8085/v1/projects/local-dev"
Invoke-RestMethod -Method Put -Uri "$base/topics/listing-created"
Invoke-RestMethod -Method Put -Uri "$base/subscriptions/listing-created-sub" `
    -ContentType "application/json" `
    -Body '{"topic":"projects/local-dev/topics/listing-created"}'

# Neo4j seed
Wait-Neo4j
Write-Host "Seeding Neo4j..." -ForegroundColor Yellow
Get-Content "$Root\shared\go\neo4j\seed.cypher" -Raw |
    docker compose exec -T neo4j cypher-shell -u neo4j -p changeme

# Fraud batch scores → Redis
Write-Host "Running fraud-batch-job (PageRank → Redis)..." -ForegroundColor Yellow
Push-Location "$Root\services\fraud-batch-job"
$env:NEO4J_URI = "bolt://localhost:7687"
$env:NEO4J_USER = "neo4j"
$env:NEO4J_PASSWORD = "changeme"
$env:REDIS_ADDR = "localhost:6379"
go run .
Pop-Location

Write-Host ""
Write-Host "Init complete. Run .\scripts\start-dev.ps1 next." -ForegroundColor Green
Write-Host "Dashboard:   http://localhost:3000/login  (mod@mlip.dev)" -ForegroundColor Green
Write-Host "Listing Lab: http://localhost:3001" -ForegroundColor Green
