import { z } from "zod";

import type { Scope } from "@/lib/work/model";

export const workLayoutSchema = z.enum(["board", "list", "table", "timeline"]);

export const workScopeSchema = z
  .union([
    z.literal("global"),
    z.string().regex(/^(namespace|project|person):[a-zA-Z0-9_-]+$/),
  ])
  .optional();

export const workRouteSearchSchema = z.object({
  view: z.string().max(100).optional(),
  scope: workScopeSchema,
  filter: z.string().max(500).optional(),
  group: z.enum(["status", "priority", "assignee", "none"]).catch("status"),
  sort: z.string().max(160).catch("rank:asc"),
  display: z.enum(["comfortable", "compact"]).catch("comfortable"),
  layout: workLayoutSchema.catch("board"),
  selected: z
    .string()
    .regex(/^work:[a-zA-Z0-9_-]+$/)
    .optional(),
});

export type WorkRouteSearch = z.infer<typeof workRouteSearchSchema>;

export function serializeWorkScope(scope: Scope): string {
  if (scope.type === "global") return "global";
  if (scope.type === "namespace") return `namespace:${scope.namespaceId}`;
  if (scope.type === "project") return `project:${scope.projectId}`;
  return `person:${scope.personId}`;
}

export function workScopeOptions(baseScope: Scope): Scope[] {
  if (baseScope.type === "project") {
    return [
      baseScope,
      ...(baseScope.namespaceId
        ? [
            {
              type: "namespace",
              namespaceId: baseScope.namespaceId,
            } as const,
          ]
        : []),
      { type: "global" },
    ];
  }
  if (baseScope.type === "namespace" || baseScope.type === "person") {
    return [baseScope, { type: "global" }];
  }
  return [baseScope];
}

export function resolveWorkScope(
  value: WorkRouteSearch["scope"],
  baseScope: Scope
): Scope {
  return (
    workScopeOptions(baseScope).find(
      (scope) => serializeWorkScope(scope) === value
    ) ?? baseScope
  );
}
