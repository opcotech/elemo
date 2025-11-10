import { useQuery } from "@tanstack/react-query";

import { v1PermissionResourceGetOptions } from "@/lib/api";
import { SYSTEM_NIL_ID } from "@/lib/utils";

export enum ResourceType {
  Assignment = "Assignment",
  Attachment = "Attachment",
  Comment = "Comment",
  Document = "Document",
  Issue = "Issue",
  IssueRelation = "IssueRelation",
  Label = "Label",
  Namespace = "Namespace",
  Notification = "Notification",
  Organization = "Organization",
  Permission = "Permission",
  Project = "Project",
  Role = "Role",
  Todo = "Todo",
  User = "User",
  UserToken = "UserToken",
}

/**
 * Constructs a resource ID string for a given resource type and ID.
 *
 * If the resource ID is not provided, it will return a resource ID string with
 * a nil ID -- which is used to check system-level permissions.
 *
 * @param resourceType - The type of resource to check permissions for.
 * @param resourceId - The ID of the resource to check permissions for.
 * @returns The resource ID string.
 */
export function withResourceType(
  resourceType: ResourceType,
  resourceId?: string
) {
  return `${resourceType}:${resourceId ? `${resourceId}` : SYSTEM_NIL_ID}`;
}

export function usePermissions(
  resourceId: string,
  hasSSRResponse: boolean = false
) {
  return useQuery({
    enabled: !hasSSRResponse,
    ...v1PermissionResourceGetOptions({
      path: {
        resourceId,
      },
    }),
  });
}
