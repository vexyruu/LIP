export interface CreateListingRequest {
  user_id: string;
  title: string;
  description: string;
  price_ask: number;
  condition: number;
  category_id: number;
  images: string[];
}

export interface CreateListingResponse {
  listing_id: string;
  status: string;
  eta_ms: number;
}

export interface ListingStatus {
  listing_id: string;
  user_id: string;
  title: string;
  description: string;
  price_ask: number;
  condition: number;
  category_id: number;
  status: string;
  suggested_price: number | null;
  price_lower_bound: number | null;
  price_upper_bound: number | null;
  risk_tier: string | null;
  risk_score: number | null;
  extracted_brand: string | null;
  extracted_product: string | null;
  extracted_size: string | null;
  policy_violation: boolean | null;
  images: string[];
  created_at: string | null;
  updated_at: string | null;
}

export interface HistoryEntry {
  listing_id: string;
  title: string;
  status: string;
  submitted_at: string;
}

export type SubmitPhase =
  | "idle"
  | "submitting"
  | "polling"
  | "done"
  | "error"
  | "timeout";
