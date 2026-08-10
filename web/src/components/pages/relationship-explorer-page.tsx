import { NetworkIcon } from "lucide-react";

import { ContentWidth } from "@/components/layout/content-width";
import { MockDataAlert } from "@/components/shared/app-feedback";
import { EntityIcon, entityHref } from "@/components/shared/entity-link";
import { Button } from "@/components/ui/button";
import { InternalLink } from "@/components/ui/internal-link";
import { ScrollArea } from "@/components/ui/scroll-area";
import { internalPath } from "@/lib/internal-url";
import {
  getDocumentBody,
  getPerson,
  getWorkItem,
  mockRelations,
  selectRelations,
} from "@/lib/mock-data";
import type { EntityRef, Relation } from "@/lib/mock-data/types";
import { cn } from "@/lib/utils";

function focusEntity(
  entityType: string,
  entityId: string
): EntityRef | undefined {
  if (entityType === "work-item") {
    const item = getWorkItem(entityId);
    return item
      ? {
          id: item.id,
          type: "work-item",
          title: `${item.key} ${item.title}`,
        }
      : undefined;
  }
  if (entityType === "document") {
    const document = getDocumentBody(entityId);
    return document
      ? { id: document.documentId, type: "document", title: document.title }
      : undefined;
  }
  if (entityType === "person") {
    const person = getPerson(entityId);
    return person
      ? { id: person.id, type: "person", title: person.displayName }
      : undefined;
  }
  return undefined;
}

export function RelationshipExplorerPage({
  entityType,
  entityId,
}: {
  entityType: string;
  entityId: string;
}) {
  const focus = focusEntity(entityType, entityId);
  const directRelations = focus
    ? selectRelations({
        entity: { id: focus.id, type: focus.type },
      })
    : [];
  const relations = directRelations.length ? directRelations : mockRelations;
  const nodes = new Map<string, EntityRef>();
  for (const relation of relations) {
    nodes.set(`${relation.from.type}:${relation.from.id}`, relation.from);
    nodes.set(`${relation.to.type}:${relation.to.id}`, relation.to);
  }
  if (focus) nodes.set(`${focus.type}:${focus.id}`, focus);

  return (
    <ContentWidth
      width="graph"
      className="flex min-h-[calc(100svh-3.5rem)] flex-col gap-4"
    >
      <div className="flex flex-wrap items-start gap-3">
        <div className="min-w-0 flex-1">
          <h1 className="text-xl font-semibold">Relationship explorer</h1>
          <p className="text-muted-foreground text-sm">
            {focus?.title ?? `${entityType}:${entityId}`}
          </p>
        </div>
        <Button variant="outline">Scope: Namespace</Button>
        <Button variant="outline">Types: All</Button>
        <Button variant="outline">Depth: 1</Button>
      </div>
      <MockDataAlert
        title={
          directRelations.length
            ? "Illustrative direct relationships"
            : "Illustrative relationship graph"
        }
      >
        {directRelations.length
          ? "Nodes and edges come from centralized relation fixtures."
          : "No direct fixture relations match this focus. The graph shows a clearly labeled example cluster and does not claim these edges belong to the requested entity."}
      </MockDataAlert>

      <div className="grid min-h-135 flex-1 overflow-hidden rounded-xl border lg:grid-cols-[minmax(0,1fr)_21rem]">
        <div className="bg-surface-sunken relative min-h-105 overflow-auto p-8">
          {focus ? (
            <div className="bg-primary text-primary-foreground shadow-float absolute top-1/2 left-1/2 z-10 w-52 -translate-x-1/2 -translate-y-1/2 rounded-xl border p-4 text-center">
              <EntityIcon type={focus.type} className="mx-auto mb-2" />
              <p className="text-sm font-semibold">{focus.title}</p>
            </div>
          ) : (
            <div className="bg-card shadow-float absolute top-1/2 left-1/2 z-10 w-52 -translate-x-1/2 -translate-y-1/2 rounded-xl border p-4 text-center">
              <NetworkIcon className="text-muted-foreground mx-auto mb-2" />
              <p className="text-sm font-semibold">
                {entityType}:{entityId}
              </p>
            </div>
          )}

          {[...nodes.values()]
            .filter((node) => !focus || node.id !== focus.id)
            .slice(0, 6)
            .map((node, index) => {
              const positions = [
                "top-[12%] left-[15%]",
                "top-[12%] right-[15%]",
                "bottom-[12%] left-[15%]",
                "bottom-[12%] right-[15%]",
                "top-[43%] left-[6%]",
                "top-[43%] right-[6%]",
              ];
              return (
                <InternalLink
                  key={`${node.type}:${node.id}`}
                  to={internalPath(entityHref(node))}
                  className={cn(
                    "bg-card hover:border-primary absolute w-44 rounded-lg border p-3 shadow-sm",
                    positions[index]
                  )}
                >
                  <div className="flex items-center gap-2">
                    <EntityIcon type={node.type} />
                    <span className="truncate text-sm font-medium">
                      {node.title}
                    </span>
                  </div>
                </InternalLink>
              );
            })}
          <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
            <div className="border-primary/30 size-80 rounded-full border border-dashed" />
          </div>
        </div>
        <aside className="bg-background border-t lg:border-t-0 lg:border-l">
          <div className="border-b p-4">
            <h2 className="text-sm font-semibold">Direct relations</h2>
            <p className="text-muted-foreground text-xs">
              Expand only on intent.
            </p>
          </div>
          <ScrollArea className="h-120 p-3">
            {relations.map((relation: Relation) => (
              <div key={relation.id} className="mb-2 rounded-lg border p-3">
                <p className="text-muted-foreground text-xs capitalize">
                  {relation.kind.replaceAll("-", " ")}
                </p>
                <p className="mt-1 truncate text-sm font-medium">
                  {relation.from.title}
                </p>
                <p className="text-muted-foreground truncate text-xs">
                  → {relation.to.title}
                </p>
              </div>
            ))}
          </ScrollArea>
        </aside>
      </div>
    </ContentWidth>
  );
}
