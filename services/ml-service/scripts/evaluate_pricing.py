from __future__ import annotations

import argparse
import sys
from pathlib import Path

import numpy as np
import pandas as pd
from sklearn.metrics import mean_absolute_error, mean_squared_error, r2_score
from sklearn.model_selection import train_test_split

_ROOT = Path(__file__).resolve().parents[1]
if str(_ROOT) not in sys.path:
    sys.path.insert(0, str(_ROOT))

import ner  # noqa: E402
import pricing  # noqa: E402

_DEFAULT_DATA = _ROOT / "data" / "train.tsv"
_COLS = [
    "name",
    "item_condition_id",
    "category_name",
    "price",
    "shipping",
    "item_description",
]


def _shipping_value(row: pd.Series) -> float:
    s = row.get("shipping")
    if pd.isna(s):
        return 0.0
    try:
        return float(s)
    except (TypeError, ValueError):
        return 1.0 if str(s).strip() in {"1", "True", "true"} else 0.0


def _predict_row(row: pd.Series) -> tuple[float, float, float]:
    title = str(row["name"]) if pd.notna(row["name"]) else ""
    desc = str(row["item_description"]) if pd.notna(row["item_description"]) else ""
    entities = ner.extract_entities(title, desc)
    category = str(row["category_name"]) if pd.notna(row["category_name"]) else ""
    condition = row["item_condition_id"]
    shipping = _shipping_value(row)
    return pricing.predict_price(
        title=title,
        desc=desc,
        condition=condition,
        shipping=shipping,
        brand=entities.get("brand") or "",
        category=category,
    )


def _category_baseline(train: pd.DataFrame, test: pd.DataFrame) -> np.ndarray:
    medians = train.groupby("category_name")["price"].median()
    global_median = float(train["price"].median())
    return test["category_name"].map(medians).fillna(global_median).to_numpy()


def main() -> int:
    parser = argparse.ArgumentParser(description="Evaluate pricing model on held-out Mercari data")
    parser.add_argument(
        "--data",
        type=Path,
        default=_DEFAULT_DATA,
        help="Path to Mercari train.tsv (default: services/ml-service/data/train.tsv)",
    )
    parser.add_argument(
        "--sample",
        type=int,
        default=10_000,
        help="Max rows to load before split (0 = all rows; slow on full 1.4M)",
    )
    parser.add_argument("--test-size", type=float, default=0.2, help="Held-out fraction")
    parser.add_argument("--seed", type=int, default=42)
    args = parser.parse_args()

    if not args.data.is_file():
        print(
            f"Dataset not found: {args.data}\n"
            "Download Mercari Price Suggestion train.tsv from Kaggle and place at:\n"
            f"  {_DEFAULT_DATA}\n"
            "See services/ml-service/README.md — Evaluation.",
            file=sys.stderr,
        )
        return 1

    print(f"Loading {args.data} ...")
    read_kw: dict = {"sep": "\t", "usecols": _COLS}
    if args.sample and args.sample > 0:
        df = pd.read_csv(args.data, nrows=args.sample, **read_kw)
    else:
        df = pd.read_csv(args.data, **read_kw)

    df = df.dropna(subset=["price"])
    df["price"] = pd.to_numeric(df["price"], errors="coerce")
    df = df[df["price"] > 0]
    if len(df) < 100:
        print("Too few valid rows after filtering.", file=sys.stderr)
        return 1

    train_df, test_df = train_test_split(
        df, test_size=args.test_size, random_state=args.seed
    )
    print(f"Rows: {len(df)} (train {len(train_df)}, test {len(test_df)})")

    y_true = []
    y_pred = []
    lowers = []
    uppers = []
    for _, row in test_df.iterrows():
        suggested, lower, upper = _predict_row(row)
        y_true.append(float(row["price"]))
        y_pred.append(suggested)
        lowers.append(lower)
        uppers.append(upper)

    y_true_arr = np.array(y_true)
    y_pred_arr = np.array(y_pred)
    lowers_arr = np.array(lowers)
    uppers_arr = np.array(uppers)

    mae = mean_absolute_error(y_true_arr, y_pred_arr)
    rmse = float(np.sqrt(mean_squared_error(y_true_arr, y_pred_arr)))
    r2 = r2_score(y_true_arr, y_pred_arr)
    coverage = float(np.mean((lowers_arr <= y_true_arr) & (y_true_arr <= uppers_arr)))

    baseline_pred = _category_baseline(train_df, test_df)
    baseline_mae = mean_absolute_error(y_true_arr, baseline_pred)

    print("\n--- XGBoost / ONNX (held-out test, production feature path) ---")
    print(f"  MAE:            ${mae:.2f}")
    print(f"  RMSE:           ${rmse:.2f}")
    print(f"  R2:             {r2:.4f}")
    print(f"  Band coverage:  {coverage * 100:.1f}%")
    print("\n--- Baseline: category median (train) ---")
    print(f"  MAE:            ${baseline_mae:.2f}")
    print(f"  MAE improvement vs baseline: {(1 - mae / baseline_mae) * 100:.1f}%")
    print(f"\nTraining margin (median_ae in feature_config): ${pricing._cfg['median_ae']:.2f}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
