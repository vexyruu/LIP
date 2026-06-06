"use client";

import { FormEvent, useState } from "react";
import { useSearchParams } from "next/navigation";
import { DEV_MODERATOR_EMAIL } from "@/lib/auth";

export function LoginClient() {
  const searchParams = useSearchParams();
  const [email, setEmail] = useState(DEV_MODERATOR_EMAIL);
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: email.trim().toLowerCase(), password }),
      });

      if (!res.ok) {
        throw new Error(await res.text());
      }

      const data = (await res.json()) as { redirectTo?: string };
      const requested = searchParams.get("from");
      const fallback = data.redirectTo ?? "/queue";
      const safeFrom =
        requested && requested.startsWith("/") && !requested.startsWith("/login")
          ? requested
          : fallback;
      window.location.href = safeFrom;
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to login");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center px-margin py-16">
      <div className="w-full max-w-md">
        <div className="mb-8 border-l-4 border-primary-container pl-4">
          <p className="mb-2 font-mono text-[11px] uppercase tracking-[0.2em] text-on-surface-variant">
            Internal access
          </p>
          <h1 className="text-3xl font-bold uppercase tracking-tight">
            MLIP Moderation
          </h1>
          <p className="mt-2 text-sm text-on-surface-variant">
            Sign in to review flagged listings and manage seller risk.
          </p>
        </div>

        <div className="panel p-6 md:p-8">
          <form className="flex flex-col gap-5" onSubmit={onSubmit} noValidate>
            <label className="flex flex-col gap-2">
              <span className="data-label">Email</span>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="min-h-[2.75rem] border border-outline bg-background px-4 py-2 font-mono text-sm text-on-surface"
                autoComplete="email"
                required
              />
            </label>

            <label className="flex flex-col gap-2">
              <span className="data-label">Password</span>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="min-h-[2.75rem] border border-outline bg-background px-4 py-2 font-mono text-sm text-on-surface"
                autoComplete="current-password"
                required
              />
            </label>

            {error && (
              <p className="status-banner status-banner-error text-sm" role="alert">
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={loading}
              className="btn-primary mt-1 w-full"
            >
              {loading ? "Signing in…" : "Sign in"}
            </button>
          </form>

          {process.env.NODE_ENV !== "production" && (
            <div className="mt-6 border-t border-outline pt-4">
              <p className="data-label mb-2">Local dev accounts</p>
              <ul className="space-y-1 font-mono text-xs text-on-surface-variant">
                <li>mod@mlip.dev · moderator123 · Moderator</li>
                <li>analyst@mlip.dev · analyst123 · Analyst</li>
                <li>admin@mlip.dev · admin123 · Admin</li>
              </ul>
              <p className="mt-3 text-xs text-on-surface-variant">
                Production would replace this credential check with marketplace SSO
                (OAuth). Roles map to the same JWT claims either way.
              </p>
            </div>
          )}
        </div>
      </div>
    </main>
  );
}
