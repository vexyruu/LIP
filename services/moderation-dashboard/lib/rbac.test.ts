import { describe, it, expect } from "vitest";
import { roleHome, canAccessPath, canModerate, roleLabel } from "./rbac";

describe("roleHome", () => {
  it("sends analysts to analytics and others to the queue", () => {
    expect(roleHome("ANALYST")).toBe("/analytics");
    expect(roleHome("MODERATOR")).toBe("/queue");
    expect(roleHome("ADMIN")).toBe("/queue");
  });
});

describe("canModerate", () => {
  it("allows moderators and admins only", () => {
    expect(canModerate("MODERATOR")).toBe(true);
    expect(canModerate("ADMIN")).toBe(true);
    expect(canModerate("ANALYST")).toBe(false);
  });
});

describe("canAccessPath", () => {
  it("restricts /admin and /api/admin to ADMIN", () => {
    expect(canAccessPath("ADMIN", "/admin")).toBe(true);
    expect(canAccessPath("ADMIN", "/admin/users")).toBe(true);
    expect(canAccessPath("ADMIN", "/api/admin/config")).toBe(true);

    expect(canAccessPath("MODERATOR", "/admin")).toBe(false);
    expect(canAccessPath("ANALYST", "/admin/users")).toBe(false);
    expect(canAccessPath("MODERATOR", "/api/admin/config")).toBe(false);
  });

  it("allows non-admin paths for every role", () => {
    expect(canAccessPath("ANALYST", "/queue")).toBe(true);
    expect(canAccessPath("ANALYST", "/analytics")).toBe(true);
    expect(canAccessPath("MODERATOR", "/listings/abc")).toBe(true);
  });

  it("does not treat lookalike prefixes as admin paths", () => {
    expect(canAccessPath("ANALYST", "/administrators")).toBe(true);
  });
});

describe("roleLabel", () => {
  it("maps roles to display labels", () => {
    expect(roleLabel("ADMIN")).toBe("Admin");
    expect(roleLabel("ANALYST")).toBe("Analyst");
    expect(roleLabel("MODERATOR")).toBe("Moderator");
  });
});
