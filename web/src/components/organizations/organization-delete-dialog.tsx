import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
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
import type { Organization } from "@/lib/api";
import {
  v1OrganizationDeleteMutation,
  v1OrganizationGetOptions,
  v1OrganizationsGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

interface OrganizationDeleteDialogProps {
  organization: Organization;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function OrganizationDeleteDialog({
  organization,
  open,
  onOpenChange,
  onSuccess,
}: OrganizationDeleteDialogProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const mutation = useMutation(v1OrganizationDeleteMutation());

  const handleDelete = () => {
    mutation.mutate(
      {
        path: { id: organization.id },
        query: { force: false }, // Soft delete (default)
      },
      {
        onSuccess: () => {
          showSuccessToast(
            "Organization deleted",
            "The organization has been deleted successfully"
          );

          // Invalidate queries to refresh the list and detail views
          queryClient.invalidateQueries({
            queryKey: v1OrganizationGetOptions({
              path: { id: organization.id },
            }).queryKey,
          });
          queryClient.invalidateQueries({
            queryKey: v1OrganizationsGetOptions().queryKey,
          });

          onSuccess?.();
          onOpenChange(false);

          // Redirect to organizations list
          navigate({ to: "/settings/organizations" });
        },
        onError: (error) => {
          showErrorToast("Failed to delete organization", error.message);
        },
      }
    );
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>
            Are you sure you want to delete {organization.name}?
          </AlertDialogTitle>
          <AlertDialogDescription className="space-y-2">
            <p>
              This will mark the organization as deleted. This action cannot be
              undone.
            </p>
            <p className="font-medium">What will happen:</p>
            <ul className="list-disc list-inside space-y-1 text-sm">
              <li>The organization will be marked as deleted</li>
              <li>All organization members will lose access</li>
              <li>Organization data will be hidden from listings</li>
              <li>You will be redirected to the organizations list</li>
            </ul>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={mutation.isPending}>Cancel</AlertDialogCancel>
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

