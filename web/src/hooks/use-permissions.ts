import { useQuery } from "@tanstack/react-query";

import { v1PermissionResourceGetOptions } from "@/lib/api";

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

export function withResourceType(
  resourceType: ResourceType,
  resourceId: string
) {
  return `${resourceType}:${resourceId}`;
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
