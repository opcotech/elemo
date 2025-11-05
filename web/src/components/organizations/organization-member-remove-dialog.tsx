import type { QueryKey } from "@tanstack/react-query";
import { UserMinus } from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import type { OrganizationMember } from "@/lib/api";
import {
  v1OrganizationMemberRemoveMutation,
  v1OrganizationMembersGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { getInitials } from "@/lib/utils";

interface OrganizationMemberRemoveDialogProps {
  member: OrganizationMember;
  organizationId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function OrganizationMemberRemoveDialog({
  member,
  organizationId,
  open,
  onOpenChange,
  onSuccess,
}: OrganizationMemberRemoveDialogProps) {
  const queryKeysToInvalidate: QueryKey[] = [
    v1OrganizationMembersGetOptions({
      path: { id: organizationId },
    }).queryKey,
  ];

  const deleteMutation = useDeleteMutation({
    mutationOptions: v1OrganizationMemberRemoveMutation(),
    successMessage: "Member removed",
    successDescription: "Member removed from organization successfully",
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
        user_id: member.id,
      },
    });
  };

  const fullName = `${member.first_name} ${member.last_name}`;

  return (
    <DeleteConfirmationDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Remove Member from Organization?"
      description="Are you sure you want to remove this member from the organization?"
      consequences={[
        "The member will lose access to all organization resources",
        "All roles assigned to this member will be removed",
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
