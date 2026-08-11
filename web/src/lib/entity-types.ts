import type { EntityRef } from "@/lib/mock-data/types";

export type AppEntityType =
  | "organization"
  | "namespace"
  | "project"
  | "work-item"
  | "document"
  | "person"
  | "relation"
  | "saved-view";

export function entityHref(entity: EntityRef): string {
  if (entity.type === "work-item") return `/work/${entity.id}`;
  if (entity.type === "document") return `/documents/${entity.id}`;
  return `/search?type=person&q=${encodeURIComponent(entity.title)}`;
}
