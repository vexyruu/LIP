import { fetchFromListingService } from "@/lib/listing-service";
import type { CreateUploadResponse } from "@/lib/upload";

export async function POST(request: Request) {
  try {
    const body = await request.text();
    const data = await fetchFromListingService<CreateUploadResponse>(
      "/v1/uploads",
      {
        method: "POST",
        body,
      }
    );
    return Response.json(data);
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    return new Response(message, { status: 502 });
  }
}
