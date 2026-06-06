"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { BanModal } from "@/components/BanModal";
import { ListingImage } from "@/components/ListingImage";
import { PageHeader } from "@/components/layout/PageHeader";
import { SectionHeader } from "@/components/SectionHeader";
import { useSellerRisk, useUser, useUserListings } from "@/lib/hooks";
import {
  formatPrice,
  formatRelativeTime,
  riskTierLabel,
  shortUUID,
  tierScoreClass,
} from "@/lib/format";
import type { BanUserResponse } from "@/lib/types";

function statusBadgeClass(status: string): string {
  if (status === "BANNED") {
    return "bg-alert-coral/20 text-alert-coral border-alert-coral";
  }
  return "bg-alert-lime/20 text-alert-lime border-alert-lime";
}

function listingStatusClass(status: string): string {
  switch (status) {
    case "UNDER_REVIEW":
      return "text-alert-lime";
    case "REJECTED":
      return "text-alert-coral";
    case "LIVE":
      return "text-on-surface";
    default:
      return "text-on-surface-variant";
  }
}

export function UserClient({
  userId,
  canModerate,
}: {
  userId: string;
  canModerate: boolean;
}) {
  const [page, setPage] = useState(1);
  const limit = 10;
  const [banOpen, setBanOpen] = useState(false);
  const [banLoading, setBanLoading] = useState(false);
  const [banNotice, setBanNotice] = useState<string | null>(null);
  const [banError, setBanError] = useState<string | null>(null);

  const { data: user, error: userErr, isLoading: userLoading, mutate: mutateUser } =
    useUser(userId);
  const {
    data: listings,
    error: listingsErr,
    isLoading: listingsIsLoading,
    mutate: mutateListings,
  } = useUserListings(userId, page, limit);
  const {
    data: sellerRisk,
    isLoading: riskIsLoading,
    mutate: mutateRisk,
  } = useSellerRisk(userId);

  const loading = userLoading && !user;
  const error = banError ?? (userErr ? (userErr as Error).message : null);
  const listingsLoading = listingsIsLoading && !listings;
  const listingsError = listingsErr ? (listingsErr as Error).message : null;
  const riskLoading = riskIsLoading && !sellerRisk;

  const totalPages = useMemo(
    () => Math.max(1, Math.ceil((listings?.total ?? 0) / limit)),
    [listings?.total, limit]
  );

  async function banUser(reason: string) {
    setBanLoading(true);
    setBanNotice(null);
    setBanError(null);
    try {
      const res = await fetch(`/api/users/${userId}/ban`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ reason }),
      });
      if (!res.ok) throw new Error(await res.text());
      const data = (await res.json()) as BanUserResponse;
      setBanOpen(false);
      setBanNotice(
        data.listings_flagged > 0
          ? `Seller banned. ${data.listings_flagged} LIVE listing${data.listings_flagged === 1 ? "" : "s"} moved to review.`
          : "Seller banned."
      );
      await Promise.all([mutateUser(), mutateListings(), mutateRisk()]);
    } catch (err) {
      setBanError(err instanceof Error ? err.message : "Failed to ban user");
    } finally {
      setBanLoading(false);
    }
  }

  if (loading && !user) {
    return (
      <div className="page-shell">
        <p className="font-mono text-sm text-on-surface-variant">
          Loading seller profile…
        </p>
      </div>
    );
  }

  if (error && !user) {
    return (
      <div className="page-shell">
        <PageHeader
          title="Seller unavailable"
          breadcrumbs={[
            { label: "Review queue", href: "/queue" },
            { label: "Seller profile" },
          ]}
        />
        <div className="status-banner status-banner-error" role="alert">
          {error}
        </div>
        <Link href="/queue" className="btn-secondary w-fit">
          Back to queue
        </Link>
      </div>
    );
  }

  if (!user) return null;

  const isBanned = user.status === "BANNED";
  const showBanAction = !isBanned && canModerate;

  return (
    <>
      <div className={`page-shell ${showBanAction ? "page-with-action-bar" : ""}`}>
        <PageHeader
          title={user.display_name}
          description={`Seller account · ${shortUUID(user.user_id)}`}
          breadcrumbs={[
            { label: "Review queue", href: "/queue" },
            { label: user.display_name },
          ]}
          actions={
            <Link href={`/users/${userId}/graph`} className="btn-secondary">
              Connection graph
            </Link>
          }
          meta={
            <span
              className={`inline-flex border-2 px-3 py-1 font-mono text-xs font-bold uppercase tracking-widest ${statusBadgeClass(user.status)}`}
            >
              {user.status}
            </span>
          }
        />

        {error && (
          <div className="status-banner status-banner-error" role="alert">
            {error}
          </div>
        )}

        {banNotice && (
          <div className="status-banner status-banner-info" role="status">
            {banNotice}
          </div>
        )}

        {isBanned && user.ban_audit && (
          <section className="panel p-4" aria-label="Ban record">
            <SectionHeader title="Ban record" />
            <dl className="space-y-3 font-mono text-sm">
              <div className="flex justify-between gap-4 border-b border-surface-variant pb-3">
                <dt className="text-xs uppercase text-on-surface-variant">
                  Reason
                </dt>
                <dd className="max-w-[65%] text-right">{user.ban_audit.reason}</dd>
              </div>
              <div className="flex justify-between gap-4 border-b border-surface-variant pb-3">
                <dt className="text-xs uppercase text-on-surface-variant">
                  Moderator
                </dt>
                <dd>{shortUUID(user.ban_audit.moderator_id)}</dd>
              </div>
              <div className="flex justify-between gap-4 border-b border-surface-variant pb-3">
                <dt className="text-xs uppercase text-on-surface-variant">
                  Listings flagged
                </dt>
                <dd>{user.ban_audit.listings_flagged}</dd>
              </div>
              <div className="flex justify-between gap-4">
                <dt className="text-xs uppercase text-on-surface-variant">
                  Banned
                </dt>
                <dd>{formatRelativeTime(user.ban_audit.created_at)}</dd>
              </div>
            </dl>
          </section>
        )}

        {isBanned && !user.ban_audit && (
          <div className="status-banner status-banner-info" role="status">
            Seller is banned. Ban audit details are unavailable (apply migration
            000008 if this is a fresh dev DB).
          </div>
        )}

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <section className="panel p-4">
            <SectionHeader title="Account details" />
            <dl className="space-y-3 font-mono text-sm">
              <div className="flex justify-between gap-4 border-b border-surface-variant pb-3">
                <dt className="text-xs uppercase text-on-surface-variant">
                  Email
                </dt>
                <dd>{user.email}</dd>
              </div>
              <div className="flex justify-between gap-4 border-b border-surface-variant pb-3">
                <dt className="text-xs uppercase text-on-surface-variant">
                  Display name
                </dt>
                <dd>{user.display_name}</dd>
              </div>
              <div className="flex justify-between gap-4 border-b border-surface-variant pb-3">
                <dt className="text-xs uppercase text-on-surface-variant">
                  User ID
                </dt>
                <dd className="max-w-[60%] break-all text-right">
                  {user.user_id}
                </dd>
              </div>
              <div className="flex justify-between gap-4">
                <dt className="text-xs uppercase text-on-surface-variant">
                  Member since
                </dt>
                <dd>{formatRelativeTime(user.created_at)}</dd>
              </div>
            </dl>
          </section>

          <section className="panel p-4">
            <SectionHeader title="Fraud intelligence" />
            {riskLoading ? (
              <p className="font-mono text-sm text-on-surface-variant">
                Loading live risk score…
              </p>
            ) : sellerRisk ? (
              <div className="mb-4 space-y-3">
                <div className="flex items-baseline gap-3">
                  <p
                    className={`font-mono text-3xl font-bold ${tierScoreClass(sellerRisk.risk_tier)}`}
                  >
                    {sellerRisk.risk_score.toFixed(2)}
                  </p>
                  <p className="text-sm font-bold">
                    {riskTierLabel(sellerRisk.risk_tier)}
                  </p>
                </div>
                <dl className="grid gap-2 font-mono text-xs">
                  <div className="flex justify-between gap-4">
                    <dt className="text-on-surface-variant">Component</dt>
                    <dd>{sellerRisk.component_score.toFixed(2)}</dd>
                  </div>
                  <div className="flex justify-between gap-4">
                    <dt className="text-on-surface-variant">PageRank</dt>
                    <dd>{sellerRisk.pagerank_proximity.toFixed(2)}</dd>
                  </div>
                  <div className="flex justify-between gap-4">
                    <dt className="text-on-surface-variant">Velocity</dt>
                    <dd>
                      {sellerRisk.velocity_score.toFixed(2)} (
                      {sellerRisk.velocity_count}/hr)
                    </dd>
                  </div>
                </dl>
              </div>
            ) : (
              <p className="mb-4 text-sm text-on-surface-variant">
                Live risk score unavailable (fraud-service may be down).
              </p>
            )}
            <p className="mb-4 text-sm leading-relaxed text-on-surface-variant">
              Inspect device, IP, and payment connections in Neo4j. Shared links
              with banned accounts increase risk scores.
            </p>
            <Link href={`/users/${userId}/graph`} className="btn-primary">
              View connection graph
            </Link>
          </section>
        </div>

        <section className="panel overflow-hidden">
          <div className="panel-header flex items-center justify-between">
            <span>Seller listings</span>
            {!listingsLoading && listings && (
              <span className="font-normal normal-case tracking-normal text-on-surface-variant">
                {listings.total} total · page {page} of {totalPages}
              </span>
            )}
          </div>

          {listingsError && (
            <div className="status-banner status-banner-error m-4" role="alert">
              {listingsError}
              <button
                type="button"
                onClick={() => mutateListings()}
                className="ml-3 underline"
              >
                Retry
              </button>
            </div>
          )}

          {listingsLoading ? (
            <p className="p-6 font-mono text-sm text-on-surface-variant">
              Loading listings…
            </p>
          ) : !listings || listings.listings.length === 0 ? (
            <p className="p-6 text-sm text-on-surface-variant">
              No listings for this seller yet.
            </p>
          ) : (
            <>
              <div className="overflow-x-auto">
                <table className="data-table min-w-[760px]">
                  <thead>
                    <tr>
                      <th scope="col">Listing</th>
                      <th scope="col">Status</th>
                      <th scope="col">Price</th>
                      <th scope="col">Risk</th>
                      <th scope="col">Submitted</th>
                      <th scope="col" className="text-right">
                        Action
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {listings.listings.map((listing) => (
                      <tr key={listing.listing_id}>
                        <td>
                          <div className="flex min-w-[12rem] items-center gap-3">
                            <ListingImage
                              images={listing.images}
                              title={listing.title}
                              className="h-10 w-10 shrink-0"
                              sizes="40px"
                              compact
                            />
                            <span className="font-semibold leading-snug">
                              {listing.title}
                            </span>
                          </div>
                        </td>
                        <td>
                          <span
                            className={`font-mono text-xs font-bold uppercase ${listingStatusClass(listing.status)}`}
                          >
                            {listing.status}
                          </span>
                        </td>
                        <td className="font-mono text-sm">
                          {formatPrice(listing.price_ask)}
                        </td>
                        <td className="font-mono text-xs text-on-surface-variant">
                          {listing.risk_tier ?? "—"}
                          {listing.risk_score != null &&
                            ` · ${listing.risk_score.toFixed(2)}`}
                        </td>
                        <td className="whitespace-nowrap font-mono text-xs text-on-surface-variant">
                          {formatRelativeTime(listing.created_at)}
                        </td>
                        <td className="text-right">
                          <Link
                            href={`/listings/${listing.listing_id}`}
                            className="font-mono text-xs text-alert-lime underline-offset-2 hover:underline"
                          >
                            Open
                          </Link>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <footer className="flex items-center justify-between gap-4 border-t border-surface-variant p-4">
                <button
                  type="button"
                  disabled={page <= 1 || listingsLoading}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  className="btn-secondary min-h-[2.25rem] px-4"
                >
                  Previous
                </button>
                <span className="font-mono text-sm text-on-surface-variant">
                  Page {page} of {totalPages}
                </span>
                <button
                  type="button"
                  disabled={page >= totalPages || listingsLoading}
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  className="btn-secondary min-h-[2.25rem] px-4"
                >
                  Next
                </button>
              </footer>
            </>
          )}
        </section>
      </div>

      {showBanAction && (
        <div className="action-bar" role="region" aria-label="Seller actions">
          <div className="action-bar-inner">
            <button
              type="button"
              disabled={banLoading}
              onClick={() => setBanOpen(true)}
              className="btn-destructive w-full sm:w-auto sm:min-w-[16rem]"
            >
              <span className="material-symbols-outlined text-lg" aria-hidden>
                block
              </span>
              Ban seller
            </button>
          </div>
        </div>
      )}

      <BanModal
        open={banOpen}
        loading={banLoading}
        onClose={() => !banLoading && setBanOpen(false)}
        onConfirm={banUser}
      />
    </>
  );
}
