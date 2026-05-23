$env:GCP_PROJECT_ID = "local-dev"
$env:PUBSUB_EMULATOR_HOST = "localhost:8085"
$env:POSTGRES_CONN = "postgres://postgres:postgres@localhost:5432/mlip"
$env:PUBSUB_SUBSCRIPTION = "listing-created-sub"
$env:FRAUD_SERVICE_URL = "http://localhost:8081"
$env:ML_SERVICE_ADDR = "localhost:50051"
Set-Location "$PSScriptRoot\..\services\analysis-worker"
if (Test-Path ".\analysis-worker.exe") {
    .\analysis-worker.exe
} else {
    go run .
}
