import { fetchFromListingService } from "@/lib/listing-service";
import type { AnalyticsSummary } from "@/lib/types";

export async function GET() {
  try {
    const data = await fetchFromListingService<AnalyticsSummary>(
      "/v1/analytics/summary"
    );
    return Response.json(data);
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    return new Response(message, { status: 502 });
  }
}
