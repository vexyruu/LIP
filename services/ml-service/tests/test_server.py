import ml_service_pb2
from server import MLServicer

SAMPLE = dict(
    listing_id="test-listing-1",
    title="Nike Air Max 90 size 10",
    description="Brand new with tags, free shipping",
    condition="1",
    category_id="Men/Shoes/Athletic",
)

POLICY_SAMPLE = dict(
    listing_id="test-listing-2",
    title="Yeezy Boost 350",
    description="DM me on instagram for faster response",
    condition="2",
    category_id="Men/Shoes/Athletic",
)


def test_analyze_listing_happy_path():
    servicer = MLServicer()
    req = ml_service_pb2.ListingRequest(**SAMPLE)
    resp = servicer.AnalyzeListing(req, None)

    assert resp.brand
    assert resp.suggested_price > 0
    assert resp.price_lower_bound >= 0
    assert resp.price_upper_bound >= resp.suggested_price
    assert resp.policy_violation is False


def test_analyze_listing_policy_violation():
    servicer = MLServicer()
    req = ml_service_pb2.ListingRequest(**POLICY_SAMPLE)
    resp = servicer.AnalyzeListing(req, None)

    assert resp.policy_violation is True
    assert resp.suggested_price > 0


def test_analyze_listing_ner_fields():
    servicer = MLServicer()
    req = ml_service_pb2.ListingRequest(
        listing_id="test-listing-3",
        title="Air Jordan 1 Retro High OG",
        description="Size 10. Worn twice.",
        condition="2",
        category_id="Men/Shoes/Athletic",
    )
    resp = servicer.AnalyzeListing(req, None)

    assert "Jordan" in resp.brand or resp.brand == "Air Jordan"
    assert resp.product
    assert resp.size
