import { SignJWT, jwtVerify } from "jose";

export const SESSION_COOKIE = "mlip_mod_session";

export const SESSION_TTL_SECONDS = 60 * 60 * 24 * 7;

export type Role = "MODERATOR" | "ANALYST" | "ADMIN";

export type ModSession = {
  moderatorId: string;
  email: string;
  displayName: string;
  role: Role;
};

// Default credentials shown on the login screen for local development
export const DEV_MODERATOR_EMAIL = "mod@mlip.dev";

function getSecret(): Uint8Array {
  const secret =
    process.env.SESSION_SECRET ??
    "dev-insecure-secret-change-me-in-production-000000000000";
  return new TextEncoder().encode(secret);
}

export async function signSession(session: ModSession): Promise<string> {
  return new SignJWT({
    email: session.email,
    displayName: session.displayName,
    role: session.role,
  })
    .setProtectedHeader({ alg: "HS256" })
    .setSubject(session.moderatorId)
    .setIssuedAt()
    .setExpirationTime(`${SESSION_TTL_SECONDS}s`)
    .sign(getSecret());
}

export async function verifySession(token: string): Promise<ModSession | null> {
  try {
    const { payload } = await jwtVerify(token, getSecret());
    const role = payload.role;
    if (role !== "MODERATOR" && role !== "ANALYST" && role !== "ADMIN") {
      return null;
    }
    return {
      moderatorId: String(payload.sub ?? ""),
      email: String(payload.email ?? ""),
      displayName: String(payload.displayName ?? ""),
      role,
    };
  } catch {
    return null;
  }
}
