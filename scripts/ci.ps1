# Run the same checks as .github/workflows/ci.yml locally.
$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent
Set-Location $Root

Write-Host "== Go ==" -ForegroundColor Cyan
go test ./shared/go/...
go test ./services/listing-service/...
go test ./services/fraud-service/...
go test ./services/analysis-worker/...
go test ./services/fraud-batch-job/...
go vet ./shared/go/...
go vet ./services/listing-service/...
go vet ./services/fraud-service/...
go vet ./services/analysis-worker/...
go vet ./services/fraud-batch-job/...

Write-Host "== Python ==" -ForegroundColor Cyan
Push-Location "$Root\services\ml-service"
if (-not (Test-Path ".\.venv\Scripts\Activate.ps1")) {
    Write-Host "Create venv first: python -m venv .venv; pip install -r requirements.txt" -ForegroundColor Red
    exit 1
}
.\.venv\Scripts\Activate.ps1
pip install -r requirements.txt -q
python -m spacy download en_core_web_sm
pytest tests/ -v
Pop-Location

Write-Host "== Dashboard ==" -ForegroundColor Cyan
Push-Location "$Root\services\moderation-dashboard"
npm ci
npm run lint
npm run build
Pop-Location

Write-Host "== Listing Lab ==" -ForegroundColor Cyan
Push-Location "$Root\tools\listing-lab"
npm ci
npm run lint
npm run build
Pop-Location

Write-Host "All CI checks passed." -ForegroundColor Green
