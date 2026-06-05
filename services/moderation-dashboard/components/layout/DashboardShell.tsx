"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { PRIMARY_NAV, usesMinimalLayout } from "@/lib/navigation";

export function DashboardShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [mobileNavOpen, setMobileNavOpen] = useState(false);

  useEffect(() => {
    setMobileNavOpen(false);
  }, [pathname]);

  useEffect(() => {
    if (!mobileNavOpen) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") setMobileNavOpen(false);
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [mobileNavOpen]);

  if (usesMinimalLayout(pathname)) {
    return <>{children}</>;
  }

  const activeNav = PRIMARY_NAV.find((item) => item.isActive(pathname));

  return (
    <div className="dashboard-shell">
      <a href="#main-content" className="skip-link">
        Skip to main content
      </a>

      {mobileNavOpen && (
        <button
          type="button"
          className="sidebar-backdrop lg:hidden"
          aria-label="Close navigation menu"
          onClick={() => setMobileNavOpen(false)}
        />
      )}

      <aside
        id="dashboard-sidebar"
        className={`dashboard-sidebar ${mobileNavOpen ? "sidebar-open" : ""}`}
        aria-label="Primary navigation"
      >
        <div className="sidebar-brand">
          <Link href="/queue" className="sidebar-brand-link">
            <span className="sidebar-brand-mark" aria-hidden>
              M
            </span>
            <span className="sidebar-brand-text">
              <span className="sidebar-brand-title">MLIP</span>
              <span className="sidebar-brand-subtitle">Moderation</span>
            </span>
          </Link>
        </div>

        <nav className="sidebar-nav">
          <p className="sidebar-nav-label">Workspaces</p>
          <ul className="sidebar-nav-list">
            {PRIMARY_NAV.map((item) => {
              const active = item.isActive(pathname);
              return (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    className={`sidebar-nav-link ${active ? "sidebar-nav-link-active" : ""}`}
                    aria-current={active ? "page" : undefined}
                  >
                    <span
                      className="material-symbols-outlined sidebar-nav-icon"
                      aria-hidden
                    >
                      {item.icon}
                    </span>
                    <span className="sidebar-nav-copy">
                      <span className="sidebar-nav-title">{item.label}</span>
                      <span className="sidebar-nav-desc">{item.description}</span>
                    </span>
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>

        <div className="sidebar-footer">
          <div className="sidebar-user">
            <span
              className="material-symbols-outlined sidebar-user-icon"
              aria-hidden
            >
              account_circle
            </span>
            <div className="sidebar-user-copy">
              <span className="sidebar-user-email">mod@mlip.dev</span>
              <span className="sidebar-user-role">Moderator</span>
            </div>
          </div>
          <button
            type="button"
            className="sidebar-logout"
            onClick={async () => {
              await fetch("/api/auth/logout", { method: "POST" });
              window.location.href = "/login";
            }}
          >
            <span className="material-symbols-outlined text-base" aria-hidden>
              logout
            </span>
            Sign out
          </button>
        </div>
      </aside>

      <div className="dashboard-main">
        <div className="dashboard-topbar lg:hidden">
          <button
            type="button"
            className="topbar-menu-btn"
            aria-expanded={mobileNavOpen}
            aria-controls="dashboard-sidebar"
            onClick={() => setMobileNavOpen((open) => !open)}
          >
            <span className="material-symbols-outlined" aria-hidden>
              menu
            </span>
            <span className="sr-only">Open navigation menu</span>
          </button>
          <p className="topbar-title">{activeNav?.shortLabel ?? "MLIP"}</p>
        </div>

        <main id="main-content" className="dashboard-content" tabIndex={-1}>
          {children}
        </main>
      </div>
    </div>
  );
}
