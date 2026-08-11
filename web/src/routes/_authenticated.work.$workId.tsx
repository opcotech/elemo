import { createFileRoute, notFound } from "@tanstack/react-router";

import { WorkItemPage } from "@/components/work";
import { getWorkItem } from "@/lib/mock-data";

export const Route = createFileRoute("/_authenticated/work/$workId")({
  loader: ({ params }) => {
    const item = getWorkItem(params.workId);
    if (!item) {
      throw notFound();
    }
    return { item };
  },
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
  const { workId } = Route.useParams();
  return <WorkItemPage workId={workId} />;
}
