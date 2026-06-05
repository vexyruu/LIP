import numpy as np
import pytest

from pricing import _assemble_feature_vector, predict_price

SAMPLE = dict(
    title="Nike Air Max 90 size 10",
    desc="Brand new with tags, free shipping",
    condition="1",
    shipping=0.0,
    brand="Nike",
    category="Men/Shoes/Athletic",
)


def test_feature_vector_shape():
    vec = _assemble_feature_vector(**SAMPLE)
    assert vec.shape == (1, 327)
    assert vec.dtype == np.float32
    assert not np.isnan(vec).any()


def test_predict_price_returns_finite_band():
    suggested, lower, upper = predict_price(**SAMPLE)
    assert suggested > 0
    assert lower >= 0
    assert upper > suggested
    assert all(np.isfinite(v) for v in (suggested, lower, upper))


def test_predict_price_bounds():
    suggested, lower, upper = predict_price(**SAMPLE)
    assert lower <= suggested <= upper
    assert lower >= 0
    assert upper - lower == pytest.approx(9.8, rel=0.01)
