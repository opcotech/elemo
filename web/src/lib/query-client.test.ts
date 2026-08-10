import { dehydrate, hydrate } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import { ApiError } from "./api/errors";
import { createQueryClient } from "./query-client";

describe("query client hydration", () => {
  it("hydrates successful data without duplicating the initial request", async () => {
    const serverClient = createQueryClient();
    serverClient.setQueryData(["organizations"], [{ id: "organization-1" }]);

    const dehydrated = dehydrate(serverClient);
    const browserClient = createQueryClient();
    hydrate(browserClient, dehydrated);

    const queryFn = vi.fn(() =>
      Promise.resolve([{ id: "unexpected-request" }])
    );
    const data = await browserClient.fetchQuery({
      queryKey: ["organizations"],
      queryFn,
    });

    expect(data).toEqual([{ id: "organization-1" }]);
    expect(queryFn).not.toHaveBeenCalled();
  });

  it("clears all user-scoped state on account changes", () => {
    const queryClient = createQueryClient();
    queryClient.setQueryData(["session-user", "user-1"], { id: "user-1" });
    queryClient.setQueryData(["organizations"], [{ id: "organization-1" }]);

    queryClient.clear();

    expect(queryClient.getQueryCache().getAll()).toHaveLength(0);
  });

  it("retries only bounded transient query failures", () => {
    const queryClient = createQueryClient();
    const retry = queryClient.getDefaultOptions().queries?.retry;

    expect(retry).toBeTypeOf("function");
    const shouldRetry = retry as (
      failureCount: number,
      error: unknown
    ) => boolean;

    expect(shouldRetry(0, new ApiError(500, "server error"))).toBe(true);
    expect(shouldRetry(1, new ApiError(429, "rate limited"))).toBe(true);
    expect(shouldRetry(2, new ApiError(500, "server error"))).toBe(false);
    expect(shouldRetry(0, new ApiError(400, "bad request"))).toBe(false);
    expect(shouldRetry(0, new Error("network failure"))).toBe(true);
    expect(shouldRetry(1, new Error("network failure"))).toBe(false);
  });

  it("does not retry mutations and disables focus refetches", () => {
    const defaults = createQueryClient().getDefaultOptions();

    expect(defaults.mutations?.retry).toBe(false);
    expect(defaults.queries?.refetchOnWindowFocus).toBe(false);
    expect(defaults.queries?.refetchOnReconnect).toBe(true);
  });
});
