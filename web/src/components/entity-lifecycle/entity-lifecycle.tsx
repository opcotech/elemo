import type { QueryKey, UseMutationOptions } from "@tanstack/react-query";
import type { useNavigate } from "@tanstack/react-router";
import { Trash2 } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  DangerZone,
  DangerZoneActions,
  DangerZoneContent,
  DangerZoneDescription,
  DangerZoneHeader,
  DangerZoneTitle,
} from "@/components/ui/danger-zone";
import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import type { Permission } from "@/lib/api/types";

type Navigate = ReturnType<typeof useNavigate>;

interface DeleteDialogConfig<TEntity> {
  title: (entity: TEntity) => string;
  description: string;
  consequences: readonly string[];
}

interface DangerZoneConfig {
  dataSection: string;
  description: string;
  summary: string;
  consequences: readonly string[];
  buttonLabel: string;
}

export interface EntityLifecycleConfig<TEntity, TContext, TVariables> {
  entityName: string;
  deleteDialog: DeleteDialogConfig<TEntity>;
  dangerZone?: DangerZoneConfig;
  canDelete: (entity: TEntity, permissions: Permission[]) => boolean;
  deleteVariables: (entity: TEntity, context: TContext) => TVariables;
  queryKeys: (entity: TEntity, context: TContext) => QueryKey[];
  navigateAfterDelete?: (
    navigate: Navigate,
    context: TContext
  ) => void | Promise<void>;
}

interface EntityDeleteDialogProps<
  TEntity,
  TContext,
  TData,
  TVariables,
  TError,
  TMutationContext,
> {
  entity: TEntity;
  context: TContext;
  config: EntityLifecycleConfig<TEntity, TContext, TVariables>;
  mutationOptions: UseMutationOptions<
    TData,
    TError,
    TVariables,
    TMutationContext
  >;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void | Promise<void>;
  navigateOnSuccess?: boolean;
}

export function EntityDeleteDialog<
  TEntity,
  TContext,
  TData,
  TVariables,
  TError,
  TMutationContext,
>({
  entity,
  context,
  config,
  mutationOptions,
  open,
  onOpenChange,
  onSuccess,
  navigateOnSuccess = false,
}: EntityDeleteDialogProps<
  TEntity,
  TContext,
  TData,
  TVariables,
  TError,
  TMutationContext
>) {
  const deleteMutation = useDeleteMutation({
    mutationOptions,
    successMessage: `${config.entityName} deleted`,
    successDescription: `The ${config.entityName.toLowerCase()} has been deleted successfully`,
    errorMessagePrefix: `Failed to delete ${config.entityName.toLowerCase()}`,
    queryKeysToInvalidate: config.queryKeys(entity, context),
    onSuccess: async () => {
      await onSuccess?.();
      onOpenChange(false);
    },
    navigateOnSuccess:
      navigateOnSuccess && config.navigateAfterDelete
        ? (navigate) => config.navigateAfterDelete?.(navigate, context)
        : undefined,
  });

  return (
    <DeleteConfirmationDialog
      open={open}
      onOpenChange={onOpenChange}
      title={config.deleteDialog.title(entity)}
      description={config.deleteDialog.description}
      consequences={[...config.deleteDialog.consequences]}
      deleteButtonText="Delete"
      onConfirm={() =>
        deleteMutation.mutate(config.deleteVariables(entity, context))
      }
      isPending={deleteMutation.isPending}
    />
  );
}

interface EntityDangerZoneProps<
  TEntity,
  TContext,
  TData,
  TVariables,
  TError,
  TMutationContext,
> {
  entity: TEntity;
  context: TContext;
  permissions: Permission[];
  config: EntityLifecycleConfig<TEntity, TContext, TVariables> & {
    dangerZone: DangerZoneConfig;
  };
  mutationOptions: UseMutationOptions<
    TData,
    TError,
    TVariables,
    TMutationContext
  >;
}

export function EntityDangerZone<
  TEntity,
  TContext,
  TData,
  TVariables,
  TError,
  TMutationContext,
>({
  entity,
  context,
  permissions,
  config,
  mutationOptions,
}: EntityDangerZoneProps<
  TEntity,
  TContext,
  TData,
  TVariables,
  TError,
  TMutationContext
>) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  if (!config.canDelete(entity, permissions)) {
    return null;
  }

  return (
    <>
      <DangerZone data-section={config.dangerZone.dataSection}>
        <DangerZoneHeader>
          <DangerZoneTitle>Danger Zone</DangerZoneTitle>
          <DangerZoneDescription>
            {config.dangerZone.description}
          </DangerZoneDescription>
        </DangerZoneHeader>
        <DangerZoneContent className="space-y-2">
          <p className="text-muted-foreground text-sm">
            {config.dangerZone.summary}
          </p>
          <p className="text-sm font-medium">Consequences:</p>
          <ul className="text-muted-foreground list-inside list-disc space-y-1 text-sm">
            {config.dangerZone.consequences.map((consequence) => (
              <li key={consequence}>{consequence}</li>
            ))}
          </ul>
        </DangerZoneContent>
        <DangerZoneActions>
          <Button
            variant="destructive"
            onClick={() => setDeleteDialogOpen(true)}
          >
            <Trash2 className="size-4" />
            {config.dangerZone.buttonLabel}
          </Button>
        </DangerZoneActions>
      </DangerZone>

      <EntityDeleteDialog
        entity={entity}
        context={context}
        config={config}
        mutationOptions={mutationOptions}
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        navigateOnSuccess
      />
    </>
  );
}
