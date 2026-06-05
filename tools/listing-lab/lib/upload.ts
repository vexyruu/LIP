export interface CreateUploadResponse {
  upload_id: string;
  upload_url: string;
  public_url: string;
  expires_at: string;
}

export interface CompleteUploadResponse {
  upload_id: string;
  public_url: string;
  status: string;
}

export interface UploadProgress {
  filename: string;
  status: "pending" | "uploading" | "completing" | "done" | "error";
  publicUrl?: string;
  error?: string;
}

const ALLOWED_TYPES = new Set(["image/jpeg", "image/png", "image/webp"]);
const MAX_BYTES = 10 * 1024 * 1024;

export function validateUploadFile(file: File): string | null {
  if (!ALLOWED_TYPES.has(file.type)) {
    return `${file.name}: must be JPEG, PNG, or WebP`;
  }
  if (file.size <= 0 || file.size > MAX_BYTES) {
    return `${file.name}: must be between 1 byte and 10MB`;
  }
  return null;
}

export async function uploadListingImage(
  file: File,
  userId: string,
  onStatus?: (status: UploadProgress["status"]) => void
): Promise<string> {
  onStatus?.("pending");

  const initRes = await fetch("/api/uploads", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      user_id: userId,
      filename: file.name,
      content_type: file.type,
      size_bytes: file.size,
    }),
  });
  if (!initRes.ok) {
    throw new Error(await initRes.text());
  }

  const init = (await initRes.json()) as CreateUploadResponse;
  onStatus?.("uploading");

  const body = new FormData();
  body.append("file", file);
  body.append("upload_url", init.upload_url);
  body.append("content_type", file.type);

  const transferRes = await fetch(`/api/uploads/${init.upload_id}/transfer`, {
    method: "POST",
    body,
  });
  if (!transferRes.ok) {
    throw new Error(await transferRes.text());
  }

  onStatus?.("completing");
  const completeRes = await fetch(`/api/uploads/${init.upload_id}/complete`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ user_id: userId }),
  });
  if (!completeRes.ok) {
    throw new Error(await completeRes.text());
  }

  const complete = (await completeRes.json()) as CompleteUploadResponse;
  onStatus?.("done");
  return complete.public_url;
}
