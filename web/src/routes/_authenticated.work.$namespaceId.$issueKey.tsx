import { createFileRoute } from "@tanstack/react-router";

import { WorkItemPage } from "@/components/work/work-item-page";
import { WorkItemPageSkeleton } from "@/components/work/work-surface-skeletons";
import { loadWorkItemPage } from "@/lib/work/load-work-item";

export const Route = createFileRoute(
  "/_authenticated/work/$namespaceId/$issueKey"
)({
  pendingComponent: WorkItemPageSkeleton,
  loader: async ({ context, params }) =>
    loadWorkItemPage(context.queryClient, params.namespaceId, params.issueKey),
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

  if (data.source === "api") {
    return (
      <WorkItemPage
        item={data.item}
        issue={data.issue}
        namespaceId={data.namespaceId}
        issueKey={data.issueKey}
      />
    );
  }

  return <WorkItemPage item={data.item} />;
}
