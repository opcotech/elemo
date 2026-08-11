import { createFileRoute } from "@tanstack/react-router";

import { OrganizationList } from "@/components/organizations";
import { loadOrganizationsWithPermissions } from "@/lib/route-data";

export const Route = createFileRoute("/_authenticated/settings/organizations/")(
  {
    loader: ({ context }) =>
      loadOrganizationsWithPermissions(context.queryClient),
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
        <h1 className="text-2xl font-bold">Organizations</h1>
        <p className="text-muted-foreground mt-2">
          View and manage organizations.
        </p>
      </div>

      <OrganizationList />
    </div>
  );
}
