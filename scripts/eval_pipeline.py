#!/usr/bin/env python3
"""Submit test listings and report pipeline latency p50/p95 (and optional Postgres SQL)."""
from __future__ import annotations

import argparse
import json
import subprocess
import sys
import time
from pathlib import Path
import urllib.error
import urllib.request
from datetime import datetime, timezone
from statistics import median

SELLER_ID = "00000000-0000-0000-0000-000000000002"
CATEGORY_ID = 1


def _post(url: str, body: dict) -> tuple[int, dict | None]:
    data = json.dumps(body).encode()
    req = urllib.request.Request(
        url + "/v1/listings",
        data=data,
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return resp.status, json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        return e.code, None
    except urllib.error.URLError as e:
        raise ConnectionError(f"Cannot reach {url}: {e}") from e


def _get_status(url: str, listing_id: str) -> str | None:
    req = urllib.request.Request(url + "/v1/listings/" + listing_id, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            data = json.loads(resp.read().decode())
            return data.get("status")
    except (urllib.error.URLError, urllib.error.HTTPError, json.JSONDecodeError):
        return None


def _percentile(sorted_vals: list[float], p: float) -> float:
    if not sorted_vals:
        return 0.0
    k = (len(sorted_vals) - 1) * (p / 100.0)
    f = int(k)
    c = min(f + 1, len(sorted_vals) - 1)
    if f == c:
        return sorted_vals[f]
    return sorted_vals[f] + (sorted_vals[c] - sorted_vals[f]) * (k - f)


def _run_sql_since(since_iso: str, compose_dir: str) -> None:
    sql = f"""
SELECT
  COUNT(*) AS n,
  percentile_cont(0.5) WITHIN GROUP (
    ORDER BY EXTRACT(EPOCH FROM (updated_at - created_at))
  ) AS p50_sec,
  percentile_cont(0.95) WITHIN GROUP (
    ORDER BY EXTRACT(EPOCH FROM (updated_at - created_at))
  ) AS p95_sec
FROM listings
WHERE status IN ('LIVE', 'UNDER_REVIEW', 'REJECTED')
  AND created_at >= TIMESTAMPTZ '{since_iso}';
"""
    cmd = [
        "docker",
        "compose",
        "exec",
        "-T",
        "postgres",
        "psql",
        "-U",
        "postgres",
        "-d",
        "mlip",
        "-c",
        sql.strip(),
    ]
    print("\n--- Postgres (same batch, created_at filter) ---")
    try:
        subprocess.run(cmd, cwd=compose_dir, check=False)
    except FileNotFoundError:
        print("docker not found; skip SQL verification.")


def main() -> int:
    parser = argparse.ArgumentParser(description="Measure listing pipeline latency")
    parser.add_argument("--url", default="http://localhost:8080", help="listing-service base URL")
    parser.add_argument("--count", type=int, default=30, help="Number of listings to submit")
    parser.add_argument("--timeout", type=float, default=60.0, help="Max seconds to wait per listing")
    parser.add_argument("--poll", type=float, default=0.5, help="Poll interval seconds")
    parser.add_argument("--sql", action="store_true", help="Run verification SQL via docker compose")
    parser.add_argument(
        "--compose-dir",
        type=Path,
        default=Path(__file__).resolve().parents[1],
        help="Directory containing docker-compose.yml",
    )
    args = parser.parse_args()
    url = args.url.rstrip("/")
    since = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%S+00")

    print(f"Pipeline eval: {args.count} listings -> {url}")
    print(f"Batch start (UTC): {since}")
    print("Requires: warm-dev + start-dev (listing-service, analysis-worker, ml-service, fraud-service)\n")

    latencies: list[float] = []
    failures = 0

    for i in range(args.count):
        body = {
            "user_id": SELLER_ID,
            "title": f"Pipeline eval #{i} — {int(time.time())}",
            "description": "Like new. No off-platform contact. Pipeline metrics run.",
            "price_ask": 25.0 + (i % 10),
            "condition": 1 + (i % 5),
            "category_id": CATEGORY_ID,
        }
        t0 = time.perf_counter()
        try:
            status_code, created = _post(url, body)
        except ConnectionError as e:
            print(e, file=sys.stderr)
            print("Start stack: .\\scripts\\warm-dev.ps1 then .\\scripts\\start-dev.ps1", file=sys.stderr)
            return 1
        if status_code != 202 or not created or not created.get("listing_id"):
            print(f"  [{i}] POST failed: HTTP {status_code}")
            failures += 1
            continue

        listing_id = created["listing_id"]
        deadline = t0 + args.timeout
        final_status = "DRAFT"
        while time.perf_counter() < deadline:
            st = _get_status(url, listing_id)
            if st and st != "DRAFT":
                final_status = st
                break
            time.sleep(args.poll)

        elapsed = time.perf_counter() - t0
        if final_status == "DRAFT":
            print(f"  [{i}] {listing_id[:8]}... still DRAFT after {args.timeout}s")
            failures += 1
        else:
            latencies.append(elapsed)
            print(f"  [{i}] {listing_id[:8]}... -> {final_status} in {elapsed:.2f}s")

    if not latencies:
        print("\nNo completed listings. Is the full stack running?", file=sys.stderr)
        return 1

    latencies.sort()
    p50 = median(latencies)
    p95 = _percentile(latencies, 95)
    print("\n--- Client-observed (POST -> status != DRAFT) ---")
    print(f"  Completed: {len(latencies)} / {args.count} (failures/timeouts: {failures})")
    print(f"  p50: {p50:.2f}s")
    print(f"  p95: {p95:.2f}s")
    print(f"  min/max: {latencies[0]:.2f}s / {latencies[-1]:.2f}s")

    if args.sql:
        _run_sql_since(since, str(args.compose_dir))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
