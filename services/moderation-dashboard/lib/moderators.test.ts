import { describe, it, expect } from "vitest";
import { findModeratorByEmail, verifyCredentials } from "./moderators";

describe("findModeratorByEmail", () => {
  it("matches case-insensitively and trims whitespace", () => {
    expect(findModeratorByEmail("  MOD@MLIP.DEV ")?.role).toBe("MODERATOR");
    expect(findModeratorByEmail("analyst@mlip.dev")?.role).toBe("ANALYST");
  });

  it("returns undefined for unknown emails", () => {
    expect(findModeratorByEmail("nobody@mlip.dev")).toBeUndefined();
  });
});

describe("verifyCredentials", () => {
  it("returns the record for valid credentials", async () => {
    const mod = await verifyCredentials("mod@mlip.dev", "moderator123");
    expect(mod?.role).toBe("MODERATOR");
    const admin = await verifyCredentials("admin@mlip.dev", "admin123");
    expect(admin?.role).toBe("ADMIN");
  });

  it("returns null for a wrong password", async () => {
    expect(await verifyCredentials("mod@mlip.dev", "wrong")).toBeNull();
  });

  it("returns null for an unknown email", async () => {
    expect(await verifyCredentials("nobody@mlip.dev", "whatever")).toBeNull();
  });
});
