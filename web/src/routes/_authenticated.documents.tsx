import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/documents")({
  staticData: {
    breadcrumb: { label: "Documents", href: "/documents" },
  },
  component: DocumentsLayoutRoute,
});

function DocumentsLayoutRoute() {
  return <Outlet />;
}
