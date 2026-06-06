import type { Role } from "./auth";

const ADMIN_ONLY_PREFIXES = ["/admin", "/api/admin"];

function matchesPrefix(pathname: string, prefix: string): boolean {
  return pathname === prefix || pathname.startsWith(`${prefix}/`);
}

// Landing page after login, per role. Analysts default to analytics since they cannot take moderation actions
export function roleHome(role: Role): string {
  return role === "ANALYST" ? "/analytics" : "/queue";
}

export function canAccessPath(role: Role, pathname: string): boolean {
  if (ADMIN_ONLY_PREFIXES.some((p) => matchesPrefix(pathname, p))) {
    return role === "ADMIN";
  }
  return true;
}

// Whether a role may take moderation actions (approve, reject, ban)
export function canModerate(role: Role): boolean {
  return role === "MODERATOR" || role === "ADMIN";
}

export function roleLabel(role: Role): string {
  switch (role) {
    case "ADMIN":
      return "Admin";
    case "ANALYST":
      return "Analyst";
    default:
      return "Moderator";
  }
}
