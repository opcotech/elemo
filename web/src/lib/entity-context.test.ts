import type { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/errors";
import { ResourceType } from "@/lib/auth/permissions";
import {
  loadResourcePermissions,
  requireResourcePermission,
} from "@/lib/entity-context";

const readPermissions = [{ kind: "read" as const }];

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

  it("returns permissions when the required kind is granted", async () => {
    const queryClient = createQueryClient([{ kind: "write" }]);

    await expect(
      requireResourcePermission(
        queryClient,
        ResourceType.Organization,
        "organization-1",
        "read"
      )
    ).resolves.toEqual([{ kind: "write" }]);
  });

  it("rejects when the required permission is missing", async () => {
    const queryClient = createQueryClient([]);

    await expect(
      requireResourcePermission(
        queryClient,
        ResourceType.Namespace,
        "namespace-1",
        "delete"
      )
    ).rejects.toBeInstanceOf(ApiError);

    await expect(
      requireResourcePermission(
        queryClient,
        ResourceType.Namespace,
        "namespace-1",
        "delete"
      )
    ).rejects.toMatchObject({ status: 403 });
  });
});
