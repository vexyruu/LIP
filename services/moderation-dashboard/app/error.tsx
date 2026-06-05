"use client";

export default function Error({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="page-shell">
      <p className="font-mono text-sm uppercase tracking-widest text-alert-coral">
        Runtime error
      </p>
      <h1 className="page-title">Something broke</h1>
      <p className="page-description">
        The page failed to render. Retry or return to the review queue.
      </p>
      <div className="flex flex-wrap gap-3">
        <button type="button" onClick={reset} className="btn-primary">
          Try again
        </button>
        <a href="/queue" className="btn-secondary">
          Back to queue
        </a>
      </div>
    </div>
  );
}
