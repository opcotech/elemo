import { useMutation, useQueryClient } from "@tanstack/react-query";
import { UserMinus } from "lucide-react";

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
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import type { User } from "@/lib/api";
import {
  v1OrganizationRoleMemberRemoveMutation,
  v1OrganizationRoleMembersGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

interface RoleMemberRemoveDialogProps {
  member: User;
  roleName: string;
  organizationId: string;
  roleId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

function getInitials(firstName: string, lastName: string): string {
  return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase();
}

export function RoleMemberRemoveDialog({
  member,
  roleName,
  organizationId,
  roleId,
  open,
  onOpenChange,
  onSuccess,
}: RoleMemberRemoveDialogProps) {
  const queryClient = useQueryClient();

  const mutation = useMutation(v1OrganizationRoleMemberRemoveMutation());

  const handleRemove = () => {
    mutation.mutate(
      {
        path: {
          id: organizationId,
          role_id: roleId,
          user_id: member.id,
        },
      },
      {
        onSuccess: () => {
          showSuccessToast(
            "Member removed",
            "Member removed from role successfully"
          );

          // Invalidate queries to refresh the members list
          queryClient.invalidateQueries({
            queryKey: v1OrganizationRoleMembersGetOptions({
              path: {
                id: organizationId,
                role_id: roleId,
              },
            }).queryKey,
          });

          onSuccess?.();
          onOpenChange(false);
        },
        onError: (error) => {
          showErrorToast("Failed to remove member", error.message);
        },
      }
    );
  };

  const fullName = `${member.first_name} ${member.last_name}`;

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Remove Member from Role?</AlertDialogTitle>
          <AlertDialogDescription className="space-y-2">
            <p>Are you sure you want to remove this member from the role?</p>
            <div className="bg-muted rounded-md p-3 text-sm">
              <div className="font-medium">Member Details:</div>
              <div className="mt-2 flex items-center gap-3">
                <Avatar className="h-10 w-10">
                  <AvatarImage
                    src={member.picture || undefined}
                    alt={fullName}
                  />
                  <AvatarFallback>
                    {getInitials(member.first_name, member.last_name)}
                  </AvatarFallback>
                </Avatar>
                <div className="flex flex-col">
                  <span className="font-medium">{fullName}</span>
                  <span className="text-muted-foreground text-sm">
                    {member.email}
                  </span>
                </div>
              </div>
              <div className="mt-2">
                <span className="text-muted-foreground">Role: </span>
                <span className="font-medium">{roleName}</span>
              </div>
            </div>
            <p className="text-muted-foreground text-sm">
              This action cannot be undone. The member will lose all permissions
              associated with this role.
            </p>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={mutation.isPending}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            onClick={handleRemove}
            disabled={mutation.isPending}
          >
            {mutation.isPending ? (
              <>
                <span>Removing...</span>
              </>
            ) : (
              <>
                <UserMinus className="size-4" />
                Remove Member
              </>
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
