import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  assertPluginJavaScriptSource,
  frontendPluginsFromQuery,
  hostScopeKey,
  instantiatePluginModule,
  loadPluginModule,
  pluginDefinitionFromModule,
  pluginScopeFromMatches,
  reconcileLoadedPlugins,
  resolveHostScope,
  shouldReloadPlugins,
  wantedPluginIdsKey,
} from "./host";
import {
  deactivatePlugin,
  getSlotContributions,
  matchPluginRoute,
  registerRoute,
  registerSlot,
  resetPluginRegistry,
} from "./registry";

function DummySidebar() {
  return null;
}
function DummyReport() {
  return null;
}

describe("plugin host", () => {
  beforeEach(() => {
    resetPluginRegistry();
  });

  it("resolves the deepest navigation scope", () => {
    expect(
      resolveHostScope({
        organizationId: "org-1",
        namespaceId: "ns-1",
        projectId: "proj-1",
      })
    ).toEqual({ id: "proj-1", type: "Project" });
    expect(
      resolveHostScope({ organizationId: "org-1", namespaceId: "ns-1" })
    ).toEqual({ id: "ns-1", type: "Namespace" });
    expect(resolveHostScope({ organizationId: "org-1" })).toEqual({
      id: "org-1",
      type: "Organization",
    });
    expect(resolveHostScope({})).toBeUndefined();
  });

  it("resolves plugin scope from route identity on settings paths", () => {
    expect(
      pluginScopeFromMatches([
        { loaderData: { organization: { id: "org-1" } } },
      ])
    ).toEqual({ id: "org-1", type: "Organization" });
    expect(
      pluginScopeFromMatches([
        {
          loaderData: {
            organization: { id: "org-1" },
            project: { id: "proj-1" },
          },
        },
      ])
    ).toEqual({ id: "proj-1", type: "Project" });
  });

  it("treats equal id and type as the same host scope", () => {
    const first = resolveHostScope({
      organizationId: "org-1",
      namespaceId: "ns-1",
    });
    const second = resolveHostScope({
      organizationId: "org-1",
      namespaceId: "ns-1",
    });
    expect(first).not.toBe(second);
    expect(hostScopeKey(first)).toBe(hostScopeKey(second));
    expect(hostScopeKey(first)).toBe("Namespace:ns-1");
  });

  it("reloads plugins when the host scope identity changes", () => {
    expect(shouldReloadPlugins("", "Organization:org-1")).toBe(false);
    expect(
      shouldReloadPlugins("Organization:org-1", "Organization:org-1")
    ).toBe(false);
    expect(shouldReloadPlugins("Project:proj-1", "Organization:org-1")).toBe(
      true
    );
  });

  it("does not drop contributions when the wanted plugin set is unchanged", () => {
    registerSlot("com.elemo.timetracking", "issue.sidebar", DummySidebar);
    registerRoute("com.elemo.timetracking", "report", DummyReport);

    const loaded = new Map<string, () => void>([
      [
        "com.elemo.timetracking",
        () => deactivatePlugin("com.elemo.timetracking"),
      ],
    ]);
    reconcileLoadedPlugins(loaded, new Set(["com.elemo.timetracking"]));

    expect(loaded.has("com.elemo.timetracking")).toBe(true);
    expect(getSlotContributions("issue.sidebar")).toHaveLength(1);
    expect(matchPluginRoute("com.elemo.timetracking", "report")?.path).toBe(
      "report"
    );
  });

  it("deactivates plugins that left the wanted set", () => {
    registerSlot("com.elemo.timetracking", "issue.sidebar", DummySidebar);
    const loaded = new Map<string, () => void>([
      [
        "com.elemo.timetracking",
        () => deactivatePlugin("com.elemo.timetracking"),
      ],
    ]);
    reconcileLoadedPlugins(loaded, new Set());
    expect(loaded.size).toBe(0);
    expect(getSlotContributions("issue.sidebar")).toHaveLength(0);
  });

  it("reads default or plugin module exports", () => {
    const def = { id: "com.example", activate: () => undefined };
    expect(pluginDefinitionFromModule({ default: def })).toBe(def);
    expect(pluginDefinitionFromModule({ plugin: def })).toBe(def);
    expect(pluginDefinitionFromModule({})).toBeUndefined();
  });

  it("normalizes frontend discovery payloads", () => {
    const plugin = {
      id: "com.elemo.timetracking",
      version: "1.0.0",
      entrypoint: "frontend/index.js",
      slots: [],
    };
    expect(frontendPluginsFromQuery([plugin])).toEqual([plugin]);
    expect(frontendPluginsFromQuery({ data: [plugin] })).toEqual([plugin]);
    expect(frontendPluginsFromQuery({ data: plugin })).toEqual([]);
    expect(frontendPluginsFromQuery(undefined)).toEqual([]);
    expect(wantedPluginIdsKey([plugin])).toBe("com.elemo.timetracking");
  });

  it("rejects HTML shells as plugin modules", () => {
    expect(() =>
      assertPluginJavaScriptSource(
        "<!DOCTYPE html><html></html>",
        "text/html",
        "com.elemo.timetracking"
      )
    ).toThrow(/is not JavaScript/);
  });

  it("imports rewritten ESM without revoking the blob URL", async () => {
    const activate = vi.fn();
    const create = vi
      .spyOn(URL, "createObjectURL")
      .mockReturnValue("blob:plugin-module");
    const revoke = vi.spyOn(URL, "revokeObjectURL").mockImplementation(() => {
      return undefined;
    });
    const importer = vi.fn(() =>
      Promise.resolve({
        default: { id: "com.example", activate },
      })
    );
    const definition = await instantiatePluginModule(
      'export default { id: "com.example" };\n',
      importer
    );
    expect(definition?.activate).toBe(activate);
    expect(importer).toHaveBeenCalledWith("blob:plugin-module");
    expect(revoke).not.toHaveBeenCalled();
    create.mockRestore();
    revoke.mockRestore();
  });

  it("loads plugin source through the injected fetcher and rejects failures", async () => {
    await expect(
      loadPluginModule(
        {
          id: "com.example.plugin",
          version: "1.0.0",
          entrypoint: "frontend/index.js",
        },
        () =>
          Promise.resolve({
            status: 200,
            contentType: "text/html",
            source: "<!DOCTYPE html>",
          })
      )
    ).rejects.toThrow(/is not JavaScript/);

    await expect(
      loadPluginModule(
        {
          id: "com.example.plugin",
          version: "1.0.0",
          entrypoint: "frontend/index.js",
        },
        () => Promise.resolve({ status: 404, contentType: "", source: "" })
      )
    ).rejects.toThrow(/returned 404/);
  });
});
