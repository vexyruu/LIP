"use client";

import { useState } from "react";
import { shortUUID } from "@/lib/format";
import type { AssignmentResponse } from "@/lib/types";

type Props = {
  listingId: string;
  assignedTo: string | null;
  moderatorId: string | null;
  reviewable: boolean;
};

export function AssignButton({
  listingId,
  assignedTo,
  moderatorId,
  reviewable,
}: Props) {
  const [assignee, setAssignee] = useState<string | null>(assignedTo);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mineClaim = assignee != null && assignee === moderatorId;
  const claimedByOther = assignee != null && assignee !== moderatorId;

  async function run(method: "POST" | "DELETE", optimistic: string | null) {
    const previous = assignee;
    setAssignee(optimistic);
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/api/listings/${listingId}/assign`, { method });
      if (!res.ok) {
        throw new Error(await res.text());
      }
      const data = (await res.json()) as AssignmentResponse;
      setAssignee(data.assigned_to);
    } catch (err) {
      setAssignee(previous);
      setError(err instanceof Error ? err.message : "Action failed");
    } finally {
      setLoading(false);
    }
  }

  if (!reviewable) {
    return null;
  }

  return (
    <div className="flex flex-col gap-2">
      {claimedByOther ? (
        <button
          type="button"
          disabled
          className="btn-secondary w-full opacity-60"
          title="Another moderator is reviewing this listing"
        >
          <span className="material-symbols-outlined text-lg" aria-hidden>
            lock
          </span>
          Claimed by {shortUUID(assignee as string)}
        </button>
      ) : mineClaim ? (
        <button
          type="button"
          disabled={loading}
          onClick={() => run("DELETE", null)}
          className="btn-secondary w-full"
        >
          <span className="material-symbols-outlined text-lg" aria-hidden>
            lock_open
          </span>
          {loading ? "Releasing…" : "Release claim"}
        </button>
      ) : (
        <button
          type="button"
          disabled={loading || !moderatorId}
          onClick={() => run("POST", moderatorId)}
          className="btn-primary w-full"
        >
          <span className="material-symbols-outlined text-lg" aria-hidden>
            how_to_reg
          </span>
          {loading ? "Claiming…" : "Claim for review"}
        </button>
      )}
      {error && (
        <p className="font-mono text-xs text-error" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}
