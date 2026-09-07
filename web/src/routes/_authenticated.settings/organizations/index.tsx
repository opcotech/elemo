import { createFileRoute } from "@tanstack/react-router";

import { OrganizationList } from "@/components/organizations/organization-list";
import { loadOrganizationsWithPermissions } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";

export const Route = createFileRoute("/_authenticated/settings/organizations/")(
  {
    loader: ({ context }) =>
      withRouteErrors(() =>
        loadOrganizationsWithPermissions(context.queryClient)
      ),
    staticData: {
      breadcrumb: "Organizations",
    },
    component: OrganizationsPage,
  }
);

function OrganizationsPage() {
  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="font-bold text-2xl">Organizations</h1>
        <p className="mt-2 text-muted-foreground">
          View and manage organizations.
        </p>
      </div>

      <OrganizationList />
    </div>
  );
}
