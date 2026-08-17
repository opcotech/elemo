import { useQueryClient } from "@tanstack/react-query";

import { EntityDeleteDialog } from "@/components/entity-lifecycle/entity-lifecycle";
import { documentLifecycleConfig } from "@/components/entity-lifecycle/entity-lifecycle-configs";
import { v1DocumentDeleteMutation } from "@/lib/api/mutation-options";
import type { Document } from "@/lib/api/types";
import { invalidateDocumentQueries } from "@/lib/documents/document-queries";
import { uiActions } from "@/lib/ui-store";

interface DocumentDeleteDialogProps {
  document: Pick<Document, "id" | "title">;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  navigateOnSuccess?: boolean;
}

export function DocumentDeleteDialog({
  document,
  open,
  onOpenChange,
  onSuccess,
  navigateOnSuccess = true,
}: DocumentDeleteDialogProps) {
  const queryClient = useQueryClient();

  return (
    <EntityDeleteDialog
      entity={document}
      context={undefined}
      config={documentLifecycleConfig}
      mutationOptions={v1DocumentDeleteMutation()}
      open={open}
      onOpenChange={onOpenChange}
      navigateOnSuccess={navigateOnSuccess}
      onSuccess={async () => {
        uiActions.forgetRecentEntity({ id: document.id, type: "document" });
        await invalidateDocumentQueries(queryClient, document.id);
        await onSuccess?.();
      }}
    />
  );
}
