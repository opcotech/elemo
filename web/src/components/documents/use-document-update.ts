import { useMutation, useQueryClient } from "@tanstack/react-query";

import { v1DocumentUpdateMutation } from "@/lib/api/mutation-options";
import { v1DocumentGetOptions } from "@/lib/api/query-options";
import type { Document, DocumentPatch } from "@/lib/api/types";
import { invalidateDocumentQueries } from "@/lib/documents/document-queries";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

export function useDocumentUpdate(documentId: string) {
  const queryClient = useQueryClient();
  const mutation = useMutation(v1DocumentUpdateMutation());

  const updateDocument = async (
    patch: DocumentPatch,
    successDescription?: string
  ) => {
    try {
      const data = await mutation.mutateAsync({
        path: { id: documentId },
        body: patch,
      });
      queryClient.setQueryData<Document>(
        v1DocumentGetOptions({
          path: { id: documentId },
        }).queryKey,
        data
      );
      await invalidateDocumentQueries(queryClient, documentId);
      showSuccessToast(
        "Document updated",
        successDescription ?? "Your changes were saved"
      );
      return data;
    } catch (error) {
      showErrorToast(
        "Failed to update document",
        error instanceof Error ? error.message : "Unknown error occurred"
      );
      throw error;
    }
  };

  return {
    updateDocument,
    isPending: mutation.isPending,
  };
}
