from classifier import detect_violation

def test_violation_detected():
    desc = "DM me on instagram"
    violation, conf = detect_violation(desc)
    assert violation is True
    assert conf in (0.7, 0.95)

def test_clean_listing():
    desc = "Brand new with tags, free shipping"
    violation, conf = detect_violation(desc)
    assert violation is False
    assert conf == 0.0