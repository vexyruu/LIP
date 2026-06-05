import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { MODERATOR_ID } from "@/lib/config";
import { DEV_MODERATOR_EMAIL, SESSION_COOKIE, type ModSession } from "@/lib/auth";

export async function POST(request: Request) {
  const body = (await request.json()) as { email?: string };
  const email = body.email?.trim().toLowerCase();

  if (email !== DEV_MODERATOR_EMAIL.toLowerCase()) {
    return new Response("Invalid credentials", { status: 401 });
  }

  const session: ModSession = {
    moderatorId: MODERATOR_ID,
    email: DEV_MODERATOR_EMAIL,
  };

  const cookieStore = await cookies();
  cookieStore.set(SESSION_COOKIE, JSON.stringify(session), {
    httpOnly: true,
    sameSite: "lax",
    path: "/",
    maxAge: 60 * 60 * 24 * 7,
  });

  return NextResponse.json({ ok: true });
}