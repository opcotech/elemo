import { createFileRoute } from "@tanstack/react-router";

import { AllNamespacesList } from "@/components/namespaces/all-namespaces-list";
import { loadAllNamespaces } from "@/lib/route-data";

export const Route = createFileRoute("/_authenticated/settings/namespaces/")({
  loader: ({ context }) => loadAllNamespaces(context.queryClient),
  staticData: {
    breadcrumb: "Namespaces",
  },
  component: NamespacesSettingsPage,
});

function NamespacesSettingsPage() {
  const { organizations, namespaces } = Route.useLoaderData();

  return (
    <div className="space-y-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Namespaces</h1>
        <p className="text-muted-foreground mt-2">
          View and manage all namespaces you have access to across
          organizations.
        </p>
      </div>

      <AllNamespacesList
        organizations={organizations}
        namespaces={namespaces}
        isLoading={false}
        error={null}
      />
    </div>
  );
}
