import type { QueryKey } from "@tanstack/react-query";
import { UserMinus } from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import type { User } from "@/lib/api";
import {
  v1OrganizationRoleMemberRemoveMutation,
  v1OrganizationRoleMembersGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { getInitials } from "@/lib/utils";

interface RoleMemberRemoveDialogProps {
  member: User;
  roleName: string;
  organizationId: string;
  roleId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
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
  const queryKeysToInvalidate: QueryKey[] = [
    v1OrganizationRoleMembersGetOptions({
      path: {
        id: organizationId,
        role_id: roleId,
      },
    }).queryKey,
  ];

  const deleteMutation = useDeleteMutation({
    mutationOptions: v1OrganizationRoleMemberRemoveMutation(),
    successMessage: "Member removed",
    successDescription: "Member removed from role successfully",
    errorMessagePrefix: "Failed to remove member",
    queryKeysToInvalidate,
    onSuccess: () => {
      onSuccess?.();
      onOpenChange(false);
    },
  });

  const handleConfirm = () => {
    deleteMutation.mutate({
      path: {
        id: organizationId,
        role_id: roleId,
        user_id: member.id,
      },
    });
  };

  const fullName = `${member.first_name} ${member.last_name}`;

  return (
    <DeleteConfirmationDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Remove ${fullName} from ${roleName}?`}
      description={`Are you sure you want to remove ${fullName} from the ${roleName} role?`}
      consequences={[
        "The member will lose all permissions assigned to this role",
        "The member will lose access to all resources assigned to this role",
        "This action cannot be undone",
      ]}
      deleteButtonIcon={UserMinus}
      deleteButtonText="Remove Member"
      onConfirm={handleConfirm}
      isPending={deleteMutation.isPending}
    >
      <div className="bg-primary/5 ring-primary/10 mt-2 rounded-md p-3 text-sm ring-1">
        <div className="flex items-center gap-3">
          <Avatar className="h-10 w-10">
            <AvatarImage src={member.picture || undefined} alt={fullName} />
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
      </div>
    </DeleteConfirmationDialog>
  );
}
