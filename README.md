# MLIP - Marketplace Listing Intelligence Platform

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Python](https://img.shields.io/badge/Python-3.12-3776AB?logo=python&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-Next.js-3178C6?logo=typescript&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)
![CI](https://img.shields.io/github/actions/workflow/status/vexyruu/LIP/ci.yml?label=CI)

**MLIP** is a **learning and portfolio project** built to understand how Go services work in a polyglot microservice environment. The domain C2C listing moderation, fraud detection, and pricing gives realistic distributed systems problems without being a production product.

**The primary goal is practicing Go** (HTTP APIs, Pub/Sub workers, batch jobs, shared libraries, SQL, gRPC clients) while integrating Python (ML inference) and TypeScript (moderator UI) as separate runtimes.
---

## About this project

### Why it exists

I built LIP as a deliberate study of Go in a microservice environment: how to structure multiple Go modules under `go.work`, how a Go HTTP service owns transactions and an outbox poller while staying decoupled from analysis logic, how a Go worker handles partial failures without corrupting state, and how Go interoperates with non-Go runtimes through Protobuf/gRPC and JSON.
The ML, graph fraud, and dashboard pieces are real and runnable in which they support the learning goal.
**The main thing this repo is meant to demonstrate is designing and wiring Go-centric distributed flows**, not marketplace-scale production maturity due to the limitation of experience that i have so far.

### AI-assisted development

Parts of this codebase were built with AI coding tools (Cursor) as guidance, bootstrapping boilerplate, exploring unfamiliar APIs/Libraries, refactoring suggestions, and documentation. The Next.js UIs were scaffolded with AI, I wired API calls and local auth against the Go backend.
I use AI like a reviewer and pair programmer, not as a substitute for understanding. Architecture choices, tradeoffs, and tests are mine to validate, I run the stack locally and fix breakages. Good questions to ask me: how the **outbox** avoids dual-write loss, why the worker **fails closed** when ML is down, or how **seller risk** differs from listing price risk.

### Disclaimer

University / internship portfolio, this project is not affiliated with Mercari or any marketplace. Pricing models use the public [Mercari Price Suggestion Challenge](https://www.kaggle.com/competitions/mercari-price-suggestion-challenge) dataset for realistic training data. Deployment target is local Docker Compose, cloud infrastructure is on the roadmap.

---

## Architecture

LIP automates post-submit listing analysis: graph-based fraud scoring, ML price guidance, policy checks, and routing to **live**, **rejected**, or **human review**.

**Request flow:** `POST /v1/listings` inserts a row as `DRAFT` and writes a `listing.created` row to `pending_events` in the same transaction → outbox poller publishes to Pub/Sub → `analysis-worker` calls `fraud-service` (HTTP) for seller risk, then `ml-service` (gRPC) for NER + policy + ONNX pricing → listing updated to `LIVE`, `REJECTED`, or `UNDER_REVIEW`. If ml-service is down, the worker fails closed to `UNDER_REVIEW` and re-enqueues via the outbox (max 5 attempts).

**Stack:** Go 1.26 · Python 3.12 · TypeScript · PostgreSQL 15 · Redis 7 · Neo4j 5 + GDS · Pub/Sub emulator · MinIO · `pgx/v5` · gRPC/Protobuf · XGBoost · spaCy · Next.js

| Service | Port | Notes |
|---------|------|-------|
| `listing-service` | 8080 | Go REST API, outbox poller, uploads |
| `fraud-service` | 8081 | Go HTTP, Neo4j + Redis risk |
| `ml-service` | 50051 | Python gRPC - NER, policy, ONNX pricing |
| `analysis-worker` | - | Go Pub/Sub consumer, pipeline orchestration |
| `fraud-batch-job` | - | Go one-shot WCC + PageRank → Redis |
| `moderation-dashboard` | 3000 | Next.js moderator UI |
| `listing-lab` *(dev harness)* | 3001 | Detachable test UI - not a platform component |

---

## Getting started

**Prerequisites:** Docker Desktop, Go 1.26+, Python 3.12, Node.js 20 LTS, PowerShell 5.1+

```powershell
git clone https://github.com/vexyruu/LIP.git
cd LIP
copy .env.example .env   # review vars: start-dev.ps1 sets them per process
```

All env vars are documented in [`.env.example`](./.env.example).

**First time** (runs migrations, seeds Postgres + Neo4j, starts Docker, configures MinIO):

```powershell
.\scripts\init-dev.ps1
```

**Daily** (Pub/Sub emulator state is ephemeral, it recreate topics after every reboot):

```powershell
.\scripts\warm-dev.ps1
.\scripts\start-dev.ps1
```

| URL | Purpose |
|-----|---------|
| http://localhost:3000/login | Dashboard - login: `mod@mlip.dev` |
| http://localhost:8080 | listing-service REST API |
| http://localhost:3001 | Listing Lab dev harness (optional) |

**Stop:** `.\scripts\stop-dev.ps1` (add `-Docker` to also stop Compose services)

**Dev sellers:** clean `00000000-...0002` · fraud ring `11111111-...` · mod `22222222-...`

---

## API

Base URLs: `http://localhost:8080` (listing-service) · `http://localhost:8081` (fraud-service)

### listing-service

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/listings` | Create listing → 202, status `processing`. Body: `user_id`, `title`, `description`, `price_ask`, `condition` (1–5), `category_id`, optional `images[]`. |
| `GET` | `/v1/listings/{id}` | Detail with ML fields and `api_status`. |
| `GET` | `/v1/listings` | Queue. Required: `status`. Optional: `tier`, `sort`, `page`, `limit`. |
| `PATCH` | `/v1/listings/{id}` | Moderate `UNDER_REVIEW` listing. Body: `action`, `moderator_id`, `reason`. |
| `GET` | `/v1/analytics/summary` | Dashboard aggregates. |
| `GET` | `/v1/moderation/decisions` | Audit log. |
| `GET` | `/v1/users/{id}` | Seller profile + ban audit. |
| `POST` | `/v1/users/{id}/ban` | Ban seller, Neo4j + Postgres. |
| `POST` | `/v1/uploads` | Start MinIO upload session. |
| `POST` | `/v1/uploads/{id}/complete` | Finalize upload. |

DB statuses: `DRAFT` → `LIVE` / `UNDER_REVIEW` / `REJECTED` · API `api_status`: `processing` / `live` / `under_review` / `rejected`

### fraud-service

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/v1/risk/{user_id}` | Risk score, tier (`LOW`/`MEDIUM`/`HIGH`), velocity flags. |
| `GET` | `/v1/graph/{user_id}` | Fraud graph neighborhood. |
| `POST` | `/v1/users/{user_id}/ban` | Mark banned in Neo4j. |

### ml-service (gRPC)

`MLService.AnalyzeListing` - proto: [`shared/proto/ml_service.proto`](shared/proto/ml_service.proto)

Input: `listing_id`, `title`, `description`, `condition`, `category_id` (category path string)  
Output: `brand`, `product`, `size`, `policy_violation`, `suggested_price`, `price_lower_bound`, `price_upper_bound`

---

## Evaluation and metrics

All numbers from **local Docker Compose** and not production SLAs. Run all evals: `.\scripts\run_eval.ps1`

**Policy classifier** (rule-based, 121-row labeled fixture):

| Metric | Value |
|--------|-------|
| F1 (violation class) | **0.986** |

**Pricing : production inference path** (XGBoost/ONNX + NER brand, 10k sample, 80/20, seed 42):

| Metric | Value |
|--------|-------|
| MAE | **$10.51** |
| R² | **0.24** |
| Band coverage (± median_ae) | **50.8%** |
| vs category-median baseline | **24% lower MAE** |

**Pricing : training notebook** (XGBoost on full Kaggle split; feature path differs slightly from live NER):

| Metric | Value |
|--------|-------|
| MAE | **$9.01** · Median AE **$4.90** · Median AE as % of p50 price ≈ **28.8%** |

Note: 28.8% is median AE relative to median price, not R².

**Pipeline latency** (30 listings, clean seller, neutral descriptions, local Docker):

| p50 | p95 | Postgres p95 |
|-----|-----|-------------|
| ~2.04s | ~2.07s | ~1.94s |

Integration tests (require running stack): `go test -tags=integration ./tests/integration/...`

---

## CI

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) on push/PR: Go `test`+`vet` for all services · `pytest` for ml-service · lint+build for dashboard and listing-lab.

Local: `.\scripts\ci.ps1`

Roadmap: retraining pipeline, BigQuery analytics, Terraform/GCP infrastructure.

---

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `failed to publish listing` / outbox stuck | `.\scripts\warm-dev.ps1`, restart listing-service |
| Listing stays `DRAFT` | Restart analysis-worker or ml-service |
| All listings `REJECTED` in eval | Policy false positive, use neutral descriptions (no "contact", "line") |
| Image upload fails | `docker compose up -d`, re-run `init-dev.ps1` |
| Everything HIGH risk | Switch to clean seller `00000000-...0002` |
| Port in use | `.\scripts\stop-dev.ps1` |

---

## License

MIT - see [LICENSE](./LICENSE). Copyright (c) 2026 Faiz Adli Nugraha.
