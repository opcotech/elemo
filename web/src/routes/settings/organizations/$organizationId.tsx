import { Link, createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/settings/organizations/$organizationId")(
  {
    beforeLoad: requireAuthBeforeLoad,
    component: OrganizationDetailPage,
  }
);

function OrganizationDetailPage() {
  const { organizationId } = Route.useParams();
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
        label: "Organization Details",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems]);

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Organization Details</h1>
        <p className="mt-2 text-gray-600">
          Organization details page is not yet implemented.
        </p>
      </div>

      <div className="rounded-lg border p-8 text-center">
        <p className="text-muted-foreground mb-4">
          This page is under development. Organization ID: {organizationId}
        </p>
        <Link
          to="/settings/organizations"
          className="text-primary hover:underline"
        >
          ← Back to Organizations
        </Link>
      </div>
    </div>
  );
}
