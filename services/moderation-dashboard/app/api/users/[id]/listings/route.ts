import { fetchFromListingService } from "@/lib/listing-service";
import type { ListUserListingsResponse } from "@/lib/types";

export async function GET(
  _request: Request,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  const url = new URL(_request.url);
  const page = url.searchParams.get("page") ?? "1";
  const limit = url.searchParams.get("limit") ?? "10";

  try {
    const data = await fetchFromListingService<ListUserListingsResponse>(
      `/v1/users/${id}/listings?page=${page}&limit=${limit}`
    );
    return Response.json(data);
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    const status = message.startsWith("404:") ? 404 : 502;
    return new Response(message, { status });
  }
}
