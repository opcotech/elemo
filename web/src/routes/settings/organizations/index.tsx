import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import { OrganizationList } from "@/components/organizations";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/settings/organizations/")({
  beforeLoad: requireAuthBeforeLoad,
  component: OrganizationsPage,
});

function OrganizationsPage() {
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
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems]);

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Organizations</h1>
        <p className="mt-2 text-gray-600">View and manage organizations.</p>
      </div>

      <OrganizationList />
    </div>
  );
}
