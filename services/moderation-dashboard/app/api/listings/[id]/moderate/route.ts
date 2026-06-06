import { fetchFromListingService } from "@/lib/listing-service";
import { authorizeApi } from "@/lib/session";

export async function PATCH(
  request: Request,
  context: { params: Promise<{ id: string }> }
) {
  const auth = await authorizeApi({ requireModerate: true });
  if (!auth.ok) {
    return new Response(auth.message, { status: auth.status });
  }

  const { id } = await context.params;
  const body = (await request.json()) as {
    action: string;
    reason?: string;
  };

  try {
    const data = await fetchFromListingService(`/v1/listings/${id}`, {
      method: "PATCH",
      body: JSON.stringify({
        action: body.action,
        reason: body.reason ?? "",
        moderator_id: auth.session.moderatorId,
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
