import bcrypt from "bcryptjs";
import type { Role } from "./auth";

export type ModeratorRecord = {
  id: string;
  email: string;
  displayName: string;
  role: Role;
  passwordHash: string;
};

// Local-dev stand-in for the Cloud SQL `moderators` table. In production this would be backed by the database and managed via the /admin/users page
// Default dev passwords: moderator123 / analyst123 / admin123
export const MODERATORS: ModeratorRecord[] = [
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

export function findModeratorByEmail(email: string): ModeratorRecord | undefined {
  const normalized = email.trim().toLowerCase();
  return MODERATORS.find((m) => m.email.toLowerCase() === normalized);
}

export async function verifyCredentials(
  email: string,
  password: string
): Promise<ModeratorRecord | null> {
  const mod = findModeratorByEmail(email);
  if (!mod) {
    return null;
  }
  const ok = await bcrypt.compare(password, mod.passwordHash);
  return ok ? mod : null;
}
