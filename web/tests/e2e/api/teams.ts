import { withErrorHandling } from "./error-handler";

import type { Client } from "@/lib/client/client";
import {
  v1OrganizationTeamGet,
  v1OrganizationTeamMemberRemove,
  v1OrganizationTeamMembersAdd,
  v1OrganizationTeamMembersGet,
  v1OrganizationTeamsCreate,
} from "@/lib/client/sdk.gen";
import type { Team, TeamCreate, User } from "@/lib/client/types.gen";

/**
 * Create a team under an organization.
 */
export async function createTeam(
  client: Client,
  organizationId: string,
  teamData: Partial<TeamCreate> & { name: string }
): Promise<Team> {
  const response = await withErrorHandling(
    async () => {
      return await v1OrganizationTeamsCreate({
        client,
        path: { id: organizationId },
        body: {
          name: teamData.name,
          description: teamData.description,
        },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organizationId}/teams`,
      method: "POST",
    }
  );

  const teamId = response.data.id || "";

  const team = await withErrorHandling(
    async () => {
      return await v1OrganizationTeamGet({
        client,
        path: { id: organizationId, team_id: teamId },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organizationId}/teams/${teamId}`,
      method: "GET",
    }
  );

  return team.data;
}

/**
 * Add a user to an organization team.
 */
export async function addTeamMember(
  client: Client,
  organizationId: string,
  teamId: string,
  userId: string
): Promise<void> {
  await withErrorHandling(
    async () => {
      return await v1OrganizationTeamMembersAdd({
        client,
        path: { id: organizationId, team_id: teamId },
        body: { user_id: userId },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organizationId}/teams/${teamId}/members`,
      method: "POST",
    }
  );
}

/**
 * Remove a user from an organization team.
 */
export async function removeTeamMember(
  client: Client,
  organizationId: string,
  teamId: string,
  userId: string
): Promise<void> {
  await withErrorHandling(
    async () => {
      return await v1OrganizationTeamMemberRemove({
        client,
        path: { id: organizationId, team_id: teamId, user_id: userId },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organizationId}/teams/${teamId}/members/${userId}`,
      method: "DELETE",
    }
  );
}

/**
 * List members of an organization team.
 */
export async function listTeamMembers(
  client: Client,
  organizationId: string,
  teamId: string
): Promise<User[]> {
  const response = await withErrorHandling(
    async () => {
      return await v1OrganizationTeamMembersGet({
        client,
        path: { id: organizationId, team_id: teamId },
        throwOnError: true,
      });
    },
    {
      endpoint: `/v1/organizations/${organizationId}/teams/${teamId}/members`,
      method: "GET",
    }
  );

  return response.data.items ?? [];
}
