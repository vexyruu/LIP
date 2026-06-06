"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { AssignButton } from "@/components/AssignButton";
import { ListingImage } from "@/components/ListingImage";
import { PageHeader } from "@/components/layout/PageHeader";
import { RejectModal } from "@/components/RejectModal";
import { SectionHeader } from "@/components/SectionHeader";
import { useListing, useSellerRisk } from "@/lib/hooks";
import {
  conditionLabel,
  formatPrice,
  listingShortId,
  policyViolationLabel,
  priceMarkerPercent,
  riskTierLabel,
  shortUUID,
} from "@/lib/format";
import type { ListingDetail } from "@/lib/types";

function IntelRow({
  label,
  value,
  missingHint,
}: {
  label: string;
  value: string;
  missingHint?: string;
}) {
  const isMissing = !value || value === "—";
  return (
    <div className="flex items-start justify-between gap-4 border-l-2 border-primary-container bg-surface-container-low p-3">
      <span className="font-mono text-xs uppercase tracking-wider text-on-surface-variant">
        {label}
      </span>
      <div className="text-right">
        <span className={`text-sm font-semibold ${isMissing ? "text-on-surface-variant" : ""}`}>
          {isMissing ? "Not detected" : value}
        </span>
        {isMissing && missingHint && (
          <p className="mt-1 max-w-[14rem] text-xs text-on-surface-variant">
            {missingHint}
          </p>
        )}
      </div>
    </div>
  );
}

function RiskFactorRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4 border-l-2 border-surface-variant bg-surface-container-lowest px-3 py-2">
      <span className="font-mono text-[10px] uppercase tracking-wider text-on-surface-variant">
        {label}
      </span>
      <span className="font-mono text-sm">{value}</span>
    </div>
  );
}

function hasMlExtraction(listing: ListingDetail): boolean {
  return Boolean(
    listing.extracted_brand ||
      listing.extracted_product ||
      listing.extracted_size
  );
}

