import { SYSTEM_NIL_ID } from "@/lib/utils";

export type Action = string;

export const Action = {
  OrganizationCreate: "organization.create",
  OrganizationRead: "organization.read",
  OrganizationUpdate: "organization.update",
  OrganizationDelete: "organization.delete",
  OrganizationMembersManage: "organization.members.manage",
  NamespaceCreate: "namespace.create",
  NamespaceRead: "namespace.read",
  NamespaceUpdate: "namespace.update",
  NamespaceDelete: "namespace.delete",
  ProjectCreate: "project.create",
  ProjectRead: "project.read",
  ProjectUpdate: "project.update",
  ProjectDelete: "project.delete",
  ProjectMembersManage: "project.members.manage",
  IssueCreate: "issue.create",
  IssueRead: "issue.read",
  IssueUpdate: "issue.update",
  IssueDelete: "issue.delete",
  IssueAssign: "issue.assign",
  DocumentCreate: "document.create",
  DocumentRead: "document.read",
  DocumentUpdate: "document.update",
  DocumentDelete: "document.delete",
  FolderCreate: "folder.create",
  RoleManage: "role.manage",
  TeamManage: "team.manage",
  PermissionManage: "permission.manage",
} as const;

export const inspectableActions = Object.values(Action).filter(
  (action) => action !== Action.OrganizationCreate
);

export enum ResourceType {
  Assignment = "Assignment",
  Attachment = "Attachment",
  Comment = "Comment",
  Document = "Document",
  Folder = "Folder",
  Installation = "Installation",
  Issue = "Issue",
  IssueRelation = "IssueRelation",
  Label = "Label",
  Namespace = "Namespace",
  Notification = "Notification",
  Organization = "Organization",
  Permission = "Permission",
  Project = "Project",
  Role = "Role",
  Team = "Team",
  Todo = "Todo",
  User = "User",
  UserToken = "UserToken",
}

export function withResourceType(
  resourceType: ResourceType,
  resourceId?: string
) {
  return `${resourceType}:${resourceId || SYSTEM_NIL_ID}`;
}

/**
 * Checks whether the actor's effective actions include the requested action.
 * Match is exact: wildcards and write-implies-read are not supported.
 */
export function can(
  effective: { actions?: string[] } | readonly string[] | undefined,
  action: string
): boolean {
  if (!effective) {
    return false;
  }

  const actions = Array.isArray(effective)
    ? effective
    : "actions" in effective
      ? (effective.actions ?? [])
      : [];

  return actions.includes(action);
}
