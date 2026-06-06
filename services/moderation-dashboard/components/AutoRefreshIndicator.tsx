"use client";

import { useEffect, useState } from "react";

// shows "updated Xs ago" with a pulsing dot
export function AutoRefreshIndicator({ lastUpdated }: { lastUpdated: number | null }) {
  const [now, setNow] = useState(lastUpdated ?? 0);

  useEffect(() => {
    setNow(Date.now());
    const timer = window.setInterval(() => setNow(Date.now()), 5000);
    return () => window.clearInterval(timer);
  }, []);

  if (!lastUpdated) {
    return null;
  }

  const seconds = Math.max(0, Math.floor((now - lastUpdated) / 1000));
  const label =
    seconds < 60 ? `${seconds}s ago` : `${Math.floor(seconds / 60)}m ago`;

  return (
    <span className="inline-flex items-center gap-2 font-mono text-xs text-on-surface-variant">
      <span
        className="inline-block h-2 w-2 animate-pulse rounded-full bg-alert-lime"
        aria-hidden
      />
      Updated {label}
    </span>
  );
}
