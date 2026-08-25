import { withErrorHandling } from "./error-handler";

import type { Client } from "@/lib/api/client";
import {
  v1OrganizationRoleGet,
  v1OrganizationRoleUpdate,
  v1OrganizationRolesCreate,
} from "@/lib/api/sdk";
import type { Role, RoleCreate, RolePatch } from "@/lib/api/types";

/**
 * Create a role via API.
 *
 * @param client - Authenticated API client
 * @param organizationId - Organization ID
 * @param roleData - Role data (name is required, description is optional)
 * @returns Created role with ID
 */
export async function createRole(
  client: Client,
  organizationId: string,
  roleData: Partial<RoleCreate> & { name: string }
): Promise<Role> {
  const roleCreateData: RoleCreate = {
    name: roleData.name,
    description: roleData.description,
    key: roleData.key,
    actions: roleData.actions,
  };

  const response = await withErrorHandling(
    async () => {
      return await v1OrganizationRolesCreate({
        client,
        path: { organizationRef: organizationId },
        body: roleCreateData,
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organizationId}/roles`,
      method: "POST",
    }
  );

  const role = await withErrorHandling(
    async () => {
      return await v1OrganizationRoleGet({
        client,
        path: { organizationRef: organizationId, role_id: response.data.id },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organizationId}/roles/${response.data.id}`,
      method: "GET",
    }
  );

  return role.data;
}

/**
 * Update a role bundle, including its actions.
 */
export async function updateRole(
  client: Client,
  organizationId: string,
  roleId: string,
  patch: RolePatch
): Promise<Role> {
  const response = await withErrorHandling(
    async () => {
      return await v1OrganizationRoleUpdate({
        client,
        path: { organizationRef: organizationId, role_id: roleId },
        body: patch,
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organizationId}/roles/${roleId}`,
      method: "PATCH",
    }
  );

  return response.data;
}
