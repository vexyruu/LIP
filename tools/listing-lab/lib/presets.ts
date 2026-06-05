import type { CreateListingRequest } from "./types";

/** Default Listing Lab seller — no fraud graph links (LOW/MEDIUM risk). */
export const SELLER_ID = "00000000-0000-0000-0000-000000000002";

/** Seeded fraud-ring seller — always HIGH risk (~0.9) for queue demos. */
export const FRAUD_SELLER_ID = "11111111-1111-1111-1111-111111111111";

export const SAMPLE_IMAGES = [
  "https://images.unsplash.com/photo-1556906781-9a412961c28c?auto=format&fit=crop&w=900&q=80",
  "https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=900&q=80",
];

export interface Preset {
  id: string;
  label: string;
  description: string;
  data: CreateListingRequest;
}

export const DEFAULT_FORM: CreateListingRequest = {
  user_id: SELLER_ID,
  title: "Air Jordan 1 Retro High OG",
  description: "Size 10. Worn twice. Comes with original box. Chicago colorway.",
  price_ask: 285,
  condition: 2,
  category_id: 1,
  images: [SAMPLE_IMAGES[0]],
};

export const PRESETS: Preset[] = [
  {
    id: "jordan",
    label: "Jordan 1",
    description: "Standard sneaker listing (clean seller).",
    data: DEFAULT_FORM,
  },
  {
    id: "syte",
    label: "S'yte jacket",
    description: "Streetwear listing at retail-ish price (clean seller).",
    data: {
      user_id: SELLER_ID,
      title: "S'yte COTTON/LINEN FLOWER EMBROIDERY JACKET",
      description: "Size M. Excellent condition. Retail was around $700.",
      price_ask: 700,
      condition: 2,
      category_id: 1,
      images: [],
    },
  },
  {
    id: "low-price",
    label: "Low ask (pricing)",
    description: "Absurdly low price — tests XGBoost pricing bounds, not fraud tier.",
    data: {
      ...DEFAULT_FORM,
      title: "Air Jordan 1 Retro — must sell today",
      description: "Brand new in box. Need cash fast. Serious buyers only.",
      price_ask: 12,
    },
  },
  {
    id: "policy",
    label: "Policy violation",
    description: "Off-platform contact → auto REJECTED.",
    data: {
      ...DEFAULT_FORM,
      title: "Yeezy Boost 350 V2",
      description:
        "WhatsApp me at 555-0199 for faster response. PayPal friends only, no returns.",
      price_ask: 180,
      images: [SAMPLE_IMAGES[1]],
    },
  },
  {
    id: "fraud-seller",
    label: "Fraud-ring seller",
    description: "Same listing but seller 1111… — expect HIGH tier + queue.",
    data: {
      ...DEFAULT_FORM,
      user_id: FRAUD_SELLER_ID,
    },
  },
];
