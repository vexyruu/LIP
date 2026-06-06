import type { ListingQueueItem } from "@/lib/types";

function countTier(listings: ListingQueueItem[], tier: string): number {
  return listings.filter((l) => l.risk_tier?.toUpperCase() === tier).length;
}

function oldestWait(listings: ListingQueueItem[]): string {
  if (listings.length === 0) return "—";
  const oldest = listings.reduce((acc, l) => {
    const t = new Date(l.created_at).getTime();
    return t < acc ? t : acc;
  }, Date.now());
  const minutes = Math.floor((Date.now() - oldest) / 60000);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;
  return `${Math.floor(hours / 24)}d`;
}

function Stat({ label, value, accent }: { label: string; value: string; accent?: string }) {
  return (
    <div className="flex flex-col gap-1 px-4 py-3">
      <span className="font-mono text-[10px] uppercase tracking-wider text-on-surface-variant">
        {label}
      </span>
      <span className={`font-mono text-2xl font-bold leading-none ${accent ?? ""}`}>
        {value}
      </span>
    </div>
  );
}

// summary row above the queue table
export function QueueStats({
  listings,
  total,
}: {
  listings: ListingQueueItem[];
  total: number;
}) {
  return (
    <section
      className="panel grid grid-cols-2 divide-x divide-y divide-surface-variant sm:grid-cols-4 sm:divide-y-0"
      aria-label="Queue summary"
    >
      <Stat label="In queue" value={String(total)} />
      <Stat
        label="High (page)"
        value={String(countTier(listings, "HIGH"))}
        accent="text-error"
      />
      <Stat label="Low (page)" value={String(countTier(listings, "LOW"))} />
      <Stat label="Oldest wait" value={oldestWait(listings)} />
    </section>
  );
}
