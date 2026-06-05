import ner


def test_air_jordan_listing():
    result = ner.extract_entities(
        "Air Jordan 1 Retro High OG",
        "Size 10. Worn twice. Comes with original box. Chicago colorway.",
    )
    assert result["brand"] == "Air Jordan"
    assert "1" in result["product"]
    assert result["size"] in {"Size 10", "10"}


def test_nike_listing():
    result = ner.extract_entities(
        "Nike boy sneakers",
        "Size 6Y. High tops. Super light.",
    )
    assert result["brand"] == "Nike"
    assert result["product"]


def test_no_false_jordan_on_air_jordan_title():
    result = ner.extract_entities(
        "Air Jordan 1 Retro High OG",
        "Gently used.",
    )
    assert result["brand"] != "Jordan"


def test_syte_brand():
    result = ner.extract_entities(
        "S'yte COTTON/LINEN FLOWER EMBROIDERY JACKET",
        "Size M. Excellent condition.",
    )
    assert result["brand"].lower() in {"s'yte", "syte"}
