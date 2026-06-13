import bcrypt from "bcryptjs";
import type { Role } from "./auth";

export type ModeratorRecord = {
  id: string;
  email: string;
  displayName: string;
  role: Role;
  passwordHash: string;
};

type DBModeratorRow = {
  id: string;
  email: string;
  display_name: string;
  role: Role;
  password_hash: string;
};

// Seed list used when listing-service is unreachable (tests, offline dev).
// Default dev passwords: moderator123 / analyst123 / admin123
const MODERATORS_SEED: ModeratorRecord[] = [
  {
    id: "22222222-2222-2222-2222-222222222222",
    email: "mod@mlip.dev",
    displayName: "Moderator",
    role: "MODERATOR",
    passwordHash: "$2b$10$HKTuUf10h9ecIInNnVYxm.rlXFOkD/Kw3Da061ts/15BwMepvey4i",
  },
  {
    id: "44444444-4444-4444-4444-444444444444",
    email: "analyst@mlip.dev",
    displayName: "Analyst",
    role: "ANALYST",
    passwordHash: "$2b$10$XomwP8AWsj1DAPc86iJD8uC1I8YP4ot6WmwzLGYjaIEQcsHyP3hCC",
  },
  {
    id: "55555555-5555-5555-5555-555555555555",
    email: "admin@mlip.dev",
    displayName: "Admin",
    role: "ADMIN",
    passwordHash: "$2b$10$QFQ4TzWfzpjrbFbut0JWC.8g6sDn/QNStMUNidDg2.IAVgFOUkx8G",
  },
];

// Keep the exported name so tests that import MODERATORS still compile.
export const MODERATORS = MODERATORS_SEED;

function toRecord(row: DBModeratorRow): ModeratorRecord {
  return {
    id: row.id,
    email: row.email,
    displayName: row.display_name,
    role: row.role,
    passwordHash: row.password_hash,
  };
}

async function fetchModeratorByEmail(
  email: string
): Promise<ModeratorRecord | null> {
  const base =
    process.env.LISTING_SERVICE_URL ?? "http://localhost:8080";
  try {
    const res = await fetch(
      `${base}/v1/moderators/lookup?email=${encodeURIComponent(email)}`,
      { cache: "no-store" }
    );
    if (!res.ok) return null;
    const row = (await res.json()) as DBModeratorRow;
    return toRecord(row);
  } catch {
    return null;
  }
}

export function findModeratorByEmail(email: string): ModeratorRecord | undefined {
  const normalized = email.trim().toLowerCase();
  return MODERATORS_SEED.find((m) => m.email.toLowerCase() === normalized);
}

export async function verifyCredentials(
  email: string,
  password: string
): Promise<ModeratorRecord | null> {
  const normalized = email.trim().toLowerCase();

  // Try the DB first; fall back to the seed list so tests and offline dev work.
  const mod =
    (await fetchModeratorByEmail(normalized)) ??
    findModeratorByEmail(normalized);
  if (!mod) return null;

  const ok = await bcrypt.compare(password, mod.passwordHash);
  return ok ? mod : null;
}
