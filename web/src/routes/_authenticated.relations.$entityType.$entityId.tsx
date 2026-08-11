import { createFileRoute } from "@tanstack/react-router";

import { RelationshipExplorerPage } from "@/components/pages/relationship-explorer-page";

export const Route = createFileRoute(
  "/_authenticated/relations/$entityType/$entityId"
)({
  staticData: { breadcrumb: "Relations" },
  component: RelationshipExplorerRoute,
});

function RelationshipExplorerRoute() {
  const { entityType, entityId } = Route.useParams();
  return (
    <RelationshipExplorerPage entityType={entityType} entityId={entityId} />
  );
}
