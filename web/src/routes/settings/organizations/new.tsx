import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import { OrganizationCreateForm } from "@/components/organizations";
import { PageHeader } from "@/components/page-header";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { ResourceType } from "@/hooks/use-permissions";
import { useRequirePermission } from "@/hooks/use-require-permission";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/settings/organizations/new")({
  beforeLoad: requireAuthBeforeLoad,
  component: OrganizationCreatePage,
});

function OrganizationCreatePage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();

  const { isLoading: isCheckingPermission } = useRequirePermission({
    resourceType: ResourceType.Organization,
    permissionKind: "create",
  });

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

  // Show nothing while checking permissions (redirect will happen if denied)
  if (isCheckingPermission) {
    return null;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="Create Organization"
        description="Create a new organization to manage your team and projects."
      />

      <OrganizationCreateForm />
    </div>
  );
}
