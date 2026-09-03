import type { ComponentType } from "react";

export type PluginSlotName =
  | "issue.sidebar"
  | "issue.actions"
  | "issue.activity"
  | "organization.settings"
  | "project.settings"
  | "project.sidebar";

export interface PluginScope {
  id: string;
  type: string;
}

export interface PluginGraphNode {
  id: string;
  plugin_id: string;
  kind: string;
  properties: Record<string, unknown>;
  parent_id?: string | null;
  parent_type?: string | null;
  created_at?: string | null;
  updated_at?: string | null;
}

export interface PluginGraphRelation {
  id: string;
  kind: string;
  from: string;
  from_type: string;
  to: string;
  to_type: string;
  created_at?: string | null;
}

export interface PluginIssue {
  id: string;
  key: string;
  title: string;
  projectId?: string;
  namespaceSlug?: string;
}

export interface PluginUser {
  id: string;
  first_name: string;
  last_name: string;
  picture?: string | null;
}

export interface PluginGraphAPI {
  nodes: {
    list: (opts: {
      kind: string;
      scopeId: string;
      scopeType: string;
      equals?: Record<string, unknown>;
      ownerPluginId?: string;
    }) => Promise<Array<PluginGraphNode>>;
    get: (
      id: string,
      opts?: { ownerPluginId?: string }
    ) => Promise<PluginGraphNode>;
    create: (opts: {
      kind: string;
      parentId: string;
      parentType: string;
      properties?: Record<string, unknown>;
    }) => Promise<PluginGraphNode>;
    update: (
      id: string,
      properties: Record<string, unknown>
    ) => Promise<PluginGraphNode>;
    move: (
      id: string,
      parent: { parentId: string; parentType: string }
    ) => Promise<PluginGraphNode>;
    delete: (id: string) => Promise<void>;
  };
  relations: {
    list: (opts: {
      kind: string;
      nodeId: string;
      nodeType: string;
      direction?: "outgoing" | "incoming" | "both";
    }) => Promise<Array<PluginGraphRelation>>;
    create: (opts: {
      kind: string;
      fromId: string;
      fromType: string;
      toId: string;
      toType: string;
    }) => Promise<PluginGraphRelation>;
    delete: (id: string) => Promise<void>;
  };
}

export interface ElemoPluginAPI {
  pluginId: string;
  scope: PluginScope;
  context: Record<string, unknown>;
  api: {
    invoke: (functionName: string, payload?: unknown) => Promise<unknown>;
    graph: PluginGraphAPI;
    issues: {
      get: (id: string) => Promise<PluginIssue>;
      list: (opts: { projectId: string }) => Promise<Array<PluginIssue>>;
    };
    users: {
      get: (id: string) => Promise<PluginUser>;
    };
  };
  slots: {
    register: (
      slot: PluginSlotName,
      component: ComponentType<Record<string, unknown>>,
      options?: { order?: number; title?: string }
    ) => () => void;
  };
  routes: {
    register: (
      path: string,
      component: ComponentType<Record<string, unknown>>
    ) => () => void;
  };
}

export interface ElemoPluginDefinition {
  id: string;
  activate: (elemo: ElemoPluginAPI) => void | (() => void);
  deactivate?: (elemo: ElemoPluginAPI) => void;
}

export function defineElemoPlugin(
  plugin: ElemoPluginDefinition
): ElemoPluginDefinition {
  return plugin;
}
