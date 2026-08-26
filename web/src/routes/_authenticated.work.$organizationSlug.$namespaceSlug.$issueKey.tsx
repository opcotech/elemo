import { createFileRoute } from "@tanstack/react-router";

import { NotFound } from "@/components/shared/not-found";
import { WorkItemPage } from "@/components/work/work-item-page";
import { WorkItemPageSkeleton } from "@/components/work/work-surface-skeletons";
import { loadWorkItemPage } from "@/lib/work/load-work-item";

export const Route = createFileRoute(
  "/_authenticated/work/$organizationSlug/$namespaceSlug/$issueKey"
)({
  pendingComponent: WorkItemPageSkeleton,
  notFoundComponent: NotFound,
  loader: async ({ context, params }) =>
    loadWorkItemPage(
      context.queryClient,
      params.organizationSlug,
      params.namespaceSlug,
      params.issueKey
    ),
  staticData: {
    breadcrumb: (data) =>
      data &&
      typeof data === "object" &&
      "item" in data &&
      data.item &&
      typeof data.item === "object" &&
      "key" in data.item
        ? String(data.item.key)
        : "Work",
  },
  component: WorkItemRoute,
});

function WorkItemRoute() {
  const data = Route.useLoaderData();

  return (
    <WorkItemPage
      item={data.item}
      issue={data.issue}
      organizationId={data.organizationId}
      organizationSlug={data.organizationSlug}
      namespaceId={data.namespaceId}
      namespaceSlug={data.namespaceSlug}
      issueKey={data.issueKey}
    />
  );
}
