import { describe, it, expect } from "vitest";
import { signSession, verifySession, type ModSession } from "./auth";

const session: ModSession = {
  moderatorId: "22222222-2222-2222-2222-222222222222",
  email: "mod@mlip.dev",
  displayName: "Mona Moderator",
  role: "MODERATOR",
};

describe("session JWT", () => {
  it("signs and verifies a round-trip", async () => {
    const token = await signSession(session);
    expect(typeof token).toBe("string");

    const decoded = await verifySession(token);
    expect(decoded).toEqual(session);
  });

  it("returns null for a malformed token", async () => {
    expect(await verifySession("not-a-jwt")).toBeNull();
  });

  it("returns null for a tampered token", async () => {
    const token = await signSession(session);
    const tampered = `${token}tamper`;
    expect(await verifySession(tampered)).toBeNull();
  });
});
