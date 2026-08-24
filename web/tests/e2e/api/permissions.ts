import { withErrorHandling } from "./error-handler";

import type { Client } from "@/lib/api/client";
import {
  v1PermissionDelete,
  v1PermissionResourceGet,
  v1PermissionsCreate,
} from "@/lib/api/sdk";
import type {
  Action,
  EffectiveActions,
  GrantCreate,
  GrantPrincipalType,
  ResourceType,
} from "@/lib/api/types";

export type CreateGrantInput = GrantCreate;

/**
 * Create a grant for a principal on a scope.
 *
 * Either role_id or a non-empty actions array is required.
 */
export async function createGrant(
  client: Client,
  grant: CreateGrantInput
): Promise<{ id: string }> {
  const response = await withErrorHandling(
    async () => {
      return await v1PermissionsCreate({
        client,
        body: grant,
        throwOnError: true,
      });
    },
    {
      endpoint: "/v1/permissions",
      method: "POST",
    }
  );

  return response.data;
}

/**
 * Grant actions to a user principal on a scope via the permissions API.
 */
export async function grantActions(
  client: Client,
  userId: string,
  scopeId: string,
  scopeType: ResourceType,
  actions: Action[]
): Promise<{ id: string }> {
  return await createGrant(client, {
    principal: {
      resourceType: "User",
      id: userId,
    },
    scope: {
      resourceType: scopeType,
      id: scopeId,
    },
    actions,
  });
}

/**
 * Grant a role bundle to a principal on a scope via the permissions API.
 */
export async function grantRoleToPrincipal(
  client: Client,
  principal: { resourceType: GrantPrincipalType; id: string },
  scope: { resourceType: ResourceType; id: string },
  roleId: string
): Promise<{ id: string }> {
  return await createGrant(client, {
    principal,
    scope,
    role_id: roleId,
  });
}

/**
 * Delete a grant by id.
 */
export async function deleteGrant(
  client: Client,
  grantId: string
): Promise<void> {
  await withErrorHandling(
    async () => {
      return await v1PermissionDelete({
        client,
        path: { id: grantId },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/permissions/${grantId}`,
      method: "DELETE",
    }
  );
}

/**
 * Get the caller's effective actions on a resource.
 */
export async function getEffectiveActions(
  client: Client,
  resourceType: ResourceType,
  resourceId: string
): Promise<EffectiveActions> {
  const response = await withErrorHandling(
    async () => {
      return await v1PermissionResourceGet({
        client,
        path: { resourceId: `${resourceType}:${resourceId}` },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/permissions/resources/${resourceType}:${resourceId}`,
      method: "GET",
    }
  );

  return response.data;
}
