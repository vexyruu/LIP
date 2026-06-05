"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { ListingImage } from "@/components/ListingImage";
import { PageHeader } from "@/components/layout/PageHeader";
import {
  formatPrice,
  formatRelativeTime,
  shortUUID,
  tierBadgeClass,
  tierScoreClass,
} from "@/lib/format";
import type { ListListingsResponse, SortOption, TierFilter } from "@/lib/types";

const TIER_FILTERS: TierFilter[] = ["ALL", "HIGH", "MEDIUM", "LOW"];
const QUEUE_POLL_MS = 30_000;

function ListingThumb({
  title,
  images,
}: {
  title: string;
  images?: string[];
}) {
  return (
    <ListingImage
      images={images}
      title={title}
      className="h-10 w-10 shrink-0"
      sizes="40px"
      compact
    />
  );
}

export function QueueClient() {
  const [tier, setTier] = useState<TierFilter>("ALL");
  const [sort, setSort] = useState<SortOption>("risk_score");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(25);
  const [data, setData] = useState<ListListingsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadQueue = useCallback(async (silent = false) => {
    if (!silent) {
      setLoading(true);
    }
    setError(null);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: String(limit),
        sort,
      });
      if (tier !== "ALL") {
        params.set("tier", tier);
      }
      const res = await fetch(`/api/queue?${params.toString()}`);
      if (!res.ok) {
        throw new Error(await res.text());
      }
      const json = (await res.json()) as ListListingsResponse;
      setData(json);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load queue");
      if (!silent) {
        setData(null);
      }
    } finally {
      if (!silent) {
        setLoading(false);
      }
    }
  }, [limit, page, sort, tier]);

  useEffect(() => {
    loadQueue();
  }, [loadQueue]);

  useEffect(() => {
    const timer = window.setInterval(() => {
      void loadQueue(true);
    }, QUEUE_POLL_MS);
    return () => window.clearInterval(timer);
  }, [loadQueue]);

  const totalPages = useMemo(() => {
    if (!data) return 1;
    return Math.max(1, Math.ceil(data.total / limit));
  }, [data, limit]);

  const pageNumbers = useMemo(() => {
    const pages: number[] = [];
    const start = Math.max(1, page - 1);
    const end = Math.min(totalPages, start + 2);
    for (let i = start; i <= end; i += 1) {
      pages.push(i);
    }
    return pages;
  }, [page, totalPages]);

  return (
    <div className="page-shell">
      <PageHeader
        title="Review queue"
        description="Listings flagged for human moderation. Risk tier reflects the seller's fraud graph score, not listing price."
        meta={
          <p className="font-mono text-xs text-on-surface-variant">
            {loading
              ? "Refreshing…"
              : `${data?.total ?? 0} listing${data?.total === 1 ? "" : "s"} in queue`}
            {!loading && " · auto-refreshes every 30s"}
          </p>
        }
      />

      <section className="panel" aria-label="Queue filters and sorting">
        <div className="flex flex-col gap-4 p-4 lg:flex-row lg:items-center lg:justify-between">
          <div
            className="flex flex-wrap gap-2"
            role="group"
            aria-label="Filter by risk tier"
          >
            {TIER_FILTERS.map((value) => {
              const active = tier === value;
              return (
                <button
                  key={value}
                  type="button"
                  aria-pressed={active}
                  onClick={() => {
                    setTier(value);
                    setPage(1);
                  }}
                  className={`filter-chip ${active ? "filter-chip-active" : ""}`}
                >
                  {value === "ALL" ? "All tiers" : value}
                </button>
              );
            })}
          </div>

          <div className="flex flex-wrap items-center gap-4">
            <label className="flex items-center gap-2 font-mono text-xs uppercase text-on-surface-variant">
              Sort
              <select
                className="min-h-[2.5rem] cursor-pointer border border-outline bg-background px-3 py-2 text-sm text-on-surface"
                value={sort}
                onChange={(e) => {
                  setSort(e.target.value as SortOption);
                  setPage(1);
                }}
                aria-label="Sort queue"
              >
                <option value="risk_score">Highest risk first</option>
                <option value="created_at">Oldest first</option>
              </select>
            </label>
          </div>
        </div>
      </section>

      {error && (
        <div className="status-banner status-banner-error" role="alert">
          <span className="material-symbols-outlined text-lg" aria-hidden>
            error
          </span>
          <div className="flex-1">
            <p>{error}</p>
            <button
              type="button"
              onClick={() => loadQueue()}
              className="mt-2 font-mono text-xs underline underline-offset-2"
            >
              Try again
            </button>
          </div>
        </div>
      )}

      <section
        className="panel overflow-hidden"
        aria-live="polite"
        aria-busy={loading}
      >
        <div className="panel-header flex items-center justify-between">
          <span>Pending listings</span>
          {!loading && data && (
            <span className="font-normal normal-case tracking-normal text-on-surface-variant">
              Page {page} of {totalPages}
            </span>
          )}
        </div>

        {loading ? (
          <p className="p-8 font-mono text-sm text-on-surface-variant">
            Loading queue…
          </p>
        ) : !data || data.listings.length === 0 ? (
          <div className="p-10 text-center">
            <p className="text-lg font-bold uppercase tracking-tight">
              No listings under review
            </p>
            <p className="mt-2 text-sm text-on-surface-variant">
              New flagged listings appear here automatically.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="data-table min-w-[820px]">
              <thead>
                <tr>
                  <th scope="col">Listing</th>
                  <th scope="col">Seller</th>
                  <th scope="col">Price</th>
                  <th scope="col">Tier</th>
                  <th scope="col">Score</th>
                  <th scope="col">Submitted</th>
                  <th scope="col" className="text-right">
                    Action
                  </th>
                </tr>
              </thead>
              <tbody>
                {data.listings.map((listing) => (
                  <tr key={listing.listing_id}>
                    <td>
                      <div className="flex min-w-[12rem] items-center gap-3">
                        <ListingThumb
                          title={listing.title}
                          images={listing.images}
                        />
                        <span className="font-semibold leading-snug">
                          {listing.title}
                        </span>
                      </div>
                    </td>
                    <td className="font-mono text-xs text-on-surface-variant">
                      {shortUUID(listing.user_id)}
                    </td>
                    <td className="font-mono text-sm">
                      {formatPrice(listing.price_ask)}
                    </td>
                    <td>
                      <span
                        className={`inline-block border px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider ${tierBadgeClass(listing.risk_tier)}`}
                      >
                        {listing.risk_tier ?? "—"}
                      </span>
                    </td>
                    <td
                      className={`font-mono text-xs ${tierScoreClass(listing.risk_tier)}`}
                    >
                      {listing.risk_score != null
                        ? listing.risk_score.toFixed(2)
                        : "—"}
                    </td>
                    <td className="whitespace-nowrap font-mono text-xs text-on-surface-variant">
                      {formatRelativeTime(listing.created_at)}
                    </td>
                    <td className="text-right">
                      <Link
                        href={`/listings/${listing.listing_id}`}
                        className="btn-primary min-h-[2.25rem] px-4 text-[11px]"
                      >
                        Review
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {!loading && data && data.listings.length > 0 && (
          <footer className="flex flex-col gap-4 border-t border-surface-variant p-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-3">
              <button
                type="button"
                disabled={page <= 1 || loading}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className="btn-secondary min-h-[2.25rem] px-3"
                aria-label="Previous page"
              >
                <span className="material-symbols-outlined text-lg" aria-hidden>
                  chevron_left
                </span>
              </button>
              <div className="flex items-center gap-2">
                {pageNumbers.map((n) => (
                  <button
                    key={n}
                    type="button"
                    onClick={() => setPage(n)}
                    aria-current={n === page ? "page" : undefined}
                    className={
                      n === page
                        ? "min-w-[2rem] border-b-2 border-primary-container pb-1 font-mono text-sm text-primary-container"
                        : "min-w-[2rem] pb-1 font-mono text-sm text-on-surface-variant hover:text-on-surface"
                    }
                  >
                    {n}
                  </button>
                ))}
              </div>
              <button
                type="button"
                disabled={page >= totalPages || loading}
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                className="btn-secondary min-h-[2.25rem] px-3"
                aria-label="Next page"
              >
                <span className="material-symbols-outlined text-lg" aria-hidden>
                  chevron_right
                </span>
              </button>
            </div>

            <label className="flex items-center gap-2 font-mono text-xs uppercase text-on-surface-variant">
              Rows
              <select
                className="min-h-[2.25rem] cursor-pointer border border-outline bg-background px-2 py-1 text-sm"
                value={limit}
                onChange={(e) => {
                  setLimit(Number(e.target.value));
                  setPage(1);
                }}
                aria-label="Rows per page"
              >
                <option value={25}>25</option>
                <option value={50}>50</option>
                <option value={100}>100</option>
              </select>
            </label>
          </footer>
        )}
      </section>
    </div>
  );
}
