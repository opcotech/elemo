import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Trash2 } from "lucide-react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import type { Role } from "@/lib/api";
import {
  v1OrganizationRoleDeleteMutation,
  v1OrganizationRoleGetOptions,
  v1OrganizationRolesGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

interface OrganizationRoleDeleteDialogProps {
  role: Role;
  organizationId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function OrganizationRoleDeleteDialog({
  role,
  organizationId,
  open,
  onOpenChange,
  onSuccess,
}: OrganizationRoleDeleteDialogProps) {
  const queryClient = useQueryClient();

  const mutation = useMutation(v1OrganizationRoleDeleteMutation());

  const handleDelete = () => {
    mutation.mutate(
      {
        path: {
          id: organizationId,
          role_id: role.id,
        },
      },
      {
        onSuccess: () => {
          showSuccessToast(
            "Role deleted",
            "The role has been deleted successfully"
          );

          // Invalidate queries to refresh the list
          queryClient.invalidateQueries({
            queryKey: v1OrganizationRolesGetOptions({
              path: { id: organizationId },
            }).queryKey,
          });
          queryClient.invalidateQueries({
            queryKey: v1OrganizationRoleGetOptions({
              path: {
                id: organizationId,
                role_id: role.id,
              },
            }).queryKey,
          });

          onSuccess?.();
          onOpenChange(false);
        },
        onError: (error) => {
          showErrorToast("Failed to delete role", error.message);
        },
      }
    );
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            Are you sure you want to delete {role.name}?
          </AlertDialogTitle>
          <AlertDialogDescription className="space-y-2">
            <p>
              This will permanently delete the role. This action cannot be
              undone.
            </p>
            <p className="font-medium">What will happen:</p>
            <ul className="list-inside list-disc space-y-1 text-sm">
              <li>The role will be permanently deleted</li>
              <li>
                All members assigned to this role will lose their role
                assignment
              </li>
              <li>Role permissions will be removed</li>
            </ul>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={mutation.isPending}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            onClick={handleDelete}
            disabled={mutation.isPending}
          >
            {mutation.isPending ? (
              <>
                <span>Deleting...</span>
              </>
            ) : (
              <>
                <Trash2 className="size-4" />
                Delete
              </>
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
