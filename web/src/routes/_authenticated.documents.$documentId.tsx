import { createFileRoute, notFound } from "@tanstack/react-router";

import { DocumentPage } from "@/components/documents";
import { getDocumentBody } from "@/lib/mock-data";

export const Route = createFileRoute("/_authenticated/documents/$documentId")({
  loader: ({ params }) => {
    const document = getDocumentBody(params.documentId);
    if (!document) {
      throw notFound();
    }
    return { document };
  },
  staticData: {
    breadcrumb: (data) =>
      data &&
      typeof data === "object" &&
      "document" in data &&
      data.document &&
      typeof data.document === "object" &&
      "title" in data.document
        ? String(data.document.title)
        : "Document",
  },
  component: DocumentRoute,
});

function DocumentRoute() {
  const { documentId } = Route.useParams();
  return <DocumentPage documentId={documentId} />;
}
