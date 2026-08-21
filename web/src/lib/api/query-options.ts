import "@/lib/api/client";

import { keepPreviousData } from "@tanstack/react-query";

import {
  v1DocumentGetOptions as generatedDocumentGetOptions,
  v1IssueGetOptions as generatedIssueGetOptions,
  v1IssueRelationsGetOptions as generatedIssueRelationsGetOptions,
  v1LabelsGetOptions as generatedLabelsGetOptions,
  v1NamespacesIssuesGetOptions as generatedNamespacesIssuesGetOptions,
  v1NamespacesIssuesKeyGetOptions as generatedNamespacesIssuesKeyGetOptions,
  v1NotificationsGetOptions as generatedNotificationsGetOptions,
  v1OrganizationsGetOptions as generatedOrganizationsGetOptions,
  v1PermissionResourceGetOptions as generatedPermissionResourceGetOptions,
  v1ProjectsIssuesGetOptions as generatedProjectsIssuesGetOptions,
  v1SearchGetOptions as generatedSearchGetOptions,
  v1TodosGetOptions as generatedTodosGetOptions,
  v1UsersIssuesGetOptions as generatedUsersIssuesGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { cacheProfiles } from "@/lib/query-client";

export {
  v1IssuesDocumentsGetOptions,
  v1NamespaceGetOptions,
  v1NamespacesDocumentsGetOptions,
  v1NamespacesFoldersGetOptions,
  v1NamespacesGetOptions,
  v1NamespacesProjectsGetOptions,
  v1OrganizationGetOptions,
  v1OrganizationMembersGetOptions,
  v1OrganizationRoleGetOptions,
  v1OrganizationRolesGetOptions,
  v1OrganizationTeamGetOptions,
  v1OrganizationTeamMembersGetOptions,
  v1OrganizationTeamsGetOptions,
  v1OrganizationsDocumentsGetOptions,
  v1OrganizationsFoldersGetOptions,
  v1OrganizationsNamespacesGetOptions,
  v1FolderGetOptions,
  v1ProjectGetOptions,
  v1ProjectsDocumentsGetOptions,
  v1UserRequestPasswordResetOptions,
} from "@/lib/client/@tanstack/react-query.gen";

export function v1DocumentGetOptions(
  ...args: Parameters<typeof generatedDocumentGetOptions>
) {
  return {
    ...generatedDocumentGetOptions(...args),
    ...cacheProfiles.volatile,
  };
}

export function v1OrganizationsGetOptions(
  ...args: Parameters<typeof generatedOrganizationsGetOptions>
) {
  return {
    ...generatedOrganizationsGetOptions(...args),
    ...cacheProfiles.reference,
  };
}

export function v1PermissionResourceGetOptions(
  ...args: Parameters<typeof generatedPermissionResourceGetOptions>
) {
  return {
    ...generatedPermissionResourceGetOptions(...args),
    ...cacheProfiles.volatile,
  };
}

export function v1NotificationsGetOptions(
  ...args: Parameters<typeof generatedNotificationsGetOptions>
) {
  return {
    ...generatedNotificationsGetOptions(...args),
    ...cacheProfiles.volatile,
  };
}

export function v1TodosGetOptions(
  ...args: Parameters<typeof generatedTodosGetOptions>
) {
  return {
    ...generatedTodosGetOptions(...args),
    ...cacheProfiles.volatile,
  };
}

export function v1ProjectsIssuesGetOptions(
  ...args: Parameters<typeof generatedProjectsIssuesGetOptions>
) {
  return {
    ...generatedProjectsIssuesGetOptions(...args),
    ...cacheProfiles.entity,
  };
}

export function v1NamespacesIssuesGetOptions(
  ...args: Parameters<typeof generatedNamespacesIssuesGetOptions>
) {
  return {
    ...generatedNamespacesIssuesGetOptions(...args),
    ...cacheProfiles.entity,
  };
}

export function v1UsersIssuesGetOptions(
  ...args: Parameters<typeof generatedUsersIssuesGetOptions>
) {
  return {
    ...generatedUsersIssuesGetOptions(...args),
    ...cacheProfiles.entity,
  };
}

export function v1IssueGetOptions(
  ...args: Parameters<typeof generatedIssueGetOptions>
) {
  return {
    ...generatedIssueGetOptions(...args),
    ...cacheProfiles.volatile,
  };
}

export function v1IssueRelationsGetOptions(
  ...args: Parameters<typeof generatedIssueRelationsGetOptions>
) {
  return {
    ...generatedIssueRelationsGetOptions(...args),
    ...cacheProfiles.volatile,
  };
}

export function v1NamespacesIssuesKeyGetOptions(
  ...args: Parameters<typeof generatedNamespacesIssuesKeyGetOptions>
) {
  return {
    ...generatedNamespacesIssuesKeyGetOptions(...args),
    ...cacheProfiles.volatile,
  };
}

export function v1LabelsGetOptions(
  ...args: Parameters<typeof generatedLabelsGetOptions>
) {
  return {
    ...generatedLabelsGetOptions(...args),
    ...cacheProfiles.reference,
  };
}

export function v1SearchGetOptions(
  ...args: Parameters<typeof generatedSearchGetOptions>
) {
  return {
    ...generatedSearchGetOptions(...args),
    ...cacheProfiles.volatile,
    placeholderData: keepPreviousData,
  };
}
