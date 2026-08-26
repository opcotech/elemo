import type { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/errors";
import {
  loadNamespaceOperationalContext,
  loadProjectOperationalContext,
} from "@/lib/operational-route-data";

const organization = {
  id: "organization-1",
  slug: "acme",
  name: "Acme",
};

const namespace = {
  id: "namespace-1",
  slug: "product",
  name: "Product",
  organization,
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

function createQueryClient(
  resolve: (options: { queryKey: readonly unknown[] }) => unknown
) {
  return {
    fetchQuery: vi.fn((options: { queryKey: readonly unknown[] }) =>
      Promise.resolve(resolve(options))
    ),
    setQueryData: vi.fn(),
  } as unknown as QueryClient;
}

describe("operational route loaders", () => {
  it("loads a reachable namespace without requiring namespace.read", async () => {
    const queryClient = createQueryClient((options) => {
      if (queryId(options) === "v1NamespaceGet") {
        return namespace;
      }
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadNamespaceOperationalContext(queryClient, "acme", "product")
    ).resolves.toEqual({
      namespace,
      organization,
    });
  });

  it("rejects xid-shaped namespace slugs without looking the namespace up", async () => {
    const queryClient = createQueryClient(() => {
      throw new Error("lookup should not run");
    });

    await expect(
      loadNamespaceOperationalContext(
        queryClient,
        "acme",
        "9bsv0s46s6s002p9ltq0"
      )
    ).rejects.toMatchObject({ isNotFound: true });
  });

  it("loads a project by namespace-scoped key with only project.read", async () => {
    const queryClient = createQueryClient((options) => {
      const id = queryId(options);
      if (id === "v1NamespaceGet") {
        return namespace;
      }
      if (id === "v1NamespacesProjectsKeyGet") {
        return project;
      }
      if (id === "v1PermissionResourceGet") {
        return { actions: ["project.read"] };
      }
      throw new Error(`Unexpected query ${JSON.stringify(options.queryKey)}`);
    });

    await expect(
      loadProjectOperationalContext(queryClient, "acme", "product", "WEB")
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
      loadNamespaceOperationalContext(queryClient, "acme", "product")
    ).rejects.toMatchObject({ isNotFound: true });
  });
});
