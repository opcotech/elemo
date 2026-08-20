import { withErrorHandling } from "./error-handler";
import { grantActions } from "./permissions";
import { getRandomString } from "../utils/random";

import type { Client } from "@/lib/client/client";
import {
  v1OrganizationGet,
  v1OrganizationMembersAdd,
  v1OrganizationsCreate,
} from "@/lib/client/sdk.gen";
import type {
  Action,
  Organization,
  OrganizationCreate,
} from "@/lib/client/types.gen";

/**
 * Create an organization via API.
 *
 * The authenticated client must have organization.create on Installation.
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
 * Add a member to an organization and optionally grant extra actions.
 *
 * Organization membership already confers organization.read via the org
 * principal's org-member grant. Pass actions to grant additional capabilities
 * directly to the user on the organization scope.
 */
export async function addMemberToOrganization(
  client: Client,
  orgId: string,
  userId: string,
  actions: Action[] = []
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

  if (actions.length > 0) {
    await grantActions(client, userId, orgId, "Organization", actions);
  }
}
