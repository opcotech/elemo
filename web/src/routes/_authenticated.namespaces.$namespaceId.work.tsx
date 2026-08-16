import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { WorkSurface } from "@/components/work/work-surface";
import { workRouteSearchSchema } from "@/lib/work-route-search";

const namespaceRoute = getRouteApi("/_authenticated/namespaces/$namespaceId");

export const Route = createFileRoute(
  "/_authenticated/namespaces/$namespaceId/work"
)({
  staticData: { breadcrumb: "Work" },
  validateSearch: workRouteSearchSchema,
  component: NamespaceWorkRoute,
});

function NamespaceWorkRoute() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { namespace } = namespaceRoute.useLoaderData();

  return (
    <WorkSurface
      title={`${namespace.name} / Work`}
      description="Namespace-scoped work using the shared projection query."
      context={{ namespace: namespace.name }}
      scope={{ type: "namespace", namespaceId: namespace.id }}
      search={search}
      onSearchChange={(patch) =>
        void navigate({
          search: (previous) => ({ ...previous, ...patch }),
          replace: true,
        })
      }
    />
  );
}
