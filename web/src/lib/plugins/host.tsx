import type { ElemoPluginDefinition } from "@elemo/plugin-sdk";
import { useQuery } from "@tanstack/react-query";
import { useRouterState } from "@tanstack/react-router";
import { useEffect, useMemo, useRef } from "react";

import { createPluginClientAPI } from "./api";
import { isPluginJavaScriptSource } from "./asset-path";
import {
  createPluginHostAPI,
  deactivatePlugin,
  isFrontendDiscoverySettled,
  markFrontendDiscoverySettled,
  markPluginSettled,
} from "./registry";
import { bindPluginRuntimeImports, ensurePluginRuntime } from "./runtime";

import { fetchPluginFrontendSourceFn } from "@/lib/api/plugin-assets";
import { v1PluginsFrontendGetOptions } from "@/lib/api/query-options";
import type { FrontendPlugin } from "@/lib/api/types";
import { cacheProfiles } from "@/lib/query-client";
import { identityFromMatches } from "@/lib/route-identity";

export interface HostScope {
  id: string;
  type: string;
}

export interface PluginModuleSpec {
  id: string;
  version: string;
  entrypoint: string;
}

export type PluginSourceFetcher = (
  plugin: PluginModuleSpec
) => Promise<{ status: number; contentType: string; source: string }>;

function isFrontendPlugin(value: unknown): value is FrontendPlugin {
  if (!value || typeof value !== "object") {
    return false;
  }
  const record = value as Record<string, unknown>;
  return (
    typeof record.id === "string" &&
    record.id.length > 0 &&
    typeof record.version === "string" &&
    typeof record.entrypoint === "string"
  );
}

export function frontendPluginsFromQuery(data: unknown): FrontendPlugin[] {
  if (Array.isArray(data)) {
    return data.filter(isFrontendPlugin);
  }
  if (
    data &&
    typeof data === "object" &&
    "data" in data &&
    Array.isArray((data as { data: unknown }).data)
  ) {
    return (data as { data: unknown[] }).data.filter(isFrontendPlugin);
  }
  return [];
}

export function wantedPluginIdsKey(plugins: readonly FrontendPlugin[]): string {
  return plugins
    .map((plugin) => plugin.id)
    .sort()
    .join("\0");
}

export function pluginDefinitionFromModule(mod: {
  default?: ElemoPluginDefinition;
  plugin?: ElemoPluginDefinition;
}): ElemoPluginDefinition | undefined {
  return mod.default ?? mod.plugin;
}

export function assertPluginJavaScriptSource(
  source: string,
  contentType: string,
  label: string
): void {
  if (!isPluginJavaScriptSource(source, contentType)) {
    throw new Error(
      `Plugin module ${label} is not JavaScript (${contentType})`
    );
  }
}

async function defaultFetchPluginSource(
  plugin: PluginModuleSpec
): Promise<{ status: number; contentType: string; source: string }> {
  const result = await fetchPluginFrontendSourceFn({
    data: {
      pluginId: plugin.id,
      version: plugin.version,
      entrypoint: plugin.entrypoint,
    },
  });
  if (!result) {
    throw new Error(`Plugin module ${plugin.id} returned no source`);
  }
  return result;
}

export async function instantiatePluginModule(
  source: string,
  importer: (url: string) => Promise<unknown> = (url) =>
    import(/* @vite-ignore */ url)
): Promise<ElemoPluginDefinition | undefined> {
  ensurePluginRuntime();
  const bound = bindPluginRuntimeImports(source);
  const blob = new Blob([bound], { type: "text/javascript" });
  const blobUrl = URL.createObjectURL(blob);
  const mod = (await importer(blobUrl)) as {
    default?: ElemoPluginDefinition;
    plugin?: ElemoPluginDefinition;
  };
  return pluginDefinitionFromModule(mod);
}

export async function loadPluginModule(
  plugin: PluginModuleSpec,
  fetchSource: PluginSourceFetcher = defaultFetchPluginSource
): Promise<ElemoPluginDefinition | undefined> {
  ensurePluginRuntime();
  const result = await fetchSource(plugin);
  if (result.status !== 200) {
    throw new Error(`Plugin module ${plugin.id} returned ${result.status}`);
  }
  assertPluginJavaScriptSource(result.source, result.contentType, plugin.id);
  return instantiatePluginModule(result.source);
}

