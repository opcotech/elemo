import { beforeEach, describe, expect, it } from "vitest";

import {
  deactivatePlugin,
  getSlotContributions,
  isFrontendDiscoverySettled,
  isPluginPending,
  markFrontendDiscoverySettled,
  markPluginSettled,
  matchPluginRoute,
  pluginPageState,
  registerRoute,
  registerSlot,
  resetPluginRegistry,
} from "./registry";

function DummyA() {
  return null;
}
function DummyB() {
  return null;
}
function DummyC() {
  return null;
}

describe("plugin registry", () => {
  beforeEach(() => {
    resetPluginRegistry();
  });

  it("orders slot contributions by order then plugin id", () => {
    registerSlot("com.z", "issue.sidebar", DummyA, { order: 1 });
    registerSlot("com.a", "issue.sidebar", DummyB, { order: 1 });
    registerSlot("com.m", "issue.sidebar", DummyC, { order: 0 });

    const ids = getSlotContributions("issue.sidebar").map(
      (item) => item.pluginId
    );
    expect(ids).toEqual(["com.m", "com.a", "com.z"]);
  });

  it("isolates contributions by slot name", () => {
    registerSlot("com.a", "issue.sidebar", DummyA);
    registerSlot("com.a", "issue.actions", DummyB);
    expect(getSlotContributions("issue.sidebar")).toHaveLength(1);
    expect(getSlotContributions("issue.actions")).toHaveLength(1);
  });

  it("stores an optional title on slot contributions", () => {
    registerSlot("com.elemo.timetracking", "issue.activity", DummyA, {
      title: "Logged time",
    });
    expect(getSlotContributions("issue.activity")[0]?.title).toBe(
      "Logged time"
    );
  });

  it("returns a stable snapshot until the registry changes", () => {
    const emptyFirst = getSlotContributions("issue.sidebar");
    const emptySecond = getSlotContributions("issue.sidebar");
    expect(emptySecond).toBe(emptyFirst);

    registerSlot("com.a", "issue.sidebar", DummyA);
    const first = getSlotContributions("issue.sidebar");
    const second = getSlotContributions("issue.sidebar");
    expect(second).toBe(first);
    expect(first).toHaveLength(1);

    registerSlot("com.b", "issue.sidebar", DummyB);
    const third = getSlotContributions("issue.sidebar");
    expect(third).not.toBe(first);
    expect(third.map((item) => item.pluginId)).toEqual(["com.a", "com.b"]);
  });

  it("removes contributions on deactivate", () => {
    registerSlot("com.a", "issue.sidebar", DummyA);
    registerSlot("com.b", "issue.sidebar", DummyB);
    registerRoute("com.a", "report", DummyC);
    deactivatePlugin("com.a");
    expect(
      getSlotContributions("issue.sidebar").map((item) => item.pluginId)
    ).toEqual(["com.b"]);
    expect(matchPluginRoute("com.a", "report")).toBeUndefined();
  });

  it("dispatches splat paths to the registered plugin route", () => {
    registerRoute("com.elemo.timetracking", "report", DummyA);
    expect(matchPluginRoute("com.elemo.timetracking", "report")?.path).toBe(
      "report"
    );
    expect(
      matchPluginRoute("com.elemo.timetracking", "report/extra")?.path
    ).toBe("report");
    expect(matchPluginRoute("other", "report")).toBeUndefined();
  });

  it("reports loading until discovery settles and the plugin finishes loading", () => {
    expect(pluginPageState("com.elemo.timetracking", "report")).toBe("loading");
    expect(isFrontendDiscoverySettled()).toBe(false);

    markFrontendDiscoverySettled(["com.elemo.timetracking"]);
    expect(isPluginPending("com.elemo.timetracking")).toBe(true);
    expect(pluginPageState("com.elemo.timetracking", "report")).toBe("loading");

    registerRoute("com.elemo.timetracking", "report", DummyA);
    expect(pluginPageState("com.elemo.timetracking", "report")).toBe("ready");

    markPluginSettled("com.elemo.timetracking");
    expect(isPluginPending("com.elemo.timetracking")).toBe(false);
    expect(pluginPageState("com.elemo.timetracking", "report")).toBe("ready");
  });

  it("reports missing after discovery when the plugin has no matching route", () => {
    markFrontendDiscoverySettled(["com.elemo.timetracking"]);
    markPluginSettled("com.elemo.timetracking");
    expect(pluginPageState("com.elemo.timetracking", "report")).toBe("missing");
    expect(pluginPageState("com.other", "report")).toBe("missing");
  });
});
