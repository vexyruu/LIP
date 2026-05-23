import spacy, pandas as pd, os

# Load model
nlp = spacy.load("en_core_web_sm")

# Add entity ruler to the pipeline
ruler = nlp.add_pipe("entity_ruler", before="ner")

# Read brand and size data
_DATA_PATH = os.path.join(os.path.dirname(__file__), "data", "train.tsv")
df = pd.read_csv(_DATA_PATH, sep="\t", usecols=["brand_name"])
brands = df["brand_name"].dropna().unique().tolist()

# Create patterns for brand and size
brand_patterns = [{"label": "BRAND", "pattern": brand} for brand in brands]
size_patterns = [
                {"label": "SIZE", "pattern": [{"LOWER": "size"}, {"IS_DIGIT": True}]},
                {"label": "SIZE", "pattern": [{"LOWER": {"IN": ["xs", "s", "m", "l", "xl", "xxl"]}}]},
                ]

# Add patterns to the entity ruler
ruler.add_patterns(brand_patterns + size_patterns)

# Extract entities from title and description
def extract_entities(title: str, description:str) -> dict:
    text = title + " " + description
    doc = nlp(text)
    
    brand, product, size = "", "", ""
    for ent in doc.ents:
        if ent.label_ == "BRAND" and not brand:
            brand = ent.text
        elif ent.label_ == "SIZE" and not size:
            size = ent.text
        elif ent.label_ == "PRODUCT" and not product:
            product = ent.text
    return {"brand": brand, "product": product, "size": size}
                