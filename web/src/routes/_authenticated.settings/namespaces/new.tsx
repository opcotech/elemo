import { createFileRoute } from "@tanstack/react-router";

import { NamespaceCreateForm } from "@/components/namespaces";
import { PageHeader } from "@/components/shared/page-header";

export const Route = createFileRoute("/_authenticated/settings/namespaces/new")(
  {
    staticData: {
      breadcrumb: "Create namespace",
    },
    component: NamespaceCreatePage,
  }
);

function NamespaceCreatePage() {
  return (
    <div className="space-y-6">
      <PageHeader
        title="Create Namespace"
        description="Create a new namespace in an organization."
      />

      <NamespaceCreateForm />
    </div>
  );
}
