"use client";

import { useState } from "react";
import useSWR from "swr";
import bcrypt from "bcryptjs";
import { PageHeader } from "@/components/layout/PageHeader";
import { SectionHeader } from "@/components/SectionHeader";
import { jsonFetcher } from "@/lib/hooks";

type ModeratorRow = {
  id: string;
  email: string;
  display_name: string;
  role: string;
  password_hash: string;
};

const ROLES = ["MODERATOR", "ANALYST", "ADMIN"] as const;

export function AdminUsersClient() {
  const { data, error, isLoading, mutate } = useSWR<ModeratorRow[]>(
    "/api/admin/moderators",
    jsonFetcher
  );

  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [role, setRole] = useState<(typeof ROLES)[number]>("MODERATOR");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault();
    setFormError(null);
    setSubmitting(true);
    try {
      const passwordHash = await bcrypt.hash(password, 10);
      const res = await fetch("/api/admin/moderators", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email: email.trim().toLowerCase(),
          display_name: displayName.trim(),
          role,
          password_hash: passwordHash,
        }),
      });
      if (!res.ok) {
        setFormError(await res.text());
        return;
      }
      setEmail("");
      setDisplayName("");
      setPassword("");
      setRole("MODERATOR");
      mutate();
    } catch (err) {
      setFormError(String(err));
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Remove this moderator?")) return;
    await fetch(`/api/admin/moderators/${id}`, { method: "DELETE" });
    mutate();
  }

  return (
    <div className="page-shell">
      <PageHeader
        title="Moderator accounts"
        description="Add or remove dashboard users. Passwords are hashed in the browser before being stored."
      />

      <section className="panel overflow-hidden">
        <div className="panel-header">Active accounts</div>
        {isLoading && (
          <p className="p-4 text-sm text-on-surface-variant">Loading…</p>
        )}
        {error && (
          <p className="p-4 text-sm text-alert-coral">
            {(error as Error).message}
          </p>
        )}
        {data && (
          <div className="overflow-x-auto">
            <table className="data-table min-w-[500px]">
              <thead>
                <tr>
                  <th scope="col">Email</th>
                  <th scope="col">Name</th>
                  <th scope="col">Role</th>
                  <th scope="col" className="text-right">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {data.map((mod) => (
                  <tr key={mod.id}>
                    <td className="font-mono text-xs">{mod.email}</td>
                    <td>{mod.display_name}</td>
                    <td>
                      <span className="font-mono text-xs uppercase tracking-wider">
                        {mod.role}
                      </span>
                    </td>
                    <td className="text-right">
                      <button
                        onClick={() => handleDelete(mod.id)}
                        className="font-mono text-xs text-alert-coral underline-offset-2 hover:underline"
                      >
                        Remove
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section className="panel p-4">
        <SectionHeader title="Add account" />
        <form onSubmit={handleAdd} className="flex flex-col gap-3 max-w-sm">
          <input
            className="input"
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
          <input
            className="input"
            type="text"
            placeholder="Display name"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            required
          />
          <select
            className="input"
            value={role}
            onChange={(e) => setRole(e.target.value as (typeof ROLES)[number])}
          >
            {ROLES.map((r) => (
              <option key={r} value={r}>
                {r}
              </option>
            ))}
          </select>
          <input
            className="input"
            type="password"
            placeholder="Password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
          />
          {formError && (
            <p className="text-sm text-alert-coral">{formError}</p>
          )}
          <button type="submit" className="btn-primary" disabled={submitting}>
            {submitting ? "Adding…" : "Add moderator"}
          </button>
        </form>
      </section>
    </div>
  );
}
