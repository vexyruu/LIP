"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  MODERATION_DASHBOARD_URL,
  POLL_INTERVAL_MS,
  POLL_TIMEOUT_MS,
} from "@/lib/config";
import { DEFAULT_FORM, FRAUD_SELLER_ID, PRESETS, SELLER_ID } from "@/lib/presets";
import type {
  CreateListingRequest,
  CreateListingResponse,
  HistoryEntry,
  ListingStatus,
  SubmitPhase,
} from "@/lib/types";
import {
  uploadListingImage,
  validateUploadFile,
  type UploadProgress,
} from "@/lib/upload";

const HISTORY_KEY = "listing-lab-history";
const MAX_HISTORY = 10;

function parseImages(raw: string): string[] {
  return raw
    .split(/\r?\n|,/)
    .map((line) => line.trim())
    .filter(Boolean);
}

function formatImages(images: string[]): string {
  return images.join("\n");
}

function statusClass(status: string): string {
  return `status-${status.toLowerCase()}`;
}

function loadHistory(): HistoryEntry[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = localStorage.getItem(HISTORY_KEY);
    if (!raw) return [];
    return JSON.parse(raw) as HistoryEntry[];
  } catch {
    return [];
  }
}

function saveHistory(entry: HistoryEntry) {
  const prev = loadHistory().filter((e) => e.listing_id !== entry.listing_id);
  const next = [entry, ...prev].slice(0, MAX_HISTORY);
  localStorage.setItem(HISTORY_KEY, JSON.stringify(next));
  return next;
}

function buildCurl(body: CreateListingRequest): string {
  const payload = JSON.stringify(body, null, 2);
  return [
    "curl -X POST http://localhost:8080/v1/listings \\",
    "  -H 'Content-Type: application/json' \\",
    `  -d '${payload.replace(/'/g, "'\\''")}'`,
  ].join("\n");
}

async function pollListing(
  listingId: string,
  onTick: (listing: ListingStatus, elapsedMs: number) => void,
  signal: AbortSignal
): Promise<ListingStatus> {
  const started = Date.now();

  while (!signal.aborted) {
    const elapsedMs = Date.now() - started;
    if (elapsedMs >= POLL_TIMEOUT_MS) {
      throw new Error(`Timed out after ${POLL_TIMEOUT_MS / 1000}s — still DRAFT`);
    }

    const res = await fetch(`/api/listings/${listingId}`, { signal });
    if (!res.ok) {
      throw new Error(await res.text());
    }

    const listing = (await res.json()) as ListingStatus;
    onTick(listing, elapsedMs);

    if (listing.status !== "DRAFT") {
      return listing;
    }

    await new Promise((resolve) => setTimeout(resolve, POLL_INTERVAL_MS));
  }

  throw new Error("Polling cancelled");
}

function ResultField({
  label,
  value,
}: {
  label: string;
  value: string | number | boolean | null | undefined;
}) {
  const display =
    value === null || value === undefined
      ? "—"
      : typeof value === "boolean"
        ? value
          ? "true"
          : "false"
        : String(value);

  return (
    <div className="lab-panel p-3">
      <p className="lab-label">{label}</p>
      <p className="font-mono text-sm break-all">{display}</p>
    </div>
  );
}

