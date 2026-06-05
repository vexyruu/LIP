import { fetchFromListingService } from "@/lib/listing-service";
import type { CompleteUploadResponse } from "@/lib/upload";

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  try {
    const body = await request.text();
    const data = await fetchFromListingService<CompleteUploadResponse>(
      `/v1/uploads/${id}/complete`,
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
