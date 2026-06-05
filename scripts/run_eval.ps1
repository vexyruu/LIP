# Run MLIP evaluation scripts (classifier, pricing, pipeline).
# Pricing requires data/train.tsv; pipeline requires start-dev stack.
$ErrorActionPreference = "Stop"
$Root = Split-Path $PSScriptRoot -Parent
Set-Location $Root

Write-Host "== 1/3 Policy classifier (fixtures) ==" -ForegroundColor Cyan
Push-Location "$Root\services\ml-service"
python scripts/evaluate_classifier.py
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

Write-Host "`n== 2/3 Pricing model (held-out) ==" -ForegroundColor Cyan
Push-Location "$Root\services\ml-service"
if (Test-Path "data\train.tsv") {
    python scripts/evaluate_pricing.py --sample 10000
} else {
    Write-Host "SKIP: place Kaggle train.tsv at services/ml-service/data/train.tsv" -ForegroundColor Yellow
}
Pop-Location

Write-Host "`n== 3/3 Pipeline latency (30 listings) ==" -ForegroundColor Cyan
python scripts/eval_pipeline.py --count 30 --sql

Write-Host "`nDone. Paste metrics into resume / README.md (Evaluation and metrics)" -ForegroundColor Green
