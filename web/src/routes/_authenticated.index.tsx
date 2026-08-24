import { createFileRoute } from "@tanstack/react-router";

import { HomePage } from "@/components/pages/home-page";
import { accessibleNamespacesOptions } from "@/lib/api/accessible-namespaces";
import { v1TodosGetOptions } from "@/lib/api/query-options";
import { withRouteErrors } from "@/lib/route-errors";

export const Route = createFileRoute("/_authenticated/")({
  staticData: { breadcrumb: "Home" },
  loader: async ({ context }) => {
    await withRouteErrors(() =>
      Promise.all([
        context.queryClient.fetchQuery(
          accessibleNamespacesOptions(context.queryClient)
        ),
        context.queryClient.fetchQuery(v1TodosGetOptions()),
      ])
    );
  },
  component: HomePage,
});
