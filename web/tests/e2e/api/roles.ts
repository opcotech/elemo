import { withErrorHandling } from "./error-handler";

import type { Client } from "@/lib/client/client";
import {
  v1OrganizationRoleGet,
  v1OrganizationRolesCreate,
} from "@/lib/client/sdk.gen";
import type { Role, RoleCreate } from "@/lib/client/types.gen";

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
  };

  const response = await withErrorHandling(
    async () => {
      return await v1OrganizationRolesCreate({
        client,
        path: { id: organizationId },
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
        path: { id: organizationId, role_id: response.data.id },
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
