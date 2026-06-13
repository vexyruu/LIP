import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { SESSION_COOKIE, verifySession } from "@/lib/auth";
import { LISTING_SERVICE_URL } from "@/lib/config";

async function requireAdmin() {
  const cookieStore = await cookies();
  const token = cookieStore.get(SESSION_COOKIE)?.value ?? "";
  const session = await verifySession(token);
  if (!session || session.role !== "ADMIN") return null;
  return session;
}

export async function GET() {
  if (!(await requireAdmin())) {
    return new Response("Forbidden", { status: 403 });
  }
  const res = await fetch(`${LISTING_SERVICE_URL}/v1/moderators`, {
    cache: "no-store",
  });
  if (!res.ok) {
    return new Response("Failed to fetch moderators", { status: 502 });
  }
  return NextResponse.json(await res.json());
}

export async function POST(request: Request) {
  if (!(await requireAdmin())) {
    return new Response("Forbidden", { status: 403 });
  }
  const body = await request.json();
  const res = await fetch(`${LISTING_SERVICE_URL}/v1/moderators`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const text = await res.text();
  return new Response(text, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}
