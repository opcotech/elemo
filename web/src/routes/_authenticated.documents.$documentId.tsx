import { createFileRoute } from "@tanstack/react-router";

import { DocumentPage, DocumentPageSkeleton } from "@/components/documents";
import { entityBreadcrumb } from "@/lib/breadcrumb";
import { loadDocumentPage } from "@/lib/documents/load-document";

export const Route = createFileRoute("/_authenticated/documents/$documentId")({
  pendingComponent: DocumentPageSkeleton,
  loader: async ({ context, params }) =>
    loadDocumentPage(context.queryClient, params.documentId),
  staticData: {
    breadcrumb: (data) =>
      entityBreadcrumb(data, "document", "Document", "title"),
  },
  component: DocumentRoute,
});

function DocumentRoute() {
  const { document } = Route.useLoaderData();
  return <DocumentPage initialDocument={document} />;
}
