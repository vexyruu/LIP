import { fetchFromListingService } from "@/lib/listing-service";
import { authorizeApi } from "@/lib/session";
import type { AssignmentResponse } from "@/lib/types";

function errorStatus(message: string): number {
  if (message.startsWith("404:")) return 404;
  if (message.startsWith("409:")) return 409;
  return 502;
}

// claim the listing for the current moderator
export async function POST(
  _request: Request,
  context: { params: Promise<{ id: string }> }
) {
  const auth = await authorizeApi({ requireModerate: true });
  if (!auth.ok) {
    return new Response(auth.message, { status: auth.status });
  }

  const { id } = await context.params;
  try {
    const data = await fetchFromListingService<AssignmentResponse>(
      `/v1/listings/${id}/assign`,
      {
        method: "POST",
        body: JSON.stringify({ moderator_id: auth.session.moderatorId }),
      }
    );
    return Response.json(data);
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    return new Response(message, { status: errorStatus(message) });
  }
}

// Release the current moderator's claim.
export async function DELETE(
  _request: Request,
  context: { params: Promise<{ id: string }> }
) {
  const auth = await authorizeApi({ requireModerate: true });
  if (!auth.ok) {
    return new Response(auth.message, { status: auth.status });
  }

  const { id } = await context.params;
  try {
    const data = await fetchFromListingService<AssignmentResponse>(
      `/v1/listings/${id}/assign`,
      {
        method: "DELETE",
        body: JSON.stringify({ moderator_id: auth.session.moderatorId }),
      }
    );
    return Response.json(data);
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    return new Response(message, { status: errorStatus(message) });
  }
}
