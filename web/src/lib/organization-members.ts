import { zUserStatus } from "@/lib/api/schemas";
import type { OrganizationMember } from "@/lib/api/types";

/** Pending first, deleted last, then alphabetical by full name. */
export function sortOrganizationMembers(
  members: readonly OrganizationMember[]
): OrganizationMember[] {
  return [...members].sort((a, b) => {
    if (
      a.status === zUserStatus.enum.pending &&
      b.status !== zUserStatus.enum.pending
    ) {
      return -1;
    }
    if (
      a.status !== zUserStatus.enum.pending &&
      b.status === zUserStatus.enum.pending
    ) {
      return 1;
    }
    if (
      a.status === zUserStatus.enum.deleted &&
      b.status !== zUserStatus.enum.deleted
    ) {
      return 1;
    }
    if (
      a.status !== zUserStatus.enum.deleted &&
      b.status === zUserStatus.enum.deleted
    ) {
      return -1;
    }

    return `${a.first_name} ${a.last_name}`
      .toLowerCase()
      .localeCompare(`${b.first_name} ${b.last_name}`.toLowerCase());
  });
}
