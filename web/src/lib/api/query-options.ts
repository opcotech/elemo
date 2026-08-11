import "@/lib/api/client";

import {
  v1NotificationsGetOptions as generatedNotificationsGetOptions,
  v1OrganizationsGetOptions as generatedOrganizationsGetOptions,
  v1PermissionResourceGetOptions as generatedPermissionResourceGetOptions,
  v1TodosGetOptions as generatedTodosGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { cacheProfiles } from "@/lib/query-client";

export {
  v1NamespaceGetOptions,
  v1NamespacesProjectsGetOptions,
  v1OrganizationGetOptions,
  v1OrganizationMembersGetOptions,
  v1OrganizationRoleGetOptions,
  v1OrganizationRoleMembersGetOptions,
  v1OrganizationRolePermissionsGetOptions,
  v1OrganizationRolesGetOptions,
  v1OrganizationsNamespacesGetOptions,
  v1ProjectGetOptions,
  v1UserRequestPasswordResetOptions,
} from "@/lib/client/@tanstack/react-query.gen";

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