export function resolveHostScope(navigation: {
  organizationId?: string;
  namespaceId?: string;
  projectId?: string;
}): HostScope | undefined {
  if (navigation.projectId) {
    return { id: navigation.projectId, type: "Project" };
  }
  if (navigation.namespaceId) {
    return { id: navigation.namespaceId, type: "Namespace" };
  }
  if (navigation.organizationId) {
    return { id: navigation.organizationId, type: "Organization" };
  }
  return undefined;
}

export function pluginScopeFromMatches(
  matches: readonly { loaderData?: unknown }[]
): HostScope | undefined {
  return resolveHostScope(identityFromMatches(matches));
}

export function hostScopeKey(scope: HostScope | undefined): string {
  return scope ? `${scope.type}:${scope.id}` : "";
}

export function shouldReloadPlugins(
  loadedScopeKey: string,
  nextScopeKey: string
): boolean {
  return loadedScopeKey !== "" && loadedScopeKey !== nextScopeKey;
}

export function reconcileLoadedPlugins(
  loaded: Map<string, () => void>,
  wantedIds: Set<string>
): void {
  for (const [id, dispose] of loaded) {
    if (!wantedIds.has(id)) {
      dispose();
      deactivatePlugin(id);
      loaded.delete(id);
    }
  }
}

export function PluginHost() {
  const matches = useRouterState({
    select: (state) => state.matches,
  });
  const identity = useMemo(() => identityFromMatches(matches), [matches]);
  const scope = useMemo(() => resolveHostScope(identity), [identity]);
  const scopeKey = hostScopeKey(scope);
  const active = useRef(new Map<string, () => void>());
  const loadedScopeKey = useRef("");

  const query = useQuery({
    ...v1PluginsFrontendGetOptions({
      query: {
        scope_id: scope?.id ?? "",
        scope_type: (scope?.type ?? "Organization") as never,
      },
    }),
    enabled: Boolean(scope),
    ...cacheProfiles.volatile,
    refetchOnWindowFocus: true,
  });

  const plugins = useMemo(() => {
    if (query.status !== "success") {
      return [];
    }
    return frontendPluginsFromQuery(query.data);
  }, [query.data, query.status]);
  const wantedIds = useMemo(() => wantedPluginIdsKey(plugins), [plugins]);
  const pluginsRef = useRef<FrontendPlugin[]>(plugins);
  pluginsRef.current = plugins;

  useEffect(() => {
    if (!scope || query.status === "pending") {
      return;
    }
    if (query.status === "error") {
      if (!isFrontendDiscoverySettled()) {
        markFrontendDiscoverySettled([]);
      }
      return;
    }
    ensurePluginRuntime();
    if (shouldReloadPlugins(loadedScopeKey.current, scopeKey)) {
      for (const dispose of active.current.values()) {
        dispose();
      }
      active.current.clear();
    }
    loadedScopeKey.current = scopeKey;
    const discovered = pluginsRef.current;
    markFrontendDiscoverySettled(discovered.map((plugin) => plugin.id));
    const wanted = new Set(discovered.map((plugin) => plugin.id));
    reconcileLoadedPlugins(active.current, wanted);

    let cancelled = false;
    void (async () => {
      for (const plugin of discovered) {
        if (cancelled) {
          continue;
        }
        if (active.current.has(plugin.id)) {
          markPluginSettled(plugin.id);
          continue;
        }
        try {
          const definition = await loadPluginModule(plugin);
          if (!definition) {
            console.error(
              `Plugin ${plugin.id} has no default or plugin export`
            );
            markPluginSettled(plugin.id);
            continue;
          }
          if (cancelled || active.current.has(plugin.id)) {
            markPluginSettled(plugin.id);
            continue;
          }
          const api = createPluginHostAPI(
            plugin.id,
            createPluginClientAPI(plugin.id, scope)
          );
          const dispose = definition.activate(api);
          active.current.set(plugin.id, () => {
            dispose?.();
            definition.deactivate?.(api);
            deactivatePlugin(plugin.id);
          });
          markPluginSettled(plugin.id);
        } catch (error) {
          console.error(`Failed to load plugin ${plugin.id}`, error);
          markPluginSettled(plugin.id);
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [query.status, scope, scopeKey, wantedIds]);

  useEffect(() => {
    const live = active.current;
    return () => {
      for (const dispose of live.values()) {
        dispose();
      }
      live.clear();
    };
  }, []);

  return null;
}
