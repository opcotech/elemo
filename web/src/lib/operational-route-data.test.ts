import type { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/errors";
import {
  loadNamespaceOperationalContext,
  loadProjectOperationalContext,
} from "@/lib/operational-route-data";

const organization = {
  id: "organization-1",
  name: "Acme",
  namespaces: ["namespace-1"],
};

const namespace = {
  id: "namespace-1",
  name: "Product",
  projects: [{ id: "project-1", name: "Web" }],
  organization,
  organizationId: organization.id,
  organizationName: organization.name,
};

const project = {
  id: "project-1",
  name: "Web",
  key: "WEB",
};

function queryId(options: { queryKey: readonly unknown[] }) {
  const key = options.queryKey[0];
  if (key && typeof key === "object" && "_id" in key) {
    return String((key as { _id: string })._id);
  }
  return JSON.stringify(options.queryKey);
}

function listedPage<T>(items: T[]) {
  return { items, page_info: { has_more: false } };
}

function isAccessibleNamespacesQuery(options: {
  queryKey: readonly unknown[];
}) {
  return (
    Array.isArray(options.queryKey) &&
    options.queryKey[0] === "elemo" &&
    options.queryKey[1] === "accessible-namespaces"
  );
}

function createQueryClient(
  resolve: (options: { queryKey: readonly unknown[] }) => unknown
) {
  return {
    fetchQuery: vi.fn((options: { queryKey: readonly unknown[] }) =>
      Promise.resolve(resolve(options))
    ),
  } as unknown as QueryClient;
}

describe("operational route loaders", () => {
  it("loads an accessible namespace after verifying read permission", async () => {
    const queryClient = createQueryClient((options) => {
      if (isAccessibleNamespacesQuery(options)) {
        return {
          organizations: [organization],
          namespaces: [namespace],
        };
      }
      if (queryId(options) === "v1PermissionResourceGet") {
        return [{ kind: "read" }];
      }
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadNamespaceOperationalContext(queryClient, "namespace-1")
    ).resolves.toEqual({
      namespace,
      organization,
    });
  });

  it("throws not-found when the namespace is outside the accessible workspace", async () => {
    const queryClient = createQueryClient(() => ({
      organizations: [organization],
      namespaces: [],
    }));

    await expect(
      loadNamespaceOperationalContext(queryClient, "missing-namespace")
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("redirects to permission-denied when namespace read is forbidden", async () => {
    const queryClient = createQueryClient((options) => {
      if (isAccessibleNamespacesQuery(options)) {
        return {
          organizations: [organization],
          namespaces: [namespace],
        };
      }
      return [];
    });

    await expect(
      loadNamespaceOperationalContext(queryClient, "namespace-1")
    ).rejects.toMatchObject({
      options: { to: "/permission-denied" },
    });
  });

  it("rejects projects that do not belong to the namespace URL", async () => {
    const queryClient = createQueryClient((options) => {
      if (isAccessibleNamespacesQuery(options)) {
        return {
          organizations: [organization],
          namespaces: [
            {
              ...namespace,
              projects: [{ id: "other-project", name: "Other" }],
            },
          ],
        };
      }
      if (queryId(options) === "v1PermissionResourceGet") {
        return [{ kind: "read" }];
      }
      if (queryId(options) === "v1NamespacesProjectsGet") {
        return listedPage([{ id: "other-project", name: "Other" }]);
      }
      if (queryId(options) === "v1ProjectGet") {
        return project;
      }
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadProjectOperationalContext(queryClient, "namespace-1", "project-1")
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("loads a project that belongs to the namespace hierarchy", async () => {
    const queryClient = createQueryClient((options) => {
      if (isAccessibleNamespacesQuery(options)) {
        return {
          organizations: [organization],
          namespaces: [namespace],
        };
      }
      if (queryId(options) === "v1PermissionResourceGet") {
        return [{ kind: "read" }];
      }
      if (queryId(options) === "v1NamespacesProjectsGet") {
        return listedPage([project]);
      }
      if (queryId(options) === "v1ProjectGet") {
        return project;
      }
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadProjectOperationalContext(queryClient, "namespace-1", "project-1")
    ).resolves.toEqual({
      namespace,
      organization,
      project,
    });
  });

  it("maps API not-found errors onto the router not-found boundary", async () => {
    const queryClient = createQueryClient(() => {
      throw new ApiError(404, "missing");
    });

    await expect(
      loadNamespaceOperationalContext(queryClient, "namespace-1")
    ).rejects.toMatchObject({ isNotFound: true });
  });
});
