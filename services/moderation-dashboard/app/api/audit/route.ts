import { fetchFromListingService } from "@/lib/listing-service";
import type { ListModerationDecisionsResponse } from "@/lib/types";

export async function GET(request: Request) {
    // Parse the query parameters
    const { searchParams } = new URL(request.url);
    const page = searchParams.get("page") ?? "1";
    const limit = searchParams.get("limit") ?? "25";
    const action = searchParams.get("action");

    // Build the URL parameters
    const params = new URLSearchParams({ page, limit });
    if (action && action !== "ALL") {
        params.set("action", action);
    }

    // Fetch the data from the listing service
    try {
        const data = await fetchFromListingService<ListModerationDecisionsResponse>(
            `/v1/moderation/decisions?${params.toString()}`
        );
        return Response.json(data);
    } catch (err) {
        const message = err instanceof Error ? err.message : "Unknown error";
        return new Response(message, { status: 502 });
    }
}