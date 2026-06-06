"use client";

import useSWR, { type SWRConfiguration } from "swr";
import type {
  AnalyticsSummary,
  ListListingsResponse,
  ListingDetail,
  ListModerationDecisionsResponse,
  ListUserListingsResponse,
  SellerRisk,
  UserProfile,
} from "./types";

// shared fetcher
export async function jsonFetcher<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(await res.text());
  }
  return (await res.json()) as T;
}

const QUEUE_POLL_MS = 30_000;

export type QueueParams = {
  page: number;
  limit: number;
  sort: string;
  tier: string;
};

export function useQueue(params: QueueParams, options?: SWRConfiguration) {
  const search = new URLSearchParams({
    page: String(params.page),
    limit: String(params.limit),
    sort: params.sort,
  });
  if (params.tier !== "ALL") {
    search.set("tier", params.tier);
  }
  return useSWR<ListListingsResponse>(
    `/api/queue?${search.toString()}`,
    jsonFetcher,
    { refreshInterval: QUEUE_POLL_MS, keepPreviousData: true, ...options }
  );
}

export function useListing(listingId: string, options?: SWRConfiguration) {
  return useSWR<ListingDetail>(
    listingId ? `/api/listings/${listingId}` : null,
    jsonFetcher,
    options
  );
}

// seller risk lives on the fraud-service and can be down
export function useSellerRisk(userId: string | undefined, options?: SWRConfiguration) {
  return useSWR<SellerRisk>(
    userId ? `/api/users/${userId}/risk` : null,
    jsonFetcher,
    { shouldRetryOnError: false, ...options }
  );
}

export function useUser(userId: string, options?: SWRConfiguration) {
  return useSWR<UserProfile>(
    userId ? `/api/users/${userId}` : null,
    jsonFetcher,
    options
  );
}

export function useUserListings(
  userId: string,
  page: number,
  limit: number,
  options?: SWRConfiguration
) {
  const search = new URLSearchParams({ page: String(page), limit: String(limit) });
  return useSWR<ListUserListingsResponse>(
    userId ? `/api/users/${userId}/listings?${search.toString()}` : null,
    jsonFetcher,
    options
  );
}

export function useAnalytics(options?: SWRConfiguration) {
  return useSWR<AnalyticsSummary>("/api/analytics", jsonFetcher, {
    refreshInterval: 60_000,
    ...options,
  });
}

export type AuditParams = {
  action: string;
  page: number;
  limit: number;
};

export function useModerationDecisions(params: AuditParams, options?: SWRConfiguration) {
  const search = new URLSearchParams({
    page: String(params.page),
    limit: String(params.limit),
  });
  if (params.action !== "ALL") {
    search.set("action", params.action);
  }
  return useSWR<ListModerationDecisionsResponse>(
    `/api/audit?${search.toString()}`,
    jsonFetcher,
    { keepPreviousData: true, ...options }
  );
}
