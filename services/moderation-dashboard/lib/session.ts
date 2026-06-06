import { cookies } from "next/headers";
import { SESSION_COOKIE, verifySession, type ModSession, type Role } from "./auth";
import { canModerate } from "./rbac";

// Reads and verifies the current session
export async function getSession(): Promise<ModSession | null> {
  const store = await cookies();
  const token = store.get(SESSION_COOKIE)?.value;
  if (!token) {
    return null;
  }
  return verifySession(token);
}

export type ApiAuthResult =
  | { ok: true; session: ModSession }
  | { ok: false; status: number; message: string };


export async function authorizeApi(options?: {
  roles?: Role[];
  requireModerate?: boolean;
}): Promise<ApiAuthResult> {
  const session = await getSession();
  if (!session) {
    return { ok: false, status: 401, message: "Unauthorized" };
  }
  if (options?.roles && !options.roles.includes(session.role)) {
    return { ok: false, status: 403, message: "Forbidden" };
  }
  if (options?.requireModerate && !canModerate(session.role)) {
    return { ok: false, status: 403, message: "Forbidden: moderation action requires Moderator or Admin role" };
  }
  return { ok: true, session };
}
