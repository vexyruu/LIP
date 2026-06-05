export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  await params;

  const form = await request.formData();
  const file = form.get("file");
  const uploadUrl = form.get("upload_url");
  const contentType = form.get("content_type");

  if (!(file instanceof File)) {
    return new Response("file is required", { status: 400 });
  }
  if (typeof uploadUrl !== "string" || uploadUrl.trim() === "") {
    return new Response("upload_url is required", { status: 400 });
  }

  const putRes = await fetch(uploadUrl, {
    method: "PUT",
    body: file,
    headers: {
      "Content-Type":
        typeof contentType === "string" && contentType
          ? contentType
          : file.type,
    },
  });

  if (!putRes.ok) {
    const text = await putRes.text();
    return new Response(text || `storage upload failed: ${putRes.status}`, {
      status: 502,
    });
  }

  return new Response(null, { status: 204 });
}
