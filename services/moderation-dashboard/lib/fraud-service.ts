import { FRAUD_SERVICE_URL } from "./config";

export async function fetchFromFraudService<T = unknown>(
  path: string,
  init?: RequestInit
): Promise<T> {
  const url = `${FRAUD_SERVICE_URL}${path}`;
  const res = await fetch(url, {
    cache: "no-store",
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`${res.status}: ${text}`);
  }
  return res.json() as Promise<T>;
}
