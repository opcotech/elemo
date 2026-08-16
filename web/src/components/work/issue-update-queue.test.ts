import { describe, expect, it } from "vitest";

import { enqueueIssueUpdate } from "./issue-update-queue";

async function flushMicrotasks() {
  await Promise.resolve();
  await Promise.resolve();
}

describe("enqueueIssueUpdate", () => {
  it("runs tasks for the same issue in order even when started together", async () => {
    const events: string[] = [];
    let releaseFirst!: () => void;
    const firstGate = new Promise<void>((resolve) => {
      releaseFirst = resolve;
    });

    const first = enqueueIssueUpdate("issue-1", async () => {
      events.push("first-start");
      await firstGate;
      events.push("first-end");
      return "first";
    });
    const second = enqueueIssueUpdate("issue-1", () => {
      events.push("second-start");
      events.push("second-end");
      return Promise.resolve("second");
    });

    await flushMicrotasks();
    expect(events).toEqual(["first-start"]);
    releaseFirst();

    await expect(first).resolves.toBe("first");
    await expect(second).resolves.toBe("second");
    expect(events).toEqual([
      "first-start",
      "first-end",
      "second-start",
      "second-end",
    ]);
  });

  it("still runs the next task when the previous one fails", async () => {
    const first = enqueueIssueUpdate("issue-2", () =>
      Promise.reject(new Error("move failed"))
    );
    const firstFailed = expect(first).rejects.toThrow("move failed");
    const second = enqueueIssueUpdate("issue-2", () =>
      Promise.resolve("dates")
    );

    await firstFailed;
    await expect(second).resolves.toBe("dates");
  });

  it("does not block updates to a different issue", async () => {
    const events: string[] = [];
    let releaseFirst!: () => void;
    const firstGate = new Promise<void>((resolve) => {
      releaseFirst = resolve;
    });

    const first = enqueueIssueUpdate("issue-a", async () => {
      events.push("a-start");
      await firstGate;
      events.push("a-end");
    });
    const second = enqueueIssueUpdate("issue-b", () => {
      events.push("b");
      return Promise.resolve();
    });

    await second;
    expect(events).toEqual(["a-start", "b"]);
    releaseFirst();
    await first;
    expect(events).toEqual(["a-start", "b", "a-end"]);
  });
});
