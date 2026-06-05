import os
import re

import pandas as pd
import spacy

# Load model once at import time.
nlp = spacy.load("en_core_web_sm")
ruler = nlp.add_pipe("entity_ruler", before="ner", config={"overwrite_ents": True})

_BASE_DIR = os.path.dirname(__file__)
_FULL_CATALOG = os.path.join(_BASE_DIR, "data", "train.tsv")
_FIXTURE_CATALOG = os.path.join(_BASE_DIR, "fixtures", "brands.tsv")
_CATALOG_PATH = _FULL_CATALOG if os.path.isfile(_FULL_CATALOG) else _FIXTURE_CATALOG

df = pd.read_csv(_CATALOG_PATH, sep="\t", usecols=["brand_name"])

# Prefer longest brand phrases so "Air Jordan" wins over "Jordan".
_brand_names: list[str] = []
_seen: set[str] = set()
for brand in sorted(df["brand_name"].dropna().astype(str), key=len, reverse=True):
    cleaned = brand.strip()
    key = cleaned.lower()
    if cleaned and key not in _seen:
        _seen.add(key)
        _brand_names.append(cleaned)

# High-signal resale brands that may be absent or fragmented in the training slice.
_priority_brands = [
    "Air Jordan",
    "New Balance",
    "Yeezy",
    "Off-White",
    "Louis Vuitton",
    "Saint Laurent",
    "s'yte",
    "S'yte",
]
for brand in reversed(_priority_brands):
    key = brand.lower()
    if key not in _seen:
        _brand_names.insert(0, brand)
        _seen.add(key)

brand_patterns = [{"label": "BRAND", "pattern": brand} for brand in _brand_names]
brand_patterns.extend(
    [
        {"label": "BRAND", "pattern": [{"LOWER": "s'yte"}]},
        {"label": "BRAND", "pattern": [{"LOWER": "syte"}]},
    ]
)
size_patterns = [
    {"label": "SIZE", "pattern": [{"LOWER": "size"}, {"IS_DIGIT": True}]},
    {"label": "SIZE", "pattern": [{"LOWER": "size"}, {"TEXT": {"REGEX": r"^\d{1,2}(\.\d)?$"}}]},
    {"label": "SIZE", "pattern": [{"LOWER": {"IN": ["xs", "s", "m", "l", "xl", "xxl"]}}]},
]
ruler.add_patterns(brand_patterns + size_patterns)

_SIZE_RE = re.compile(r"\bsize\s*(\d{1,2}(?:\.\d)?)\b", re.IGNORECASE)


def _pick_longest(entities: list[str]) -> str:
    return max(entities, key=len) if entities else ""


def _derive_product(title: str, brand: str, size: str) -> str:
    product = title.strip()
    if brand:
        product = re.sub(re.escape(brand), "", product, count=1, flags=re.IGNORECASE).strip(" -")
    if size:
        product = re.sub(re.escape(size), "", product, flags=re.IGNORECASE).strip(" -")
    product = re.sub(r"\s+", " ", product).strip()
    return product or title.strip()


def _fallback_size(text: str) -> str:
    match = _SIZE_RE.search(text)
    return match.group(1) if match else ""


def extract_entities(title: str, description: str) -> dict:
    text = f"{title} {description}".strip()
    doc = nlp(text)

    brands = [ent.text for ent in doc.ents if ent.label_ == "BRAND"]
    sizes = [ent.text for ent in doc.ents if ent.label_ == "SIZE"]

    brand = _pick_longest(brands)
    size = _pick_longest(sizes) or _fallback_size(text)
    product = _derive_product(title, brand, size)

    return {"brand": brand, "product": product, "size": size}
