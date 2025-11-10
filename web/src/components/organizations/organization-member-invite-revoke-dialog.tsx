import type { QueryKey } from "@tanstack/react-query";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import type { OrganizationMember } from "@/lib/api";
import {
  v1OrganizationMemberInviteRevokeMutation,
  v1OrganizationMembersGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { getInitials } from "@/lib/utils";

interface OrganizationMemberInviteRevokeDialogProps {
  member: OrganizationMember;
  organizationId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function OrganizationMemberInviteRevokeDialog({
  member,
  organizationId,
  open,
  onOpenChange,
  onSuccess,
}: OrganizationMemberInviteRevokeDialogProps) {
  const queryKeysToInvalidate: QueryKey[] = [
    v1OrganizationMembersGetOptions({
      path: { id: organizationId },
    }).queryKey,
  ];

  const deleteMutation = useDeleteMutation({
    mutationOptions: v1OrganizationMemberInviteRevokeMutation(),
    successMessage: "Invitation revoked",
    successDescription: "Invitation has been revoked successfully",
    errorMessagePrefix: "Failed to revoke invitation",
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
        user_id: member.id,
      },
    });
  };

  const fullName = `${member.first_name} ${member.last_name}`;

  return (
    <DeleteConfirmationDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Revoke Invitation?"
      description="Are you sure you want to revoke this invitation?"
      consequences={[
        "The invitation link will no longer be valid",
        "The user will not be able to join using this invitation",
        "You can send a new invitation if needed",
      ]}
      deleteButtonText="Revoke Invitation"
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
