import { createFileRoute } from "@tanstack/react-router";

import { SearchPage } from "@/components/pages/search-page";
import { accessibleNamespacesOptions } from "@/lib/api/accessible-namespaces";
import { searchRouteSearchSchema } from "@/lib/work-route-search";

export const Route = createFileRoute("/_authenticated/search")({
  staticData: { breadcrumb: "Search" },
  validateSearch: searchRouteSearchSchema,
  loader: ({ context }) =>
    context.queryClient.fetchQuery(
      accessibleNamespacesOptions(context.queryClient)
    ),
  component: SearchRoute,
});

function SearchRoute() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  return (
    <SearchPage
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
