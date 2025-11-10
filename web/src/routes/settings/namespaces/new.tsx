import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";

import { NamespaceCreateForm } from "@/components/namespaces";
import { PageHeader } from "@/components/page-header";
import { useBreadcrumbUtils } from "@/hooks/use-breadcrumbs";
import { requireAuthBeforeLoad } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/settings/namespaces/new")({
  beforeLoad: requireAuthBeforeLoad,
  component: NamespaceCreatePage,
});

function NamespaceCreatePage() {
  const { setBreadcrumbsFromItems } = useBreadcrumbUtils();

  useEffect(() => {
    setBreadcrumbsFromItems([
      {
        label: "Settings",
        href: "/settings",
        isNavigatable: true,
      },
      {
        label: "Namespaces",
        href: "/settings/namespaces",
        isNavigatable: true,
      },
      {
        label: "Create Namespace",
        isNavigatable: false,
      },
    ]);
  }, [setBreadcrumbsFromItems]);

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
