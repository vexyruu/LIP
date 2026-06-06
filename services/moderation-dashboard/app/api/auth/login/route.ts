import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import {
  SESSION_COOKIE,
  SESSION_TTL_SECONDS,
  signSession,
  type ModSession,
} from "@/lib/auth";
import { verifyCredentials } from "@/lib/moderators";
import { roleHome } from "@/lib/rbac";

export async function POST(request: Request) {
  const body = (await request.json()) as { email?: string; password?: string };
  const email = body.email?.trim().toLowerCase() ?? "";
  const password = body.password ?? "";

  const moderator = await verifyCredentials(email, password);
  if (!moderator) {
    return new Response("Invalid credentials", { status: 401 });
  }

  const session: ModSession = {
    moderatorId: moderator.id,
    email: moderator.email,
    displayName: moderator.displayName,
    role: moderator.role,
  };

  const token = await signSession(session);

  const cookieStore = await cookies();
  cookieStore.set(SESSION_COOKIE, token, {
    httpOnly: true,
    sameSite: "lax",
    path: "/",
    maxAge: SESSION_TTL_SECONDS,
    secure: process.env.NODE_ENV === "production",
  });

  return NextResponse.json({
    ok: true,
    role: moderator.role,
    redirectTo: roleHome(moderator.role),
  });
}
