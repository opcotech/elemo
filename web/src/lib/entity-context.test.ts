import type { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/errors";
import { Action, ResourceType } from "@/lib/auth/permissions";
import {
  loadResourcePermissions,
  requireResourcePermission,
} from "@/lib/entity-context";

const readPermissions = { actions: [Action.OrganizationRead] };

function createQueryClient(permissions: unknown) {
  const fetchQuery = vi.fn(() => Promise.resolve(permissions));

  return {
    fetchQuery,
  } as unknown as QueryClient;
}

describe("entity permission context", () => {
  it("loads permissions through the shared resource helper", async () => {
    const queryClient = createQueryClient(readPermissions);

    await expect(
      loadResourcePermissions(queryClient, ResourceType.Project, "project-1")
    ).resolves.toEqual(readPermissions);

    expect(queryClient.fetchQuery).toHaveBeenCalledTimes(1);
  });

  it("returns permissions when the required action is granted", async () => {
    const queryClient = createQueryClient(readPermissions);

    await expect(
      requireResourcePermission(
        queryClient,
        ResourceType.Organization,
        "organization-1",
        Action.OrganizationRead
      )
    ).resolves.toEqual(readPermissions);
  });

  it("rejects when the required permission is missing", async () => {
    const queryClient = createQueryClient({ actions: [] });

    await expect(
      requireResourcePermission(
        queryClient,
        ResourceType.Namespace,
        "namespace-1",
        Action.NamespaceDelete
      )
    ).rejects.toBeInstanceOf(ApiError);

    await expect(
      requireResourcePermission(
        queryClient,
        ResourceType.Namespace,
        "namespace-1",
        Action.NamespaceDelete
      )
    ).rejects.toMatchObject({ status: 403 });
  });
});
