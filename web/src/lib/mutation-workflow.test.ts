import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import {
  rollbackOptimisticQueryData,
  runMutationSuccessWorkflow,
} from "./mutation-workflow";

describe("mutation success workflow", () => {
  it("awaits cache, router, callback, and navigation phases in order", async () => {
    const events: string[] = [];
    const step = (name: string) => async () => {
      await Promise.resolve();
      events.push(name);
    };

    await runMutationSuccessWorkflow({
      invalidateQueries: [step("query-1"), step("query-2")],
      invalidateRouter: step("router"),
      callbacks: [step("callback-1"), step("callback-2")],
      navigate: step("navigate"),
    });

    expect(events).toEqual([
      "query-1",
      "query-2",
      "router",
      "callback-1",
      "callback-2",
      "navigate",
    ]);
  });

  it("waits for every cache invalidation before invalidating the router", async () => {
    const events: string[] = [];
    let releaseFirst!: () => void;
    const firstInvalidation = new Promise<void>((resolve) => {
      releaseFirst = resolve;
    });

    const workflow = runMutationSuccessWorkflow({
      invalidateQueries: [
        async () => {
          events.push("query-1-started");
          await firstInvalidation;
          events.push("query-1-finished");
        },
        () => {
          events.push("query-2-finished");
        },
      ],
      invalidateRouter: () => {
        events.push("router");
      },
    });

    await vi.waitFor(() => {
      expect(events).toEqual(["query-1-started", "query-2-finished"]);
    });
    releaseFirst();
    await workflow;

    expect(events).toEqual([
      "query-1-started",
      "query-2-finished",
      "query-1-finished",
      "router",
    ]);
  });

  it("stops later phases when cache invalidation fails", async () => {
    const invalidateRouter = vi.fn();
    const callback = vi.fn();
    const navigate = vi.fn();

    await expect(
      runMutationSuccessWorkflow({
        invalidateQueries: [
          () => Promise.reject(new Error("cache refresh failed")),
        ],
        invalidateRouter,
        callbacks: [callback],
        navigate,
      })
    ).rejects.toThrow("cache refresh failed");

    expect(invalidateRouter).not.toHaveBeenCalled();
    expect(callback).not.toHaveBeenCalled();
    expect(navigate).not.toHaveBeenCalled();
  });
});

describe("optimistic query rollback", () => {
  it("restores the previous cached value", () => {
    const queryClient = new QueryClient();
    const queryKey = ["todos"];
    const previous = [{ id: "todo-1", completed: false }];

    queryClient.setQueryData(queryKey, previous);
    queryClient.setQueryData(queryKey, [{ id: "todo-1", completed: true }]);

    rollbackOptimisticQueryData(queryClient, queryKey, { previous });

    expect(queryClient.getQueryData(queryKey)).toStrictEqual(previous);
  });

  it("removes optimistic data when no cached value existed", () => {
    const queryClient = new QueryClient();
    const queryKey = ["todos"];

    queryClient.setQueryData(queryKey, []);
    rollbackOptimisticQueryData(queryClient, queryKey, {
      previous: undefined,
    });

    expect(queryClient.getQueryState(queryKey)).toBeUndefined();
  });

  it("leaves cached data unchanged without mutation context", () => {
    const queryClient = new QueryClient();
    const queryKey = ["todos"];
    const current = [{ id: "todo-1", completed: true }];
    queryClient.setQueryData(queryKey, current);

    rollbackOptimisticQueryData(queryClient, queryKey, undefined);

    expect(queryClient.getQueryData(queryKey)).toBe(current);
  });
});
