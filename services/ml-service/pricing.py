"""XGBoost pricing inference via ONNX Runtime (see models/pricing_model.onnx)."""
import os
import re
import joblib
import numpy as np
import onnxruntime as ort
from scipy.sparse import hstack

_MODELS_DIR = os.path.join(os.path.dirname(__file__), "models")


def _load_artifacts():
    cfg = joblib.load(os.path.join(_MODELS_DIR, "feature_config.pkl"))
    assert len(cfg["numeric_features"]) == 9
    assert len(cfg["flag_specs"]) == 18
    assert len(cfg["svd_features"]) == 327

    session = ort.InferenceSession(
        os.path.join(_MODELS_DIR, "pricing_model.onnx"),
        providers=["CPUExecutionProvider"],
    )
    tfidf_name = joblib.load(os.path.join(_MODELS_DIR, "tfidf_name.pkl"))
    tfidf_desc = joblib.load(os.path.join(_MODELS_DIR, "tfidf_desc.pkl"))
    svd = joblib.load(os.path.join(_MODELS_DIR, "svd.pkl"))
    return session, tfidf_name, tfidf_desc, svd, cfg


_onnx_session, _tfidf_name, _tfidf_desc, _svd, _cfg = _load_artifacts()
_ONNX_INPUT_NAME = _onnx_session.get_inputs()[0].name
_N_SVD = int(_cfg["n_svd_components"])


def _lookup_enc(mapping: dict, key: str) -> float:
    key = (key or "").strip()
    if not key:
        return float(_cfg["global_mean"])
    return float(mapping.get(key, _cfg["global_mean"]))


def _split_category(category: str) -> tuple[str, str, str]:
    if not category or "/" not in category:
        return "", "", ""
    parts = category.split("/")
    cat1 = parts[0] if len(parts) > 0 else ""
    cat2 = parts[1] if len(parts) > 1 else ""
    cat3 = parts[2] if len(parts) > 2 else ""
    return cat1, cat2, cat3


def _parse_condition(condition) -> float:
    try:
        return float(int(condition))
    except (TypeError, ValueError):
        return 0.0


def _build_numeric_dict(
    title: str,
    desc: str,
    condition,
    shipping: float,
    brand: str,
    category: str,
) -> dict[str, float]:
    cat1, cat2, cat3 = _split_category(category)
    brand = brand or ""
    return {
        "item_condition_id": _parse_condition(condition),
        "shipping": float(shipping),
        "brand_enc": _lookup_enc(_cfg["brand_map"], brand),
        "cat1_enc": _lookup_enc(_cfg["cat1_map"], cat1),
        "cat2_enc": _lookup_enc(_cfg["cat2_map"], cat2),
        "cat3_enc": _lookup_enc(_cfg["cat3_map"], cat3),
        "name_len": float(len(title or "")),
        "desc_len": float(len(desc or "")),
        "has_brand": 1.0 if brand and brand != "Unknown" else 0.0,
    }


def _build_flags_dict(title: str, desc: str) -> dict[str, float]:
    text = f"{title or ''} {desc or ''}".lower()
    return {
        name: 1.0 if re.search(pattern, text) else 0.0
        for name, pattern in _cfg["flag_specs"].items()
    }


def _build_svd_dict(title: str, desc: str) -> dict[str, float]:
    name_sp = _tfidf_name.transform([title or ""])
    desc_sp = _tfidf_desc.transform([desc or ""])
    combined = hstack([name_sp, desc_sp])
    svd_row = _svd.transform(combined)
    return {f"svd_{i}": float(svd_row[0, i]) for i in range(_N_SVD)}


def _assemble_feature_vector(
    title: str,
    desc: str,
    condition,
    shipping: float,
    brand: str,
    category: str,
) -> np.ndarray:
    feature_dict = {
        **_build_numeric_dict(title, desc, condition, shipping, brand, category),
        **_build_flags_dict(title, desc),
        **_build_svd_dict(title, desc),
    }
    row = [feature_dict[name] for name in _cfg["svd_features"]]
    assert len(row) == 327
    return np.array([row], dtype=np.float32)


def predict_price(
    title: str,
    desc: str,
    condition,
    shipping: float,
    brand: str,
    category: str,
) -> tuple[float, float, float]:
    x = _assemble_feature_vector(title, desc, condition, shipping, brand, category)
    log_price = float(_onnx_session.run(None, {_ONNX_INPUT_NAME: x})[0].reshape(-1)[0])

    suggested = min(float(np.expm1(log_price)), float(_cfg["p99_cap"]))
    margin = float(_cfg["median_ae"])
    lower = max(0.0, suggested - margin)
    upper = suggested + margin
    return round(suggested, 2), round(lower, 2), round(upper, 2)
