import { useQueries, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useMemo } from "react";

import { AllNamespacesList } from "@/components/namespaces/all-namespaces-list";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import {
  v1OrganizationsGetOptions,
  v1OrganizationsNamespacesGetOptions,
} from "@/lib/api";
import type { Namespace } from "@/lib/api";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

interface NamespaceWithOrganization extends Namespace {
  organizationId: string;
  organizationName: string;
}

export const Route = createFileRoute("/settings/namespaces/")({
  beforeLoad: requireAuthBeforeLoad,
  component: NamespacesSettingsPage,
});

function NamespacesSettingsPage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();

  const {
    data: organizations,
    isLoading: isLoadingOrgs,
    error: organizationsError,
  } = useQuery(v1OrganizationsGetOptions());

  // Fetch namespaces for all organizations using useQueries
  const namespaceQueries = useQueries({
    queries:
      organizations && organizations.length > 0
        ? organizations.map((org) =>
            v1OrganizationsNamespacesGetOptions({
              path: { id: org.id },
            })
          )
        : [],
  });

  // Combine all namespaces with organization info
  const allNamespaces = useMemo(() => {
    if (!organizations) return [];
    const results: NamespaceWithOrganization[] = [];
    namespaceQueries.forEach((query, index) => {
      const org = organizations[index];
      if (org && query.data) {
        for (const ns of query.data) {
          results.push({
            ...ns,
            organizationId: org.id,
            organizationName: org.name,
          });
        }
      }
    });
    return results;
  }, [organizations, namespaceQueries]);

  const isLoadingNamespaces = namespaceQueries.some((q) => q.isLoading);
  const isLoading = isLoadingOrgs || isLoadingNamespaces;
  const error =
    organizationsError ||
    namespaceQueries.find((q) => q.error)?.error ||
    undefined;

  useEffect(() => {
    setBreadcrumbsFromItems([
      {
        label: "Settings",
        href: "/settings",
        isNavigatable: true,
      },
      {
        label: "Namespaces",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems]);

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Namespaces</h1>
        <p className="mt-2 text-gray-600">
          View and manage all namespaces you have access to across
          organizations.
        </p>
      </div>

      <AllNamespacesList
        namespaces={allNamespaces}
        isLoading={isLoading}
        error={error}
      />
    </div>
  );
}
