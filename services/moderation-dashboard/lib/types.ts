export type ListingQueueItem = {
  listing_id: string;
  title: string;
  user_id: string;
  price_ask: number;
  risk_score: number | null;
  risk_tier: string | null;
  category_id: number;
  images: string[];
  created_at: string;
};

export type ListListingsResponse = {
  listings: ListingQueueItem[];
  total: number;
  page: number;
  limit: number;
};

export type ListingDetail = {
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
};

export type TierFilter = "ALL" | "HIGH" | "MEDIUM" | "LOW";

export type SortOption = "risk_score" | "created_at";

export type ModerationDecision = {
  id: string;
  timestamp: string;
  listing_id: string;
  title: string;
  action: string;
  risk_tier: string | null;
  reason: string;
  moderator_id: string;
};

export type ListModerationDecisionsResponse = {
  decisions: ModerationDecision[];
  total: number;
  page: number;
  limit: number;
};

export type RecentModerationDecision = Pick<
  ModerationDecision,
  "timestamp" | "listing_id" | "title" | "action" | "risk_tier" | "reason"
>;

export type AnalyticsSummary = {
  total_listings: number;
  under_review: number;
  by_status: Record<string, number>;
  by_risk_tier: Record<string, number>;
  policy_rejections: number;
  avg_pipeline_seconds: number | null;
  last_24h_submissions: number;
  daily_submissions: {
    date: string;
    total: number;
    high_risk: number;
  }[];
  recent_decisions: RecentModerationDecision[];
};

export type UserBanAudit = {
  reason: string;
  moderator_id: string;
  listings_flagged: number;
  created_at: string;
};

export type UserProfile = {
  user_id: string;
  email: string;
  display_name: string;
  status: string;
  created_at: string;
  ban_audit?: UserBanAudit | null;
};

export type GraphNode = {
  id: string;
  type: string;
  label: string;
  banned?: boolean;
  risk_score?: number;
};

export type GraphEdge = {
  source: string;
  target: string;
  type: string;
};

export type UserGraph = {
  user_id: string;
  nodes: GraphNode[];
  edges: GraphEdge[];
};

export type BanUserResponse = {
  user_id: string;
  status: string;
  listings_flagged: number;
};

export type SellerRisk = {
  user_id: string;
  risk_score: number;
  risk_tier: string;
  component_score: number;
  pagerank_proximity: number;
  velocity_count: number;
  velocity_score: number;
  velocity_flags: string[];
  cache_hit: boolean;
  computed_at?: string;
};

export type UserListingItem = {
  listing_id: string;
  title: string;
  price_ask: number;
  status: string;
  risk_score: number | null;
  risk_tier: string | null;
  images: string[];
  created_at: string;
};

export type ListUserListingsResponse = {
  listings: UserListingItem[];
  total: number;
  page: number;
  limit: number;
};