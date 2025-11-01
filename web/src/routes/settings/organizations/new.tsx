import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import {
  OrganizationCreateForm,
  OrganizationDetailHeader,
} from "@/components/organizations";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { ResourceType } from "@/hooks/use-permissions";
import { requirePermissionBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/settings/organizations/new")({
  beforeLoad: requirePermissionBeforeLoad({
    resourceType: ResourceType.Organization,
    permissionKind: "create",
  }),
  component: OrganizationCreatePage,
});

function OrganizationCreatePage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();

  useEffect(() => {
    setBreadcrumbsFromItems([
      {
        label: "Settings",
        href: "/settings",
        isNavigatable: true,
      },
      {
        label: "Organizations",
        href: "/settings/organizations",
        isNavigatable: true,
      },
      {
        label: "Create Organization",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems]);

  return (
    <div className="space-y-6">
      <OrganizationDetailHeader
        title="Create Organization"
        description="Create a new organization to manage your team and projects."
      />

      <OrganizationCreateForm />
    </div>
  );
}
