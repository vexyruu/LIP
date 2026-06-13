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

export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  if (!(await requireAdmin())) {
    return new Response("Forbidden", { status: 403 });
  }
  const { id } = await params;
  const res = await fetch(`${LISTING_SERVICE_URL}/v1/moderators/${id}`, {
    method: "DELETE",
  });
  return new Response(null, { status: res.status });
}
