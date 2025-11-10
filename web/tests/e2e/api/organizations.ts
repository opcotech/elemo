import { withErrorHandling } from "./error-handler";
import { grantPermission } from "./permissions";
import { getRandomString } from "../utils/random";

import type { Client } from "@/lib/client/client";
import {
  v1OrganizationGet,
  v1OrganizationMembersAdd,
  v1OrganizationsCreate,
} from "@/lib/client/sdk.gen";
import type {
  Organization,
  OrganizationCreate,
  PermissionKind,
} from "@/lib/client/types.gen";

/**
 * Create an organization via API.
 *
 * @param client - Authenticated API client
 * @param orgData - Organization data (name and email are required)
 * @returns Created organization with ID
 */
export async function createOrganization(
  client: Client,
  orgData: Partial<OrganizationCreate> & { name: string; email: string }
): Promise<Organization> {
  const orgCreateData: OrganizationCreate = {
    name: orgData.name,
    email: orgData.email,
    logo: orgData.logo || "https://picsum.photos/id/64/100/100",
    website:
      orgData.website ||
      `https://${getRandomString(8).toLowerCase()}.example.com`,
  };

  const response = await withErrorHandling(
    async () => {
      return await v1OrganizationsCreate({
        client,
        body: orgCreateData,
        throwOnError: true,
      });
    },
    {
      endpoint: "/v1/organizations",
      method: "POST",
    }
  );

  const orgId = response.data.id || "";

  // Fetch the full organization to get all fields
  const orgResponse = await withErrorHandling(
    async () => {
      return await v1OrganizationGet({
        client,
        path: { id: orgId },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${orgId}`,
      method: "GET",
    }
  );

  return orgResponse.data;
}

/**
 * Add a member to an organization with a specific permission.
 *
 * @param client - Authenticated API client
 * @param orgId - Organization ID
 * @param userId - User ID to add
 * @param permissionKind - Permission kind (default: "*")
 */
export async function addMemberToOrganization(
  client: Client,
  orgId: string,
  userId: string,
  permissionKind: PermissionKind = "*"
): Promise<void> {
  await withErrorHandling(
    async () => {
      return await v1OrganizationMembersAdd({
        client,
        path: { id: orgId },
        body: { user_id: userId },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${orgId}/members`,
      method: "POST",
    }
  );

  // Grant permission to the user for the organization
  await grantPermission(client, userId, orgId, "Organization", permissionKind);
}
