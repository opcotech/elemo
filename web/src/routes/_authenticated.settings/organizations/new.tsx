import { createFileRoute } from "@tanstack/react-router";

import { OrganizationCreateForm } from "@/components/organizations/organization-create-form";
import { PageHeader } from "@/components/ui/page-header";
import { Action, ResourceType } from "@/lib/auth/permissions";
import { requirePermissionBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute(
  "/_authenticated/settings/organizations/new"
)({
  beforeLoad: ({ context }) =>
    requirePermissionBeforeLoad({
      queryClient: context.queryClient,
      resourceType: ResourceType.Installation,
      action: Action.OrganizationCreate,
    }),
  staticData: {
    breadcrumb: "Create organization",
  },
  component: OrganizationCreatePage,
});

function OrganizationCreatePage() {
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