export function LabClient() {
  const [form, setForm] = useState<CreateListingRequest>(DEFAULT_FORM);
  const [imagesText, setImagesText] = useState(formatImages(DEFAULT_FORM.images));
  const [activePreset, setActivePreset] = useState<string>("jordan");
  const [phase, setPhase] = useState<SubmitPhase>("idle");
  const [error, setError] = useState<string | null>(null);
  const [acceptResponse, setAcceptResponse] =
    useState<CreateListingResponse | null>(null);
  const [listing, setListing] = useState<ListingStatus | null>(null);
  const [elapsedMs, setElapsedMs] = useState(0);
  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const [showRaw, setShowRaw] = useState(false);
  const [spamCount, setSpamCount] = useState(0);
  const [uploadBusy, setUploadBusy] = useState(false);
  const [uploadProgress, setUploadProgress] = useState<UploadProgress[]>([]);
  const [uploadError, setUploadError] = useState<string | null>(null);

  useEffect(() => {
    setHistory(loadHistory());
  }, []);

  const requestBody = useMemo(
    (): CreateListingRequest => ({
      ...form,
      images: parseImages(imagesText),
    }),
    [form, imagesText]
  );

  const curlCommand = useMemo(() => buildCurl(requestBody), [requestBody]);

  const applyPreset = useCallback((presetId: string) => {
    const preset = PRESETS.find((p) => p.id === presetId);
    if (!preset) return;
    setActivePreset(presetId);
    setForm(preset.data);
    setImagesText(formatImages(preset.data.images));
  }, []);

  const updateField = <K extends keyof CreateListingRequest>(
    key: K,
    value: CreateListingRequest[K]
  ) => {
    setActivePreset("");
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  const submitOnce = useCallback(
    async (body: CreateListingRequest, signal: AbortSignal) => {
      setPhase("submitting");
      setError(null);
      setAcceptResponse(null);
      setListing(null);
      setElapsedMs(0);

      const res = await fetch("/api/submit", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
        signal,
      });

      if (!res.ok) {
        throw new Error(await res.text());
      }

      const accepted = (await res.json()) as CreateListingResponse;
      setAcceptResponse(accepted);
      setPhase("polling");

      const finalListing = await pollListing(
        accepted.listing_id,
        (current, elapsed) => {
          setListing(current);
          setElapsedMs(elapsed);
        },
        signal
      );

      setListing(finalListing);
      setPhase("done");

      const entry: HistoryEntry = {
        listing_id: finalListing.listing_id,
        title: finalListing.title,
        status: finalListing.status,
        submitted_at: new Date().toISOString(),
      };
      setHistory(saveHistory(entry));

      return finalListing;
    },
    []
  );

  const handleSubmit = useCallback(async () => {
    const controller = new AbortController();
    try {
      await submitOnce(requestBody, controller.signal);
    } catch (err) {
      if (controller.signal.aborted) return;
      const message = err instanceof Error ? err.message : "Submit failed";
      setError(message);
      setPhase(message.includes("Timed out") ? "timeout" : "error");
    }
  }, [requestBody, submitOnce]);

  const handleSpamBurst = useCallback(async () => {
    const controller = new AbortController();
    setSpamCount(0);
    setError(null);

    for (let i = 0; i < 5; i += 1) {
      if (controller.signal.aborted) return;
      const body: CreateListingRequest = {
        ...requestBody,
        title: `${requestBody.title} (#${i + 1})`,
        price_ask: Math.max(5, requestBody.price_ask - i * 3),
      };

      try {
        setSpamCount(i + 1);
        await submitOnce(body, controller.signal);
      } catch (err) {
        if (controller.signal.aborted) return;
        const message = err instanceof Error ? err.message : "Burst submit failed";
        setError(message);
        setPhase("error");
        return;
      }
    }
  }, [requestBody, submitOnce]);

  const busy = phase === "submitting" || phase === "polling" || uploadBusy;
  const handleFileUpload = useCallback(
    async (files: FileList | null) => {
      if (!files?.length) return;

      const selected = Array.from(files);
      for (const file of selected) {
        const validationError = validateUploadFile(file);
        if (validationError) {
          setUploadError(validationError);
          return;
        }
      }

      const currentCount = parseImages(imagesText).length;
      if (currentCount + selected.length > 10) {
        setUploadError("Listings support at most 10 images");
        return;
      }

      setUploadError(null);
      setUploadBusy(true);
      setUploadProgress(
        selected.map((file) => ({ filename: file.name, status: "pending" }))
      );

      const uploadedUrls: string[] = [];

      try {
        for (let i = 0; i < selected.length; i += 1) {
          const file = selected[i];
          const publicUrl = await uploadListingImage(
            file,
            form.user_id,
            (status) => {
              setUploadProgress((prev) =>
                prev.map((item, index) =>
                  index === i ? { ...item, status } : item
                )
              );
            }
          );
          uploadedUrls.push(publicUrl);
          setUploadProgress((prev) =>
            prev.map((item, index) =>
              index === i ? { ...item, status: "done", publicUrl } : item
            )
          );
        }

        setActivePreset("");
        setImagesText((prev) => {
          const existing = parseImages(prev);
          return formatImages([...existing, ...uploadedUrls]);
        });
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Image upload failed";
        setUploadError(message);
        setUploadProgress((prev) =>
          prev.map((item) =>
            item.status === "done"
              ? item
              : { ...item, status: "error", error: message }
          )
        );
      } finally {
        setUploadBusy(false);
      }
    },
    [form.user_id, imagesText]
  );

  const dashboardUrl = listing
    ? `${MODERATION_DASHBOARD_URL}/listings/${listing.listing_id}`
    : null;

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <header className="mb-8 border-b border-[var(--color-lab-border)] pb-6">
        <p className="text-xs uppercase tracking-[0.2em] text-[var(--color-lab-muted)]">
          MLIP · dev only
        </p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">Listing Lab</h1>
        <p className="mt-2 max-w-2xl text-sm text-[var(--color-lab-muted)]">
          Seller simulator — submit to the pipeline, poll until analysis finishes,
          then open the result in the moderation dashboard.
        </p>
      </header>

      <div className="grid gap-6 lg:grid-cols-[280px_1fr]">
        <aside className="space-y-4">
          <section className="lab-panel p-4">
            <h2 className="lab-label mb-3">Presets</h2>
            <div className="space-y-2">
              {PRESETS.map((preset) => (
                <button
                  key={preset.id}
                  type="button"
                  className="lab-btn lab-preset w-full"
                  data-active={activePreset === preset.id}
                  onClick={() => applyPreset(preset.id)}
                  disabled={busy}
                >
                  <span className="block font-medium">{preset.label}</span>
                  <span className="mt-1 block text-xs text-[var(--color-lab-muted)]">
                    {preset.description}
                  </span>
                </button>
              ))}
            </div>
          </section>

          <section className="lab-panel p-4">
            <div className="mb-3 flex items-center justify-between">
              <h2 className="lab-label mb-0">History</h2>
              {history.length > 0 && (
                <button
                  type="button"
                  className="text-xs text-[var(--color-lab-muted)] hover:text-[var(--color-lab-accent)]"
                  onClick={() => {
                    localStorage.removeItem(HISTORY_KEY);
                    setHistory([]);
                  }}
                >
                  Clear
                </button>
              )}
            </div>
            {history.length === 0 ? (
              <p className="text-xs text-[var(--color-lab-muted)]">
                Last {MAX_HISTORY} submits appear here.
              </p>
            ) : (
              <ul className="space-y-2">
                {history.map((entry) => (
                  <li key={entry.listing_id} className="text-xs">
                    <a
                      href={`${MODERATION_DASHBOARD_URL}/listings/${entry.listing_id}`}
                      className="font-mono text-[var(--color-lab-accent)] hover:underline"
                      target="_blank"
                      rel="noreferrer"
                    >
                      {entry.listing_id.slice(0, 8)}…
                    </a>
                    <p className="truncate text-[var(--color-lab-muted)]">
                      {entry.title}
                    </p>
                    <p className={statusClass(entry.status)}>{entry.status}</p>
                  </li>
                ))}
              </ul>
            )}
          </section>
        </aside>

        <main className="space-y-6">
          <form
            className="lab-panel p-5 space-y-4"
            onSubmit={(e) => {
              e.preventDefault();
              void handleSubmit();
            }}
          >
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="sm:col-span-2">
                <label className="lab-label" htmlFor="user_id">
                  Seller user_id
                </label>
                <input
                  id="user_id"
                  className="lab-input"
                  value={form.user_id}
                  onChange={(e) => updateField("user_id", e.target.value)}
                  disabled={busy}
                />
                <p className="mt-1 text-xs text-[var(--color-lab-muted)]">
                  Clean seller (default): {SELLER_ID}. Fraud demo:{" "}
                  {FRAUD_SELLER_ID}
                </p>
              </div>

              <div className="sm:col-span-2">
                <label className="lab-label" htmlFor="title">
                  Title
                </label>
                <input
                  id="title"
                  className="lab-input"
                  value={form.title}
                  onChange={(e) => updateField("title", e.target.value)}
                  disabled={busy}
                />
              </div>

              <div className="sm:col-span-2">
                <label className="lab-label" htmlFor="description">
                  Description
                </label>
                <textarea
                  id="description"
                  className="lab-input min-h-24 resize-y"
                  value={form.description}
                  onChange={(e) => updateField("description", e.target.value)}
                  disabled={busy}
                />
              </div>

              <div>
                <label className="lab-label" htmlFor="price_ask">
                  Price ask
                </label>
                <input
                  id="price_ask"
                  type="number"
                  min={0.01}
                  step={0.01}
                  className="lab-input"
                  value={form.price_ask}
                  onChange={(e) =>
                    updateField("price_ask", Number(e.target.value))
                  }
                  disabled={busy}
                />
              </div>

              <div>
                <label className="lab-label" htmlFor="condition">
                  Condition (1–5)
                </label>
                <input
                  id="condition"
                  type="number"
                  min={1}
                  max={5}
                  className="lab-input"
                  value={form.condition}
                  onChange={(e) =>
                    updateField("condition", Number(e.target.value))
                  }
                  disabled={busy}
                />
              </div>

              <div>
                <label className="lab-label" htmlFor="category_id">
                  Category ID
                </label>
                <input
                  id="category_id"
                  type="number"
                  min={1}
                  className="lab-input"
                  value={form.category_id}
                  onChange={(e) =>
                    updateField("category_id", Number(e.target.value))
                  }
                  disabled={busy}
                />
              </div>

              <div className="sm:col-span-2">
                <label className="lab-label" htmlFor="listing-images">
                  Upload photos
                </label>
                <input
                  id="listing-images"
                  type="file"
                  accept="image/jpeg,image/png,image/webp"
                  multiple
                  className="lab-input py-2 file:mr-3 file:border-0 file:bg-[var(--color-lab-accent)] file:px-3 file:py-1 file:text-sm file:font-semibold file:text-[#0a0a0a]"
                  disabled={busy}
                  onChange={(e) => {
                    void handleFileUpload(e.target.files);
                    e.target.value = "";
                  }}
                />
                <p className="mt-1 text-xs text-[var(--color-lab-muted)]">
                  JPEG, PNG, or WebP · max 10MB each · up to 10 images per listing
                </p>
                {uploadError && (
                  <p className="mt-2 text-xs text-[var(--color-lab-danger)]">
                    {uploadError}
                  </p>
                )}
                {uploadProgress.length > 0 && (
                  <ul className="mt-3 space-y-1 text-xs font-mono">
                    {uploadProgress.map((item) => (
                      <li key={item.filename} className="text-[var(--color-lab-muted)]">
                        {item.filename} · {item.status}
                        {item.publicUrl ? (
                          <span className="block truncate text-[var(--color-lab-accent)]">
                            {item.publicUrl}
                          </span>
                        ) : null}
                      </li>
                    ))}
                  </ul>
                )}
              </div>

              <div className="sm:col-span-2">
                <label className="lab-label" htmlFor="images">
                  Image URLs (one per line)
                </label>
                <textarea
                  id="images"
                  className="lab-input min-h-20 resize-y"
                  value={imagesText}
                  onChange={(e) => {
                    setActivePreset("");
                    setImagesText(e.target.value);
                  }}
                  disabled={busy}
                  placeholder="https://images.unsplash.com/..."
                />
              </div>
            </div>

            <div className="flex flex-wrap gap-2 pt-2">
              <button
                type="submit"
                className="lab-btn lab-btn-primary"
                disabled={busy}
              >
                {phase === "submitting"
                  ? "Submitting…"
                  : phase === "polling"
                    ? "Polling…"
                    : "Submit listing"}
              </button>
              <button
                type="button"
                className="lab-btn"
                disabled={busy}
                onClick={() => void handleSpamBurst()}
              >
                Spam burst (5×)
              </button>
              <button
                type="button"
                className="lab-btn"
                disabled={busy}
                onClick={() => navigator.clipboard.writeText(curlCommand)}
              >
                Copy as curl
              </button>
            </div>
          </form>

          {(phase !== "idle" || error) && (
            <section className="lab-panel p-5 space-y-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-sm font-semibold uppercase tracking-wide">
                  Pipeline
                </h2>
                {spamCount > 0 && phase !== "idle" && (
                  <span className="text-xs text-[var(--color-lab-muted)]">
                    Burst {spamCount}/5
                  </span>
                )}
              </div>

              {error && (
                <p className="rounded border border-[var(--color-lab-danger)] bg-[#2d1214] px-3 py-2 text-sm text-[var(--color-lab-danger)]">
                  {error}
                </p>
              )}

              <div className="flex flex-wrap items-center gap-3 text-sm font-mono">
                <span className={phase === "submitting" ? "text-[var(--color-lab-accent)]" : "text-[var(--color-lab-muted)]"}>
                  POST /v1/listings
                </span>
                <span className="text-[var(--color-lab-muted)]">→</span>
                <span className={phase === "polling" ? "text-[var(--color-lab-accent)]" : listing?.status ? statusClass(listing.status) : "text-[var(--color-lab-muted)]"}>
                  {listing?.status ?? (phase === "polling" ? "DRAFT (polling)" : "awaiting")}
                </span>
                {elapsedMs > 0 && (
                  <span className="text-xs text-[var(--color-lab-muted)]">
                    {(elapsedMs / 1000).toFixed(1)}s
                  </span>
                )}
              </div>

              {acceptResponse && (
                <div className="text-xs font-mono text-[var(--color-lab-muted)]">
                  202 accepted · listing_id{" "}
                  <span className="text-[var(--color-lab-accent)]">
                    {acceptResponse.listing_id}
                  </span>
                  {" · "}
                  eta {acceptResponse.eta_ms}ms
                </div>
              )}

              {listing && (
                <>
                  <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    <ResultField label="Status" value={listing.status} />
                    <ResultField label="Risk tier" value={listing.risk_tier} />
                    <ResultField label="Risk score" value={listing.risk_score} />
                    <ResultField
                      label="Suggested price"
                      value={listing.suggested_price}
                    />
                    <ResultField label="Brand" value={listing.extracted_brand} />
                    <ResultField
                      label="Product"
                      value={listing.extracted_product}
                    />
                    <ResultField label="Size" value={listing.extracted_size} />
                    <ResultField
                      label="Policy violation"
                      value={listing.policy_violation}
                    />
                  </div>

                  <div className="flex flex-wrap gap-2">
                    {dashboardUrl && (
                      <a
                        href={dashboardUrl}
                        className="lab-btn lab-btn-primary inline-block"
                        target="_blank"
                        rel="noreferrer"
                      >
                        Open in moderation dashboard
                      </a>
                    )}
                    <button
                      type="button"
                      className="lab-btn"
                      onClick={() => setShowRaw((v) => !v)}
                    >
                      {showRaw ? "Hide" : "Show"} raw JSON
                    </button>
                  </div>

                  {showRaw && (
                    <pre className="overflow-x-auto border border-[var(--color-lab-border)] bg-[var(--color-lab-bg)] p-3 text-xs font-mono">
                      {JSON.stringify(listing, null, 2)}
                    </pre>
                  )}
                </>
              )}
            </section>
          )}
        </main>
      </div>
    </div>
  );
}