export function ListingDetailClient({
  listingId,
  moderatorId,
  canModerate,
}: {
  listingId: string;
  moderatorId: string | null;
  canModerate: boolean;
}) {
  const router = useRouter();
  const {
    data: listing,
    error: loadError,
    isLoading,
  } = useListing(listingId);
  const { data: sellerRisk } = useSellerRisk(listing?.user_id);
  const [actionLoading, setActionLoading] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const [rejectOpen, setRejectOpen] = useState(false);

  const loading = isLoading && !listing;
  const error =
    actionError ?? (loadError ? (loadError as Error).message : null);

  const moderate = async (action: "APPROVE" | "REJECT", reason = "") => {
    setActionLoading(true);
    setActionError(null);
    try {
      const res = await fetch(`/api/listings/${listingId}/moderate`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action, reason }),
      });
      if (!res.ok) {
        throw new Error(await res.text());
      }
      router.push("/queue");
      router.refresh();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Moderation failed");
      setActionLoading(false);
      setRejectOpen(false);
    }
  };

  if (loading) {
    return (
      <div className="page-shell">
        <p className="font-mono text-sm text-on-surface-variant">
          Loading listing…
        </p>
      </div>
    );
  }

  if (error && !listing) {
    return (
      <div className="page-shell">
        <PageHeader
          title="Listing unavailable"
          breadcrumbs={[
            { label: "Review queue", href: "/queue" },
            { label: "Listing" },
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

  if (!listing) return null;

  const isUnderReview = listing.status === "UNDER_REVIEW";
  const showActions = isUnderReview && canModerate;
  const markerPct = priceMarkerPercent(
    listing.price_ask,
    listing.price_lower_bound,
    listing.price_upper_bound
  );

  return (
    <>
      <div className={`page-shell ${showActions ? "page-with-action-bar" : ""}`}>
        <PageHeader
          title={listing.title}
          description={`Status: ${listing.status} · Seller ${shortUUID(listing.user_id)}`}
          breadcrumbs={[
            { label: "Review queue", href: "/queue" },
            { label: listing.title },
          ]}
          actions={
            <Link
              href={`/users/${listing.user_id}`}
              className="btn-secondary"
            >
              Seller profile
            </Link>
          }
          meta={
            <span className="status-chip font-mono">{listing.status}</span>
          }
        />

        {error && (
          <div className="status-banner status-banner-error" role="alert">
            {error}
          </div>
        )}

        <div className="grid grid-cols-1 gap-6 xl:grid-cols-12 xl:gap-8">
          <div className="xl:col-span-7 flex flex-col gap-6">
            <ListingImage
              images={listing.images}
              title={listing.title}
              priority
              className="h-72 w-full sm:h-96"
              sizes="(max-width: 1280px) 100vw, 640px"
            />
            <p className="font-mono text-xs text-on-surface-variant">
              ID: {listingShortId(listing.listing_id)}
              {listing.images?.length ? ` · ${listing.images.length} photo${listing.images.length === 1 ? "" : "s"}` : ""}
            </p>

            <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
              <section className="panel p-4">
                <SectionHeader title="Description" />
                <p className="whitespace-pre-wrap text-sm leading-relaxed text-on-surface">
                  {listing.description || "—"}
                </p>
              </section>

              <section className="panel p-4">
                <SectionHeader title="Listing details" />
                <dl className="space-y-3 text-sm">
                  <div className="flex justify-between gap-4 border-b border-surface-variant pb-2">
                    <dt className="font-mono text-xs uppercase text-on-surface-variant">
                      Price
                    </dt>
                    <dd className="text-lg font-semibold text-primary">
                      {formatPrice(listing.price_ask)}
                    </dd>
                  </div>
                  <div className="flex justify-between gap-4 border-b border-surface-variant pb-2">
                    <dt className="font-mono text-xs uppercase text-on-surface-variant">
                      Condition
                    </dt>
                    <dd className="font-semibold">
                      {conditionLabel(listing.condition)}
                    </dd>
                  </div>
                  <div className="flex justify-between gap-4 border-b border-surface-variant pb-2">
                    <dt className="font-mono text-xs uppercase text-on-surface-variant">
                      Category
                    </dt>
                    <dd className="font-mono">
                      CAT-{String(listing.category_id).padStart(4, "0")}
                    </dd>
                  </div>
                  <div className="flex justify-between gap-4">
                    <dt className="font-mono text-xs uppercase text-on-surface-variant">
                      Seller
                    </dt>
                    <dd>
                      <Link
                        href={`/users/${listing.user_id}`}
                        className="font-mono text-sm text-alert-lime underline-offset-2 hover:underline"
                      >
                        {shortUUID(listing.user_id)}
                      </Link>
                    </dd>
                  </div>
                </dl>
              </section>
            </div>
          </div>

          <aside className="xl:col-span-5 flex flex-col gap-6">
            <section className="panel p-4">
              <SectionHeader title="Risk assessment" />
              <div className="flex items-start gap-6">
                <div>
                  <p className="font-mono text-4xl font-bold leading-none text-alert-coral">
                    {listing.risk_score != null
                      ? listing.risk_score.toFixed(2)
                      : "—"}
                  </p>
                  <p className="mt-1 font-mono text-xs text-on-surface-variant">
                    Seller risk score / 1.00
                  </p>
                </div>
                <div>
                  <p className="text-sm font-bold">{riskTierLabel(listing.risk_tier)}</p>
                  <p className="mt-3 font-mono text-xs text-on-surface-variant">
                    Policy violation:{" "}
                    <span className="text-on-surface">
                      {policyViolationLabel(listing.policy_violation)}
                    </span>
                  </p>
                </div>
              </div>
              <p className="mt-4 text-xs leading-relaxed text-on-surface-variant">
                Risk is computed for the{" "}
                <Link
                  href={`/users/${listing.user_id}`}
                  className="text-alert-lime underline-offset-2 hover:underline"
                >
                  seller
                </Link>
                , not this listing&apos;s price. Graph proximity, fraud-ring
                membership, and listing velocity drive the tier — a fair retail
                ask can still be HIGH if the seller is linked to banned accounts.
              </p>
              {sellerRisk && (
                <div className="mt-4 grid gap-2">
                  <p className="font-mono text-[10px] uppercase tracking-wider text-on-surface-variant">
                    Live score breakdown (fraud-service)
                  </p>
                  <RiskFactorRow
                    label="Component (WCC ring size)"
                    value={sellerRisk.component_score.toFixed(2)}
                  />
                  <RiskFactorRow
                    label="PageRank proximity"
                    value={sellerRisk.pagerank_proximity.toFixed(2)}
                  />
                  <RiskFactorRow
                    label="Velocity"
                    value={`${sellerRisk.velocity_score.toFixed(2)} (${sellerRisk.velocity_count}/hr)`}
                  />
                  {sellerRisk.velocity_flags.length > 0 && (
                    <p className="font-mono text-[10px] text-alert-coral">
                      {sellerRisk.velocity_flags.join(" · ")}
                    </p>
                  )}
                </div>
              )}
            </section>

            <section className="panel p-4">
              <SectionHeader title="ML extraction" />
              {!hasMlExtraction(listing) ? (
                <p className="text-sm leading-relaxed text-on-surface-variant">
                  No structured entities were stored for this listing. That
                  usually means the analysis worker has not finished, or the
                  listing was seeded before ML fields were populated.
                </p>
              ) : (
                <>
                  <p className="mb-3 text-xs leading-relaxed text-on-surface-variant">
                    Parsed from title and description by ml-service (brand
                    dictionary + size rules). Used for pricing features, not
                    moderation decisions alone.
                  </p>
                  <div className="grid gap-2">
                    <IntelRow
                      label="Brand"
                      value={listing.extracted_brand ?? ""}
                      missingHint="Brand phrase not matched in catalog."
                    />
                    <IntelRow
                      label="Model"
                      value={listing.extracted_product ?? ""}
                      missingHint="Derived from title after brand/size removal."
                    />
                    <IntelRow
                      label="Size"
                      value={listing.extracted_size ?? ""}
                      missingHint='Look for patterns like "Size 10".'
                    />
                  </div>
                </>
              )}
            </section>

            <section className="panel p-4">
              <SectionHeader title="Pricing intelligence" />
              <p className="mb-2 text-xs leading-relaxed text-on-surface-variant">
                XGBoost model (ONNX) trained on marketplace listings — point estimate
                plus ± median absolute error band.
              </p>
              <p className="font-mono text-xs uppercase text-on-surface-variant">
                Suggested market value
              </p>
              <p className="mt-1 font-mono text-2xl font-bold">
                {listing.suggested_price != null
                  ? formatPrice(listing.suggested_price)
                  : "—"}
              </p>
              {listing.price_lower_bound != null &&
                listing.price_upper_bound != null && (
                  <div className="mt-4 space-y-2">
                    <div className="relative h-2 bg-surface-container-high">
                      <div
                        className="absolute inset-y-0 left-[10%] right-[10%] bg-primary-container/20"
                        aria-hidden
                      />
                      <div
                        className="absolute top-[-3px] h-3 w-0.5 bg-alert-lime"
                        style={{ left: `${markerPct}%` }}
                        aria-hidden
                      />
                    </div>
                    <div className="flex justify-between font-mono text-xs text-on-surface-variant">
                      <span>{formatPrice(listing.price_lower_bound)} min</span>
                      <span>{formatPrice(listing.price_upper_bound)} max</span>
                    </div>
                  </div>
                )}
            </section>

            {isUnderReview && canModerate && (
              <section className="panel p-4">
                <SectionHeader title="Review assignment" />
                <p className="mb-3 text-xs leading-relaxed text-on-surface-variant">
                  Claim this listing so other moderators know it&apos;s being
                  handled. Claims use an optimistic lock — if two people claim at
                  once, only the first succeeds.
                </p>
                <AssignButton
                  listingId={listing.listing_id}
                  assignedTo={listing.assigned_to}
                  moderatorId={moderatorId}
                  reviewable={isUnderReview}
                />
              </section>
            )}

            <div className="status-banner status-banner-info">
              <span className="material-symbols-outlined text-lg" aria-hidden>
                info
              </span>
              <p className="text-sm leading-relaxed">
                Verify ML confidence and seller connections before approving.
                Open the seller profile to inspect the fraud graph.
              </p>
            </div>
          </aside>
        </div>
      </div>

      {showActions && (
        <div className="action-bar" role="region" aria-label="Moderation actions">
          <div className="action-bar-inner flex-col sm:flex-row">
            <button
              type="button"
              disabled={actionLoading}
              onClick={() => moderate("APPROVE")}
              className="btn-primary flex-1 text-sm"
            >
              <span className="material-symbols-outlined text-lg" aria-hidden>
                check_circle
              </span>
              Approve listing
            </button>
            <button
              type="button"
              disabled={actionLoading}
              onClick={() => setRejectOpen(true)}
              className="btn-destructive flex-1 text-sm"
            >
              <span className="material-symbols-outlined text-lg" aria-hidden>
                cancel
              </span>
              Reject listing
            </button>
          </div>
        </div>
      )}

      <RejectModal
        open={rejectOpen}
        loading={actionLoading}
        onClose={() => !actionLoading && setRejectOpen(false)}
        onConfirm={(reason) => moderate("REJECT", reason)}
      />
    </>
  );
}
