import type { Role } from "./auth";

export type NavItem = {
  href: string;
  label: string;
  shortLabel: string;
  icon: string;
  description: string;
  roles?: Role[];
  isActive: (pathname: string) => boolean;
};

export const PRIMARY_NAV: NavItem[] = [
  {
    href: "/queue",
    label: "Review queue",
    shortLabel: "Queue",
    icon: "inbox",
    description: "Listings awaiting moderation",
    isActive: (pathname) =>
      pathname === "/queue" ||
      pathname.startsWith("/listings/") ||
      pathname.startsWith("/users/"),
  },
  {
    href: "/audit",
    label: "Moderation history",
    shortLabel: "History",
    icon: "history",
    description: "Past approve and reject decisions",
    isActive: (pathname) => pathname === "/audit" || pathname.startsWith("/audit/"),
  },
  {
    href: "/analytics",
    label: "Analytics",
    shortLabel: "Analytics",
    icon: "monitoring",
    description: "Pipeline and queue metrics",
    isActive: (pathname) =>
      pathname === "/analytics" || pathname.startsWith("/analytics/"),
  },
];

export type BreadcrumbItem = {
  label: string;
  href?: string;
};

export const MINIMAL_LAYOUT_PATHS = ["/login", "/unauthorized"];

export function usesMinimalLayout(pathname: string): boolean {
  return MINIMAL_LAYOUT_PATHS.some(
    (path) => pathname === path || pathname.startsWith(`${path}/`)
  );
}
