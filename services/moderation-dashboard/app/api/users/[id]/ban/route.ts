import { MODERATOR_ID } from "@/lib/config";
import { fetchFromListingService } from "@/lib/listing-service";
import type { BanUserResponse } from "@/lib/types";

export async function POST(
  request: Request,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  const body = (await request.json()) as {
    reason?: string;
  };

  if (!body.reason?.trim()) {
    return new Response("Reason is required", { status: 400 });
  }

  try {
    const data = await fetchFromListingService<BanUserResponse>(
        `/v1/users/${id}/ban`, {
      method: "POST",
      body: JSON.stringify({
        moderator_id: MODERATOR_ID,
        reason: body.reason.trim(),
      }),
    });
    return Response.json(data);
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    let status = 502;
    if (message.startsWith("404:")) status = 404;
    if (message.startsWith("409:")) status = 409;
    return new Response(message, { status });
  }
}
