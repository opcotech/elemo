import type {
  ElemoPluginAPI,
  PluginIssue,
  PluginScope,
  PluginUser,
} from "@elemo/plugin-sdk";

import { throwIfApiFailed } from "@/lib/api/errors";
import {
  v1IssueGet,
  v1PluginGraphNodeDelete,
  v1PluginGraphNodeGet,
  v1PluginGraphNodeMove,
  v1PluginGraphNodeUpdate,
  v1PluginGraphNodesCreate,
  v1PluginGraphNodesGet,
  v1PluginGraphRelationDelete,
  v1PluginGraphRelationsCreate,
  v1PluginGraphRelationsGet,
  v1PluginInvoke,
  v1ProjectsIssuesGet,
  v1UserGet,
} from "@/lib/api/sdk";
import type { ResourceType } from "@/lib/api/types";

function asResourceType(value: string): ResourceType {
  return value as ResourceType;
}

function compositeScopeId(scope: PluginScope): string {
  if (scope.id.includes(":")) {
    return scope.id;
  }
  return `${scope.type}:${scope.id}`;
}

export function mapIssue(issue: {
  id: string;
  key: string;
  title: string;
  project?: { id?: string } | null;
  namespace?: { slug?: string } | null;
}): PluginIssue {
  return {
    id: issue.id,
    key: issue.key,
    title: issue.title,
    projectId: issue.project?.id,
    namespaceSlug: issue.namespace?.slug,
  };
}

function mapUser(user: {
  id: string;
  first_name: string;
  last_name: string;
  picture?: string | null;
}): PluginUser {
  return {
    id: user.id,
    first_name: user.first_name,
    last_name: user.last_name,
    picture: user.picture ?? null,
  };
}

export function createPluginClientAPI(
  pluginId: string,
  scope: PluginScope,
  context: Record<string, unknown> = {}
): Omit<ElemoPluginAPI, "slots" | "routes" | "pluginId"> {
  return {
    scope,
    context,
    api: {
      invoke: async (functionName, payload) => {
        const result = throwIfApiFailed(
          await v1PluginInvoke({
            path: { pluginId },
            body: {
              function: functionName,
              scope_id: compositeScopeId(scope),
              payload,
            },
          })
        );
        if (!result.ok) {
          const message = result.error || "plugin invoke failed";
          console.error(`Plugin ${pluginId} ${functionName} failed`, message);
          throw new Error(message);
        }
        return result.data;
      },
      graph: {
        nodes: {
          list: async (opts) => {
            const items = [];
            let pageToken: string | undefined;
            do {
              const page = throwIfApiFailed(
                await v1PluginGraphNodesGet({
                  path: { pluginId },
                  query: {
                    kind: opts.kind,
                    scope_id: opts.scopeId,
                    scope_type: asResourceType(opts.scopeType),
                    equals: opts.equals
                      ? JSON.stringify(opts.equals)
                      : undefined,
                    owner_plugin_id: opts.ownerPluginId,
                    page_size: 100,
                    page_token: pageToken,
                  },
                })
              );
              items.push(...page.items);
              const next = page.page_info.next_page_token;
              pageToken = next ? next : undefined;
            } while (pageToken);
            return items;
          },
          get: async (id, opts) =>
            throwIfApiFailed(
              await v1PluginGraphNodeGet({
                path: { pluginId, id },
                query: opts?.ownerPluginId
                  ? { owner_plugin_id: opts.ownerPluginId }
                  : undefined,
              })
            ),
          create: async (opts) =>
            throwIfApiFailed(
              await v1PluginGraphNodesCreate({
                path: { pluginId },
                body: {
                  kind: opts.kind,
                  parent_id: opts.parentId,
                  parent_type: asResourceType(opts.parentType),
                  properties: opts.properties,
                },
              })
            ),
          update: async (id, properties) =>
            throwIfApiFailed(
              await v1PluginGraphNodeUpdate({
                path: { pluginId, id },
                body: { properties },
              })
            ),
          move: async (id, parent) =>
            throwIfApiFailed(
              await v1PluginGraphNodeMove({
                path: { pluginId, id },
                body: {
                  parent_id: parent.parentId,
                  parent_type: asResourceType(parent.parentType),
                },
              })
            ),
          delete: async (id) => {
            throwIfApiFailed(
              await v1PluginGraphNodeDelete({
                path: { pluginId, id },
              })
            );
          },
        },
        relations: {
          list: async (opts) => {
            const page = throwIfApiFailed(
              await v1PluginGraphRelationsGet({
                path: { pluginId },
                query: {
                  kind: opts.kind,
                  node_id: opts.nodeId,
                  node_type: asResourceType(opts.nodeType),
                  direction: opts.direction,
                },
              })
            );
            return page.items;
          },
          create: async (opts) =>
            throwIfApiFailed(
              await v1PluginGraphRelationsCreate({
                path: { pluginId },
                body: {
                  kind: opts.kind,
                  from_id: opts.fromId,
                  from_type: asResourceType(opts.fromType),
                  to_id: opts.toId,
                  to_type: asResourceType(opts.toType),
                },
              })
            ),
          delete: async (id) => {
            throwIfApiFailed(
              await v1PluginGraphRelationDelete({
                path: { pluginId, id },
              })
            );
          },
        },
      },
      issues: {
        get: async (id) => {
          const issue = throwIfApiFailed(await v1IssueGet({ path: { id } }));
          return mapIssue(issue);
        },
        list: async (opts) => {
          const page = throwIfApiFailed(
            await v1ProjectsIssuesGet({
              path: { projectId: opts.projectId },
              query: { page_size: 100 },
            })
          );
          return page.items.map(mapIssue);
        },
      },
      users: {
        get: async (id) => {
          const user = throwIfApiFailed(await v1UserGet({ path: { id } }));
          return mapUser(user);
        },
      },
    },
  };
}
