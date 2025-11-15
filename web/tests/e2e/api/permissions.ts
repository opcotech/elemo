import { withErrorHandling } from "./error-handler";

import type { Client } from "@/lib/client/client";
import { v1PermissionsCreate } from "@/lib/client/sdk.gen";
import type {
  PermissionCreate,
  PermissionKind,
  ResourceType,
} from "@/lib/client/types.gen";

/**
 * Grant a permission to a user for a specific resource.
 *
 * @param client - Authenticated API client
 * @param userId - User ID (subject)
 * @param targetId - Target resource ID
 * @param targetResourceType - Target resource type
 * @param permissionKind - Permission kind
 */
export async function grantPermission(
  client: Client,
  userId: string,
  targetId: string,
  targetResourceType: ResourceType,
  permissionKind: PermissionKind
): Promise<void> {
  const permissionData: PermissionCreate = {
    kind: permissionKind,
    subject: {
      resourceType: "User",
      id: userId,
    },
    target: {
      resourceType: targetResourceType,
      id: targetId,
    },
  };

  await withErrorHandling(
    async () => {
      return await v1PermissionsCreate({
        client,
        body: permissionData,
        throwOnError: true,
      });
    },
    {
      endpoint: "/v1/permissions",
      method: "POST",
    }
  );
}
