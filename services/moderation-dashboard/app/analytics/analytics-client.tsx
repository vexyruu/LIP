"use client";

import Link from "next/link";
import { useMemo } from "react";
import { PageHeader } from "@/components/layout/PageHeader";
import { SectionHeader } from "@/components/SectionHeader";
import { useAnalytics } from "@/lib/hooks";

function formatNumber(n: number): string {
  return n.toLocaleString("en-US");
}

function formatLatency(seconds: number | null): string {
  if (seconds == null || Number.isNaN(seconds)) return "—";
  if (seconds < 1) return `${Math.round(seconds * 1000)}ms`;
  return `${seconds.toFixed(2)}s`;
}

function tierPercent(tiers: Record<string, number>, key: string): number {
  const scored = (tiers.HIGH ?? 0) + (tiers.MEDIUM ?? 0) + (tiers.LOW ?? 0);
  if (scored === 0) return 0;
  return ((tiers[key] ?? 0) / scored) * 100;
}

function formatDecisionTime(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function StatCard({
  label,
  value,
  hint,
  accent,
}: {
  label: string;
  value: string;
  hint: string;
  accent?: "lime" | "coral";
}) {
  return (
    <article className="panel p-4">
      <p className="data-label mb-2">{label}</p>
      <p
        className={`text-3xl font-bold tracking-tight ${
          accent === "lime"
            ? "text-alert-lime"
            : accent === "coral"
              ? "text-alert-coral"
              : ""
        }`}
      >
        {value}
      </p>
      <p className="mt-2 font-mono text-xs text-on-surface-variant">{hint}</p>
    </article>
  );
}

export function AnalyticsClient() {
  const { data, error: swrError, isLoading } = useAnalytics();
  const loading = isLoading && !data;
  const error = swrError ? (swrError as Error).message : null;

  const chartBars = useMemo(() => {
    if (!data?.daily_submissions?.length) return [];
    const max = Math.max(...data.daily_submissions.map((d) => d.total), 1);
    return data.daily_submissions.map((d) => ({
      ...d,
      heightPct: (d.total / max) * 100,
      highPct: d.total > 0 ? (d.high_risk / d.total) * 100 : 0,
    }));
  }, [data]);

  const highPct = data ? tierPercent(data.by_risk_tier, "HIGH") : 0;
  const medPct = data ? tierPercent(data.by_risk_tier, "MEDIUM") : 0;
  const lowPct = data ? tierPercent(data.by_risk_tier, "LOW") : 0;

  const recentDecisions = data?.recent_decisions ?? [];
  const showQueueAlert = (data?.under_review ?? 0) > 0;

  if (loading) {
    return (
      <div className="page-shell">
        <p className="font-mono text-sm text-on-surface-variant">
          Loading analytics…
        </p>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="page-shell">
        <PageHeader title="Analytics" />
        <div className="status-banner status-banner-error" role="alert">
          {error ?? "No data"}
        </div>
      </div>
    );
  }

  return (
    <div className="page-shell">
      <PageHeader
        title="Analytics"
        description="Dev metrics from Postgres — pipeline health and moderation activity."
        actions={
          <Link href="/queue" className="btn-primary">
            Open queue ({data.under_review})
          </Link>
        }
      />

      {showQueueAlert && (
        <div className="status-banner status-banner-info" role="status">
          <span className="material-symbols-outlined text-lg" aria-hidden>
            pending_actions
          </span>
          <div>
            <p className="text-sm">
              {data.under_review} listing
              {data.under_review === 1 ? "" : "s"} awaiting review.
            </p>
            <Link
              href="/queue"
              className="mt-2 inline-block font-mono text-xs text-alert-lime underline-offset-2 hover:underline"
            >
              Go to review queue
            </Link>
          </div>
        </div>
      )}

      <section
        className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4"
        aria-label="Summary metrics"
      >
        <StatCard
          label="Total listings"
          value={formatNumber(data.total_listings)}
          hint={`+${data.last_24h_submissions} in last 24h`}
        />
        <StatCard
          label="Avg pipeline time"
          value={formatLatency(data.avg_pipeline_seconds)}
          hint="Submit to final status"
        />
        <StatCard
          label="Under review"
          value={formatNumber(data.under_review)}
          hint="Moderation queue depth"
          accent="lime"
        />
        <StatCard
          label="Policy rejections"
          value={formatNumber(data.policy_rejections)}
          hint="policy_violation = true"
          accent="coral"
        />
      </section>

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-12">
        <div className="xl:col-span-8 flex flex-col gap-6">
          <section className="panel p-4">
            <SectionHeader title="Listing submissions (7 days)" />
            {chartBars.length === 0 ? (
              <p className="text-sm text-on-surface-variant">
                No submissions in the last 7 days.
              </p>
            ) : (
              <>
                <div
                  className="flex h-40 items-end gap-2 border-b border-surface-variant pb-2"
                  role="img"
                  aria-label="Bar chart of daily submissions"
                >
                  {chartBars.map((bar) => (
                    <div
                      key={bar.date}
                      className="flex min-w-0 flex-1 flex-col justify-end h-full"
                      title={`${bar.date}: ${bar.total} total, ${bar.high_risk} high risk`}
                    >
                      <div
                        className="relative w-full border-t border-alert-lime/30 bg-alert-lime/10"
                        style={{ height: `${Math.max(bar.heightPct, 4)}%` }}
                      >
                        {bar.high_risk > 0 && (
                          <div
                            className="absolute inset-x-0 bottom-0 border-t-2 border-alert-coral bg-alert-coral/40"
                            style={{ height: `${bar.highPct}%` }}
                          />
                        )}
                      </div>
                    </div>
                  ))}
                </div>
                <div className="mt-3 flex justify-between font-mono text-[10px] uppercase tracking-wider text-on-surface-variant">
                  {chartBars.map((bar) => (
                    <span key={bar.date}>{bar.date.slice(5)}</span>
                  ))}
                </div>
              </>
            )}
          </section>

          <section className="panel overflow-hidden">
            <div className="panel-header">Recent moderation decisions</div>
            {recentDecisions.length === 0 ? (
              <p className="p-4 text-sm text-on-surface-variant">
                No moderation decisions yet.
              </p>
            ) : (
              <div className="overflow-x-auto">
                <table className="data-table min-w-[640px]">
                  <thead>
                    <tr>
                      <th scope="col">Time</th>
                      <th scope="col">Listing</th>
                      <th scope="col">Action</th>
                      <th scope="col">Risk</th>
                      <th scope="col" className="text-right">
                        Link
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {recentDecisions.map((row) => (
                      <tr key={`${row.listing_id}-${row.timestamp}`}>
                        <td className="text-on-surface-variant whitespace-nowrap">
                          {formatDecisionTime(row.timestamp)}
                        </td>
                        <td className="max-w-[200px] truncate">{row.title}</td>
                        <td>
                          <span
                            className={
                              row.action === "APPROVE"
                                ? "font-bold text-alert-lime"
                                : "font-bold text-alert-coral"
                            }
                          >
                            {row.action}
                          </span>
                        </td>
                        <td className="text-on-surface-variant">
                          {row.risk_tier ?? "—"}
                        </td>
                        <td className="text-right">
                          <Link
                            href={`/listings/${row.listing_id}`}
                            className="font-mono text-xs text-alert-lime underline-offset-2 hover:underline"
                          >
                            View
                          </Link>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          <section className="panel p-4">
            <SectionHeader title="Status breakdown" />
            <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
              {Object.entries(data.by_status).map(([status, count]) => (
                <div
                  key={status}
                  className="border border-surface-variant bg-surface-container-low p-3"
                >
                  <p className="data-label mb-1">{status}</p>
                  <p className="text-xl font-bold">{count}</p>
                </div>
              ))}
            </div>
          </section>
        </div>

        <aside className="xl:col-span-4">
          <section className="panel p-4">
            <SectionHeader title="Risk tier breakdown" />
            <div className="space-y-6">
              {[
                {
                  key: "HIGH",
                  label: "High risk",
                  color: "bg-alert-coral",
                  text: "text-alert-coral",
                  pct: highPct,
                },
                {
                  key: "MEDIUM",
                  label: "Medium risk",
                  color: "bg-orange-400",
                  text: "text-orange-400",
                  pct: medPct,
                },
                {
                  key: "LOW",
                  label: "Low risk",
                  color: "bg-alert-lime",
                  text: "text-alert-lime",
                  pct: lowPct,
                },
              ].map(({ key, label, color, text, pct }) => (
                <div key={key}>
                  <div
                    className={`mb-2 flex justify-between text-xs font-bold uppercase tracking-wider ${text}`}
                  >
                    <span>{label}</span>
                    <span className="font-mono text-on-surface-variant">
                      {pct.toFixed(1)}%
                    </span>
                  </div>
                  <div className="relative h-1 bg-outline">
                    <div
                      className={`absolute inset-y-0 left-0 ${color}`}
                      style={{ width: `${pct}%` }}
                      role="progressbar"
                      aria-valuenow={Math.round(pct)}
                      aria-valuemin={0}
                      aria-valuemax={100}
                      aria-label={`${label} share`}
                    />
                  </div>
                </div>
              ))}
            </div>
            <p className="mt-6 border-t border-outline/30 pt-4 text-xs leading-relaxed text-on-surface-variant">
              Local Postgres aggregates (excludes DRAFT). Pricing inference uses
              XGBoost via ml-service. A production deployment would stream events
              to a warehouse for long-range trend analysis.
            </p>
          </section>
        </aside>
      </div>
    </div>
  );
}
