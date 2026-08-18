import type { QueryClient } from "@tanstack/react-query";
import { notFound } from "@tanstack/react-router";

import { isApiError, isNotFound, isPermissionDenied } from "@/lib/api/errors";
import { v1DocumentGetOptions } from "@/lib/api/query-options";
import type { Document } from "@/lib/api/types";

export type DocumentLoaderData = {
  document: Document;
};

function isClientRequestError(error: unknown): boolean {
  if (isNotFound(error) || isPermissionDenied(error)) {
    return true;
  }
  if (isApiError(error)) {
    return error.status >= 400 && error.status < 500;
  }
  if (
    error &&
    typeof error === "object" &&
    "status" in error &&
    typeof error.status === "number"
  ) {
    return error.status >= 400 && error.status < 500;
  }
  return false;
}

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
    if (isClientRequestError(error)) {
      throw notFound();
    }
    throw error;
  }
}
