import { fetchFromListingService } from "@/lib/listing-service";
import type { ListingDetail } from "@/lib/types";

export async function GET(
  _request: Request,
  context: { params: Promise<{ id: string }> }
) {
  const { id } = await context.params;
  try {
    const data = await fetchFromListingService<ListingDetail>(
      `/v1/listings/${id}`
    );
    return Response.json(data);
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    const status = message.startsWith("404:") ? 404 : 502;
    return new Response(message, { status });
  }
}
