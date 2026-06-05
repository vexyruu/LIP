from __future__ import annotations

import csv
import sys
from pathlib import Path

from sklearn.metrics import classification_report, f1_score

_ROOT = Path(__file__).resolve().parents[1]
if str(_ROOT) not in sys.path:
    sys.path.insert(0, str(_ROOT))

import classifier  # noqa: E402

_FIXTURE = _ROOT / "fixtures" / "policy_eval.csv"


def main() -> int:
    if not _FIXTURE.is_file():
        print(f"Missing {_FIXTURE}", file=sys.stderr)
        return 1

    y_true: list[int] = []
    y_pred: list[int] = []
    with _FIXTURE.open(newline="", encoding="utf-8") as f:
        for row in csv.DictReader(f):
            desc = row["description"]
            label = int(row["label"])
            violation, _ = classifier.detect_violation(desc)
            y_true.append(label)
            y_pred.append(1 if violation else 0)

    n = len(y_true)
    print(f"Policy classifier evaluation ({n} labeled descriptions)")
    print(classification_report(y_true, y_pred, target_names=["clean", "violation"], digits=3))
    f1 = f1_score(y_true, y_pred)
    print(f"F1 (violation class, positive=1): {f1:.3f}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
