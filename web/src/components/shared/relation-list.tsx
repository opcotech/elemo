import { Link2Icon } from "lucide-react";

import { AppEmptyState } from "@/components/shared/app-feedback";
import {
  AppList,
  EntityLink,
  entityHref,
} from "@/components/shared/entity-link";
import type { EntityRef, Relation } from "@/lib/mock-data/types";

export function RelationList({
  relations,
  entity,
  limit,
}: {
  relations: readonly Relation[];
  entity: Pick<EntityRef, "id" | "type">;
  limit?: number;
}) {
  const visible = limit ? relations.slice(0, limit) : relations;
  if (visible.length === 0) {
    return (
      <AppEmptyState
        compact
        icon={<Link2Icon />}
        title="No relationships"
        description="Connected work and documents will appear here."
      />
    );
  }

  return (
    <AppList>
      {visible.map((relation) => {
        const target =
          relation.from.id === entity.id && relation.from.type === entity.type
            ? relation.to
            : relation.from;
        return (
          <EntityLink
            key={relation.id}
            href={entityHref(target)}
            type={target.type}
            title={target.title}
            subtitle={relation.kind.replaceAll("-", " ")}
          />
        );
      })}
    </AppList>
  );
}
