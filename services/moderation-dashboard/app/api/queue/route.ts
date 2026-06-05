import { fetchFromListingService } from "@/lib/listing-service";
import type { ListListingsResponse } from "@/lib/types";

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const page = searchParams.get("page") ?? "1";
  const limit = searchParams.get("limit") ?? "25";
  const tier = searchParams.get("tier");
  const sort = searchParams.get("sort") ?? "risk_score";

  const params = new URLSearchParams({
    status: "UNDER_REVIEW",
    page,
    limit,
    sort,
  });
  if (tier) {
    params.set("tier", tier);
  }

  try {
    const data = await fetchFromListingService<ListListingsResponse>(
      `/v1/listings?${params.toString()}`
    );
    return Response.json(data);
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    return new Response(message, { status: 502 });
  }
}
