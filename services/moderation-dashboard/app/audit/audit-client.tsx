"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { PageHeader } from "@/components/layout/PageHeader";
import { useModerationDecisions } from "@/lib/hooks";
import { shortUUID } from "@/lib/format";

type ActionFilter = "ALL" | "APPROVE" | "REJECT" | "BAN";

function actionLabel(value: ActionFilter): string {
  switch (value) {
    case "ALL":
      return "All actions";
    case "APPROVE":
      return "Approve";
    case "REJECT":
      return "Reject";
    case "BAN":
      return "Ban";
  }
}

function actionClass(action: string): string {
  if (action === "APPROVE") return "font-bold text-alert-lime";
  if (action === "BAN") return "font-bold text-on-surface";
  return "font-bold text-alert-coral";
}

function formatDecisionTime(timestamp: string): string {
  const d = new Date(timestamp);
  return d.toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function AuditClient() {
  const [action, setAction] = useState<ActionFilter>("ALL");
  const [page, setPage] = useState(1);
  const limit = 25;

  const { data, error: swrError, isLoading, mutate } = useModerationDecisions({
    action,
    page,
    limit,
  });
  const loading = isLoading && !data;
  const error = swrError ? (swrError as Error).message : null;

  const totalPages = useMemo(
    () => Math.max(1, Math.ceil((data?.total ?? 0) / limit)),
    [data?.total, limit]
  );

  if (loading && !data) {
    return (
      <div className="page-shell">
        <p className="font-mono text-sm text-on-surface-variant">
          Loading history…
        </p>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="page-shell">
        <PageHeader title="Moderation history" />
        <div className="status-banner status-banner-error" role="alert">
          {error ?? "No data"}
        </div>
        <button type="button" onClick={() => mutate()} className="btn-secondary w-fit">
          Try again
        </button>
      </div>
    );
  }

  return (
    <div className="page-shell">
      <PageHeader
        title="Moderation history"
        description="Past approve, reject, and ban decisions across all listings."
        meta={
          <p className="font-mono text-xs text-on-surface-variant">
            {data.total} decision{data.total === 1 ? "" : "s"} on record
          </p>
        }
      />

      <section className="panel p-4" aria-label="Filter decisions">
        <div
          className="flex flex-wrap gap-2"
          role="group"
          aria-label="Filter by action"
        >
          {(["ALL", "APPROVE", "REJECT", "BAN"] as ActionFilter[]).map((value) => (
            <button
              key={value}
              type="button"
              aria-pressed={action === value}
              onClick={() => {
                setAction(value);
                setPage(1);
              }}
              className={`filter-chip ${action === value ? "filter-chip-active" : ""}`}
            >
              {actionLabel(value)}
            </button>
          ))}
        </div>
      </section>

      <section className="panel overflow-hidden">
        <div className="panel-header">Decisions</div>
        {data.decisions.length === 0 ? (
          <p className="p-6 text-sm text-on-surface-variant">
            No decisions yet. Approve or reject from the review queue.
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="data-table min-w-[760px]">
              <thead>
                <tr>
                  <th scope="col">Time</th>
                  <th scope="col">Listing</th>
                  <th scope="col">Action</th>
                  <th scope="col">Risk</th>
                  <th scope="col">Reason</th>
                  <th scope="col">Moderator</th>
                  <th scope="col" className="text-right">
                    Link
                  </th>
                </tr>
              </thead>
              <tbody>
                {data.decisions.map((row) => (
                  <tr key={row.id}>
                    <td className="whitespace-nowrap text-on-surface-variant">
                      {formatDecisionTime(row.timestamp)}
                    </td>
                    <td className="max-w-[180px] truncate">{row.title}</td>
                    <td>
                      <span className={actionClass(row.action)}>
                        {row.action}
                      </span>
                    </td>
                    <td className="text-on-surface-variant">
                      {row.risk_tier ?? "—"}
                    </td>
                    <td className="max-w-[220px] truncate">{row.reason}</td>
                    <td className="font-mono text-xs text-on-surface-variant">
                      {shortUUID(row.moderator_id)}
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

        {data.decisions.length > 0 && (
          <footer className="flex items-center justify-between gap-4 border-t border-surface-variant p-4">
            <button
              type="button"
              disabled={page <= 1 || loading}
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
              disabled={page >= totalPages || loading}
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              className="btn-secondary min-h-[2.25rem] px-4"
            >
              Next
            </button>
          </footer>
        )}
      </section>
    </div>
  );
}
