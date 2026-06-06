import { describe, it, expect, vi, afterEach } from "vitest";
import { jsonFetcher } from "./hooks";

afterEach(() => {
  vi.restoreAllMocks();
});

describe("jsonFetcher", () => {
  it("returns parsed JSON on a 2xx response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({
        ok: true,
        json: async () => ({ total: 3 }),
      }))
    );

    const data = await jsonFetcher<{ total: number }>("/api/queue");
    expect(data).toEqual({ total: 3 });
  });

  it("throws the response body text on a non-2xx response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({
        ok: false,
        text: async () => "502: fraud-service down",
      }))
    );

    await expect(jsonFetcher("/api/users/1/risk")).rejects.toThrow(
      "502: fraud-service down"
    );
  });
});
