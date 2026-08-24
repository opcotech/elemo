import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";

import { DocumentHubPage } from "@/components/documents/document-hub-page";
import {
  accessibleNamespacesOptions,
  useAccessibleNamespaces,
} from "@/lib/api/accessible-namespaces";
import { withRouteErrors } from "@/lib/route-errors";

const documentsHubSearchSchema = z.object({
  q: z.string().optional().catch(undefined),
});

export const Route = createFileRoute("/_authenticated/documents/")({
  validateSearch: documentsHubSearchSchema,
  loader: ({ context }) =>
    withRouteErrors(() =>
      context.queryClient.fetchQuery(
        accessibleNamespacesOptions(context.queryClient)
      )
    ),
  component: DocumentsHubRoute,
});

function DocumentsHubRoute() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { data: workspace, isLoading } = useAccessibleNamespaces();

  return (
    <DocumentHubPage
      workspace={workspace}
      isLoading={isLoading}
      query={search.q}
      onQueryChange={(query) =>
        void navigate({
          search: (previous) => ({
            ...previous,
            q: query,
          }),
          replace: true,
        })
      }
    />
  );
}
