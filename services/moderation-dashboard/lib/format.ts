export function shortUUID(id: string): string {
  if (id.length < 12) return id;
  return `${id.slice(0, 4)}…${id.slice(-4)}`;
}

export function formatPrice(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(amount);
}

export function formatRelativeTime(iso: string): string {
  const date = new Date(iso);
  const seconds = Math.floor((Date.now() - date.getTime()) / 1000);
  if (seconds < 60) return `${seconds} sec ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes} min ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} hr ago`;
  const days = Math.floor(hours / 24);
  return `${days} day${days === 1 ? "" : "s"} ago`;
}

export function tierBadgeClass(tier: string | null): string {
  switch (tier?.toUpperCase()) {
    case "HIGH":
      return "text-error border-error/20";
    case "MEDIUM":
      return "text-on-surface-variant border-white/10";
    case "LOW":
      return "text-primary-container border-primary-container/20";
    default:
      return "text-on-surface-variant border-white/10";
  }
}

export function tierScoreClass(tier: string | null): string {
  switch (tier?.toUpperCase()) {
    case "HIGH":
      return "text-error";
    case "LOW":
      return "text-primary-container";
    default:
      return "text-white/80";
  }
}

export function listingShortId(id: string): string {
  const hex = id.replace(/-/g, "").slice(0, 8).toUpperCase();
  return `${hex.slice(0, 4)}-${hex.slice(4, 8)}`;
}

export function conditionLabel(condition: number): string {
  const labels: Record<number, string> = {
    1: "New",
    2: "Like New",
    3: "Good",
    4: "Fair",
    5: "Poor",
  };
  const name = labels[condition] ?? "Unknown";
  return `${condition}/5 (${name})`;
}

export function riskTierLabel(tier: string | null): string {
  switch (tier?.toUpperCase()) {
    case "HIGH":
      return "CRITICAL RISK TIER";
    case "MEDIUM":
      return "ELEVATED RISK TIER";
    case "LOW":
      return "LOW RISK TIER";
    default:
      return "UNSCORED RISK TIER";
  }
}

export function policyViolationLabel(violation: boolean | null): string {
  if (violation === true) return "DETECTED";
  if (violation === false) return "NONE DETECTED";
  return "UNKNOWN";
}

export function priceMarkerPercent(
  ask: number,
  lower: number | null,
  upper: number | null
): number {
  if (lower == null || upper == null || upper <= lower) return 50;
  const pct = ((ask - lower) / (upper - lower)) * 100;
  return Math.min(100, Math.max(0, pct));
}
