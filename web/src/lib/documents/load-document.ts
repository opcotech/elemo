import type { QueryClient } from "@tanstack/react-query";
import { notFound } from "@tanstack/react-router";

import { isNotFoundOrForbidden } from "@/lib/api/errors";
import { v1DocumentGetOptions } from "@/lib/api/query-options";
import type { Document } from "@/lib/api/types";

export type DocumentLoaderData = {
  document: Document;
};

export async function loadDocumentPage(
  queryClient: QueryClient,
  documentId: string
): Promise<DocumentLoaderData> {
  try {
    const document = await queryClient.fetchQuery(
      v1DocumentGetOptions({
        path: { id: documentId },
      })
    );
    return { document };
  } catch (error) {
    if (isNotFoundOrForbidden(error)) {
      throw notFound();
    }
    throw error;
  }
}
