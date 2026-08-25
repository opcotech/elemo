import { Outlet, createFileRoute } from "@tanstack/react-router";

import { OrganizationDetailError } from "@/components/organizations/organization-detail-error";
import { OrganizationNotFound } from "@/components/organizations/organization-not-found";
import { DetailPageSkeleton } from "@/components/ui/detail-card";
import { requireOrganizationSlug } from "@/lib/route-identity";

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug"
)({
  loader: ({ params }) => {
    requireOrganizationSlug(params.organizationSlug);
    return { organizationSlug: params.organizationSlug };
  },
  staticData: {
    breadcrumb: (data) =>
      data &&
      typeof data === "object" &&
      "organizationSlug" in data &&
      typeof data.organizationSlug === "string"
        ? data.organizationSlug
        : "Organization",
  },
  pendingComponent: DetailPageSkeleton,
  errorComponent: OrganizationDetailError,
  notFoundComponent: OrganizationNotFound,
  component: OrganizationLayoutRoute,
});

function OrganizationLayoutRoute() {
  return <Outlet />;
}
