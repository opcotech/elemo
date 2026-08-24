import { createFileRoute } from "@tanstack/react-router";
import { useCallback } from "react";

import { SearchPage } from "@/components/pages/search-page";
import { accessibleNamespacesOptions } from "@/lib/api/accessible-namespaces";
import { v1SearchGetOptions } from "@/lib/api/query-options";
import { withRouteErrors } from "@/lib/route-errors";
import {
  hasActiveSearch,
  searchQueryFromRoute,
  searchRouteSearchSchema,
} from "@/lib/search/params";
import type { SearchRouteSearch } from "@/lib/search/params";

export const Route = createFileRoute("/_authenticated/search")({
  staticData: { breadcrumb: "Search" },
  validateSearch: searchRouteSearchSchema,
  loaderDeps: ({ search }) => search,
  loader: ({ context, deps }) =>
    withRouteErrors(async () => {
      await context.queryClient.fetchQuery(
        accessibleNamespacesOptions(context.queryClient)
      );
      if (!hasActiveSearch(deps)) {
        return;
      }
      await context.queryClient.ensureQueryData(
        v1SearchGetOptions({
          query: searchQueryFromRoute({ ...deps, page_token: undefined }),
        })
      );
    }),
  component: SearchRoute,
});

function SearchRoute() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const onSearchChange = useCallback(
    (patch: Partial<SearchRouteSearch>) => {
      void navigate({
        search: (previous) => ({ ...previous, ...patch }),
        replace: true,
      });
    },
    [navigate]
  );
  return <SearchPage search={search} onSearchChange={onSearchChange} />;
}
