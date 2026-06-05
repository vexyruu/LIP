import { fetchFromListingService } from "@/lib/listing-service";
import type {
  CreateListingRequest,
  CreateListingResponse,
} from "@/lib/types";

export async function POST(request: Request) {
  let body: CreateListingRequest;
  try {
    body = (await request.json()) as CreateListingRequest;
  } catch {
    return new Response("invalid request body", { status: 400 });
  }

  try {
    const data = await fetchFromListingService<CreateListingResponse>(
      "/v1/listings",
      {
        method: "POST",
        body: JSON.stringify(body),
      }
    );
    return Response.json(data, { status: 202 });
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    return new Response(message, { status: 502 });
  }
}
