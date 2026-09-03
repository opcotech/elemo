import type { ElemoPluginAPI, PluginSlotName } from "@elemo/plugin-sdk";
import type { ComponentType } from "react";

export interface SlotContribution {
  pluginId: string;
  slot: PluginSlotName;
  component: ComponentType<Record<string, unknown>>;
  order: number;
  title?: string;
  index: number;
}

export interface RouteContribution {
  pluginId: string;
  path: string;
  component: ComponentType<Record<string, unknown>>;
}

type Listener = () => void;

const slots: SlotContribution[] = [];
const routes: RouteContribution[] = [];
const listeners = new Set<Listener>();
const slotSnapshots = new Map<PluginSlotName, SlotContribution[]>();
const EMPTY_SLOT_SNAPSHOT: SlotContribution[] = [];
const pendingPlugins = new Set<string>();
const knownPlugins = new Set<string>();
let frontendDiscoverySettled = false;
let nextIndex = 0;

function emit() {
  slotSnapshots.clear();
  for (const listener of listeners) {
    listener();
  }
}

export function subscribePluginRegistry(listener: Listener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function registerSlot(
  pluginId: string,
  slot: PluginSlotName,
  component: ComponentType<Record<string, unknown>>,
  options: { order?: number; title?: string } = {}
): () => void {
  const contribution: SlotContribution = {
    pluginId,
    slot,
    component,
    order: options.order ?? 0,
    title: options.title,
    index: nextIndex++,
  };
  slots.push(contribution);
  emit();
  return () => {
    const i = slots.indexOf(contribution);
    if (i >= 0) {
      slots.splice(i, 1);
      emit();
    }
  };
}

export function registerRoute(
  pluginId: string,
  path: string,
  component: ComponentType<Record<string, unknown>>
): () => void {
  const contribution: RouteContribution = {
    pluginId,
    path: path.replace(/^\/+/, ""),
    component,
  };
  routes.push(contribution);
  emit();
  return () => {
    const i = routes.indexOf(contribution);
    if (i >= 0) {
      routes.splice(i, 1);
      emit();
    }
  };
}

export function getSlotContributions(slot: PluginSlotName): SlotContribution[] {
  const cached = slotSnapshots.get(slot);
  if (cached !== undefined) {
    return cached;
  }
  const result = slots
    .filter((item) => item.slot === slot)
    .sort((a, b) => {
      if (a.order !== b.order) {
        return a.order - b.order;
      }
      const byPlugin = a.pluginId.localeCompare(b.pluginId);
      if (byPlugin !== 0) {
        return byPlugin;
      }
      return a.index - b.index;
    });
  // useSyncExternalStore requires a stable snapshot when data is unchanged.
  const snapshot = result.length === 0 ? EMPTY_SLOT_SNAPSHOT : result;
  slotSnapshots.set(slot, snapshot);
  return snapshot;
}

export function matchPluginRoute(
  pluginId: string,
  splat: string
): RouteContribution | undefined {
  const path = splat.replace(/^\/+|\/+$/g, "");
  const matches = routes.filter((item) => item.pluginId === pluginId);
  return (
    matches.find((item) => item.path === path) ??
    matches.find(
      (item) => path === item.path || path.startsWith(`${item.path}/`)
    )
  );
}

export function markFrontendDiscoverySettled(pluginIds: string[]): void {
  frontendDiscoverySettled = true;
  const wanted = new Set(pluginIds);
  for (const id of pluginIds) {
    knownPlugins.add(id);
    pendingPlugins.add(id);
  }
  for (const id of [...pendingPlugins]) {
    if (!wanted.has(id)) {
      pendingPlugins.delete(id);
    }
  }
  emit();
}

export function markPluginSettled(pluginId: string): void {
  pendingPlugins.delete(pluginId);
  knownPlugins.add(pluginId);
  emit();
}

export function isPluginPending(pluginId: string): boolean {
  return pendingPlugins.has(pluginId);
}

export function isFrontendDiscoverySettled(): boolean {
  return frontendDiscoverySettled;
}

export type PluginPageState = "loading" | "missing" | "ready";

export function pluginPageState(
  pluginId: string,
  splat: string
): PluginPageState {
  if (matchPluginRoute(pluginId, splat)) {
    return "ready";
  }
  if (!frontendDiscoverySettled || pendingPlugins.has(pluginId)) {
    return "loading";
  }
  return "missing";
}

export function deactivatePlugin(pluginId: string) {
  for (let i = slots.length - 1; i >= 0; i--) {
    if (slots[i].pluginId === pluginId) {
      slots.splice(i, 1);
    }
  }
  for (let i = routes.length - 1; i >= 0; i--) {
    if (routes[i].pluginId === pluginId) {
      routes.splice(i, 1);
    }
  }
  emit();
}

export function resetPluginRegistry() {
  slots.length = 0;
  routes.length = 0;
  pendingPlugins.clear();
  knownPlugins.clear();
  frontendDiscoverySettled = false;
  nextIndex = 0;
  emit();
}

export function createPluginHostAPI(
  pluginId: string,
  api: Omit<ElemoPluginAPI, "slots" | "routes" | "pluginId">
): ElemoPluginAPI {
  return {
    pluginId,
    ...api,
    slots: {
      register: (slot, component, options) =>
        registerSlot(pluginId, slot, component, options),
    },
    routes: {
      register: (path, component) => registerRoute(pluginId, path, component),
    },
  };
}
