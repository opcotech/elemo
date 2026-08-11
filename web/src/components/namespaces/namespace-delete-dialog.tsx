import { EntityDeleteDialog } from "@/components/entity-lifecycle/entity-lifecycle";
import { namespaceLifecycleConfig } from "@/components/entity-lifecycle/entity-lifecycle-configs";
import { v1NamespaceDeleteMutation } from "@/lib/api/mutation-options";
import type { Namespace } from "@/lib/api/types";

interface NamespaceDeleteDialogProps {
  namespace: Namespace;
  organizationId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  navigateOnSuccess?: boolean;
}

export function NamespaceDeleteDialog({
  namespace,
  organizationId,
  open,
  onOpenChange,
  onSuccess,
  navigateOnSuccess = false,
}: NamespaceDeleteDialogProps) {
  return (
    <EntityDeleteDialog
      entity={namespace}
      context={{ organizationId }}
      config={namespaceLifecycleConfig}
      mutationOptions={v1NamespaceDeleteMutation()}
      open={open}
      onOpenChange={onOpenChange}
      onSuccess={onSuccess}
      navigateOnSuccess={navigateOnSuccess}
    />
  );
}
