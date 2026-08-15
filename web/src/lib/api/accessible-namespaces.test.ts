import { QueryClient } from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  accessibleNamespacesOptions,
  accessibleNamespacesQueryKey,
} from "@/lib/api/accessible-namespaces";
import { cacheProfiles } from "@/lib/query-client";

const sourceQueries = vi.hoisted(() => ({
  organizations: vi.fn(),
  namespaces: vi.fn(),
}));

vi.mock("@/lib/api/query-options", () => ({
  v1OrganizationsGetOptions: () => ({
    queryKey: ["organizations"],
    queryFn: sourceQueries.organizations,
  }),
  v1OrganizationsNamespacesGetOptions: ({
    path,
  }: {
    path: { id: string };
  }) => ({
    queryKey: ["organization-namespaces", path.id],
    queryFn: sourceQueries.namespaces,
  }),
}));

function listedPage<T>(items: T[]) {
  return { items, page_info: { has_more: false } };
}

const organization = {
  id: "organization-1",
  name: "Acme",
};

const namespace = {
  id: "namespace-1",
  name: "Product",
};

describe("accessible namespace cache", () => {
  beforeEach(() => {
    sourceQueries.organizations
      .mockReset()
      .mockResolvedValue(listedPage([organization]));
    sourceQueries.namespaces
      .mockReset()
      .mockResolvedValue(listedPage([namespace]));
  });

  it("uses the reference cache profile and reuses a fresh workspace", async () => {
    const queryClient = new QueryClient();
    const options = accessibleNamespacesOptions(queryClient);

    expect(options.staleTime).toBe(cacheProfiles.reference.staleTime);
    expect(options.gcTime).toBe(cacheProfiles.reference.gcTime);

    const first = await queryClient.fetchQuery(options);
    const second = await queryClient.fetchQuery(options);

    expect(second).toBe(first);
    expect(sourceQueries.organizations).toHaveBeenCalledTimes(1);
    expect(sourceQueries.namespaces).toHaveBeenCalledTimes(1);
  });

  it("refreshes source references after targeted invalidation", async () => {
    const queryClient = new QueryClient();
    const options = accessibleNamespacesOptions(queryClient);

    await queryClient.fetchQuery(options);
    await queryClient.invalidateQueries({
      queryKey: accessibleNamespacesQueryKey,
      exact: true,
      refetchType: "none",
    });
    await queryClient.fetchQuery(options);

    expect(sourceQueries.organizations).toHaveBeenCalledTimes(2);
    expect(sourceQueries.namespaces).toHaveBeenCalledTimes(2);
  });

  it("aggregates namespaces with their owning organization metadata", async () => {
    const secondOrganization = {
      id: "organization-2",
      name: "Globex",
    };
    const secondNamespace = {
      id: "namespace-2",
      name: "Platform",
    };
    sourceQueries.organizations.mockResolvedValue(
      listedPage([organization, secondOrganization])
    );
    sourceQueries.namespaces
      .mockResolvedValueOnce(listedPage([namespace]))
      .mockResolvedValueOnce(listedPage([secondNamespace]));

    const queryClient = new QueryClient();
    const result = await queryClient.fetchQuery(
      accessibleNamespacesOptions(queryClient)
    );

    expect(result.organizations).toEqual([organization, secondOrganization]);
    expect(result.namespaces).toEqual([
      {
        ...namespace,
        organization,
        organizationId: organization.id,
        organizationName: organization.name,
      },
      {
        ...secondNamespace,
        organization: secondOrganization,
        organizationId: secondOrganization.id,
        organizationName: secondOrganization.name,
      },
    ]);
  });

  it("skips namespace fan-out when no organizations are accessible", async () => {
    sourceQueries.organizations.mockResolvedValue(listedPage([]));
    const queryClient = new QueryClient();

    await expect(
      queryClient.fetchQuery(accessibleNamespacesOptions(queryClient))
    ).resolves.toEqual({ organizations: [], namespaces: [] });
    expect(sourceQueries.namespaces).not.toHaveBeenCalled();
  });

  it("does not cache a partially failed aggregation", async () => {
    sourceQueries.namespaces
      .mockRejectedValueOnce(new Error("namespace request failed"))
      .mockResolvedValueOnce(listedPage([namespace]));
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const options = accessibleNamespacesOptions(queryClient);

    await expect(queryClient.fetchQuery(options)).rejects.toThrow(
      "namespace request failed"
    );
    await expect(queryClient.fetchQuery(options)).resolves.toMatchObject({
      namespaces: [{ id: namespace.id }],
    });

    expect(sourceQueries.namespaces).toHaveBeenCalledTimes(2);
  });
});
