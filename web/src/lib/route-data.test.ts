import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/errors";
import { loadOrganizationWorkspace } from "@/lib/organization-workspace";
import {
  loadNamespaceDetail,
  loadNamespaceHierarchy,
  loadProjectDetail,
  loadProjectHierarchy,
} from "@/lib/route-data";

const organization = {
  id: "organization-1",
  name: "Acme",
};

const namespace = {
  id: "namespace-1",
  name: "Product",
};

const project = {
  id: "project-1",
  name: "Web",
  key: "WEB",
};

const readPermissions = [{ kind: "read" }];

function listedPage<T>(items: T[]) {
  return { items, page_info: { has_more: false } };
}

function queryId(options: { queryKey: readonly unknown[] }) {
  const key = options.queryKey[0];
  if (key && typeof key === "object" && "_id" in key) {
    return String((key as { _id: string })._id);
  }
  return JSON.stringify(options.queryKey);
}

function createQueryClient(
  resolve: (options: { queryKey: readonly unknown[] }) => unknown
) {
  const fetchQuery = vi.fn((options: { queryKey: readonly unknown[] }) =>
    Promise.resolve(resolve(options))
  );

  return {
    fetchQuery,
  } as unknown as QueryClient;
}

describe("settings hierarchy loaders", () => {
  it("loads organization and namespace when the hierarchy matches", async () => {
    const queryClient = createQueryClient((options) => {
      if (queryId(options) === "v1OrganizationGet") return organization;
      if (queryId(options) === "v1NamespaceGet") return namespace;
      if (queryId(options) === "v1OrganizationsNamespacesGet") {
        return listedPage([namespace]);
      }
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadNamespaceHierarchy(queryClient, "organization-1", "namespace-1")
    ).resolves.toEqual({ organization, namespace });
  });

  it("throws not-found when the namespace is not owned by the organization", async () => {
    const queryClient = createQueryClient((options) => {
      if (queryId(options) === "v1OrganizationGet") return organization;
      if (queryId(options) === "v1NamespaceGet") return namespace;
      if (queryId(options) === "v1OrganizationsNamespacesGet") {
        return listedPage([{ id: "other-namespace" }]);
      }
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadNamespaceHierarchy(queryClient, "organization-1", "namespace-1")
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("includes verified namespace permissions in detail context", async () => {
    const queryClient = createQueryClient((options) => {
      if (queryId(options) === "v1OrganizationGet") return organization;
      if (queryId(options) === "v1NamespaceGet") return namespace;
      if (queryId(options) === "v1OrganizationsNamespacesGet") {
        return listedPage([namespace]);
      }
      if (queryId(options) === "v1PermissionResourceGet")
        return readPermissions;
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadNamespaceDetail(queryClient, "organization-1", "namespace-1")
    ).resolves.toEqual({
      organization,
      namespace,
      permissions: readPermissions,
    });
  });

  it("rejects namespace detail without read permission", async () => {
    const queryClient = createQueryClient((options) => {
      if (queryId(options) === "v1OrganizationGet") return organization;
      if (queryId(options) === "v1NamespaceGet") return namespace;
      if (queryId(options) === "v1OrganizationsNamespacesGet") {
        return listedPage([namespace]);
      }
      if (queryId(options) === "v1PermissionResourceGet") return [];
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadNamespaceDetail(queryClient, "organization-1", "namespace-1")
    ).rejects.toMatchObject({ status: 403 });
  });

  it("throws not-found when the project is outside the namespace", async () => {
    const queryClient = createQueryClient((options) => {
      if (queryId(options) === "v1OrganizationGet") return organization;
      if (queryId(options) === "v1NamespaceGet") return namespace;
      if (queryId(options) === "v1OrganizationsNamespacesGet") {
        return listedPage([namespace]);
      }
      if (queryId(options) === "v1NamespacesProjectsGet") {
        return listedPage([{ id: "other-project" }]);
      }
      if (queryId(options) === "v1ProjectGet") return project;
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadProjectHierarchy(
        queryClient,
        "organization-1",
        "namespace-1",
        "project-1"
      )
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("loads a project after validating the namespace membership", async () => {
    const queryClient = createQueryClient((options) => {
      if (queryId(options) === "v1OrganizationGet") return organization;
      if (queryId(options) === "v1NamespaceGet") return namespace;
      if (queryId(options) === "v1OrganizationsNamespacesGet") {
        return listedPage([namespace]);
      }
      if (queryId(options) === "v1NamespacesProjectsGet") {
        return listedPage([project]);
      }
      if (queryId(options) === "v1ProjectGet") return project;
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadProjectHierarchy(
        queryClient,
        "organization-1",
        "namespace-1",
        "project-1"
      )
    ).resolves.toEqual({ organization, namespace, project });
  });

  it("refreshes the namespace when loading project detail", async () => {
    const queryClient = createQueryClient((options) => {
      if (queryId(options) === "v1OrganizationGet") return organization;
      if (queryId(options) === "v1NamespaceGet") return namespace;
      if (queryId(options) === "v1OrganizationsNamespacesGet") {
        return listedPage([namespace]);
      }
      if (queryId(options) === "v1NamespacesProjectsGet") {
        return listedPage([project]);
      }
      if (queryId(options) === "v1PermissionResourceGet")
        return readPermissions;
      if (queryId(options) === "v1ProjectGet") return project;
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadProjectDetail(
        queryClient,
        "organization-1",
        "namespace-1",
        "project-1"
      )
    ).resolves.toEqual({
      organization,
      namespace,
      project,
      permissions: readPermissions,
    });

    expect(queryClient.fetchQuery).toHaveBeenCalled();
  });

  it("returns a coherent organization bundle when read access is granted", async () => {
    const members = [{ id: "member-1" }];
    const namespaces = [namespace];
    const roles = [{ id: "role-1", name: "Owner" }];
    const queryClient = createQueryClient((options) => {
      const id = queryId(options);
      if (id === "v1OrganizationGet") return organization;
      if (id === "v1PermissionResourceGet") return readPermissions;
      if (id === "v1OrganizationMembersGet") return listedPage(members);
      if (id === "v1OrganizationsNamespacesGet") return listedPage(namespaces);
      if (id === "v1OrganizationRolesGet") return listedPage(roles);
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadOrganizationWorkspace(queryClient, "organization-1")
    ).resolves.toEqual({
      organization,
      permissions: readPermissions,
      members,
      namespaces,
      roles,
      hasReadAccess: true,
    });
  });

  it("does not load organization children without read access", async () => {
    const queryClient = createQueryClient((options) => {
      const id = queryId(options);
      if (id === "v1OrganizationGet") return organization;
      if (id === "v1PermissionResourceGet") return [];
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadOrganizationWorkspace(queryClient, "organization-1")
    ).resolves.toEqual({
      organization,
      permissions: [],
      members: [],
      namespaces: [],
      roles: [],
      hasReadAccess: false,
    });

    expect(queryClient.fetchQuery).toHaveBeenCalledTimes(2);
  });

  it("throws not-found from project detail when the refreshed namespace lacks the project", async () => {
    const queryClient = createQueryClient((options) => {
      if (queryId(options) === "v1OrganizationGet") return organization;
      if (queryId(options) === "v1NamespaceGet") return namespace;
      if (queryId(options) === "v1OrganizationsNamespacesGet") {
        return listedPage([namespace]);
      }
      if (queryId(options) === "v1NamespacesProjectsGet") {
        return listedPage([{ id: "other-project" }]);
      }
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadProjectDetail(
        queryClient,
        "organization-1",
        "namespace-1",
        "project-1"
      )
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("throws permission denied when the user cannot read the project", async () => {
    const queryClient = createQueryClient((options) => {
      if (queryId(options) === "v1OrganizationGet") return organization;
      if (queryId(options) === "v1NamespaceGet") return namespace;
      if (queryId(options) === "v1OrganizationsNamespacesGet") {
        return listedPage([namespace]);
      }
      if (queryId(options) === "v1NamespacesProjectsGet") {
        return listedPage([project]);
      }
      if (queryId(options) === "v1PermissionResourceGet") return [];
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadProjectDetail(
        queryClient,
        "organization-1",
        "namespace-1",
        "project-1"
      )
    ).rejects.toBeInstanceOf(ApiError);

    await expect(
      loadProjectDetail(
        queryClient,
        "organization-1",
        "namespace-1",
        "project-1"
      )
    ).rejects.toMatchObject({ status: 403 });
  });
});

describe("loader cache freshness after invalidation", () => {
  it("refetches with fetchQuery after invalidation while ensureQueryData stays stale", async () => {
    let membersVersion = 0;
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          staleTime: 5 * 60 * 1000,
          retry: false,
        },
      },
    });

    const membersQuery = {
      queryKey: ["members", "organization-1"] as const,
      queryFn: () => {
        membersVersion += 1;
        return membersVersion === 1
          ? [{ id: "member-1", name: "Original" }]
          : [
              { id: "member-1", name: "Original" },
              { id: "member-2", name: "Invited" },
            ];
      },
      staleTime: 5 * 60 * 1000,
    };

    await expect(queryClient.fetchQuery(membersQuery)).resolves.toHaveLength(1);

    await queryClient.invalidateQueries({ queryKey: membersQuery.queryKey });

    // Loader-only routes previously used ensureQueryData, which returns any
    // cached value and never waits for an invalidated refetch.
    await expect(
      queryClient.ensureQueryData(membersQuery)
    ).resolves.toHaveLength(1);

    // fetchQuery respects staleness and returns the post-mutation snapshot.
    await expect(queryClient.fetchQuery(membersQuery)).resolves.toHaveLength(2);
  });
});
