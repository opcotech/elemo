import type { AppEntityType } from "@/lib/entity-types";
import type { RecentEntityType } from "@/lib/ui-store";

/** Maps a remembered recent entity type to an `EntityLink` / icon type. */
export function recentEntityLinkType(type: RecentEntityType): AppEntityType {
  if (type === "work-item") {
    return "work-item";
  }
  return type;
}
