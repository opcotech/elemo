import { createFileRoute } from "@tanstack/react-router";

import { PluginRouteOutlet } from "@/components/plugins/plugin-route-outlet";
import { loadOrganization } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";
import { requireOrganizationSlug } from "@/lib/route-identity";

export const Route = createFileRoute(
  "/_authenticated/organizations/$organizationSlug/plugins/$pluginId/$"
)({
  ssr: false,
  loader: ({ context, params }) =>
    withRouteErrors(async () => {
      requireOrganizationSlug(params.organizationSlug);
      const organization = await loadOrganization(
        context.queryClient,
        params.organizationSlug
      );
      return {
        organization,
        organizationId: organization.id,
        organizationSlug: params.organizationSlug,
        pluginId: params.pluginId,
        splat: params._splat ?? "",
      };
    }),
  staticData: {
    breadcrumb: "Plugin",
  },
  component: OrganizationPluginPage,
});

function OrganizationPluginPage() {
  const { pluginId, splat, organizationSlug, organizationId } =
    Route.useLoaderData();

  return (
    <div className="p-6">
      <PluginRouteOutlet
        pluginId={pluginId}
        splat={splat}
        context={{ organizationSlug, organizationId, pluginId }}
      />
    </div>
  );
}
