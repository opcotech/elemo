import type { QueryKey } from "@tanstack/react-query";
import { UserMinus } from "lucide-react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { DeleteConfirmationDialog } from "@/components/ui/delete-confirmation-dialog";
import { useDeleteMutation } from "@/hooks/use-delete-mutation";
import { v1OrganizationTeamMemberRemoveMutation } from "@/lib/api/mutation-options";
import { v1OrganizationTeamMembersGetOptions } from "@/lib/api/query-options";
import type { User } from "@/lib/api/types";
import { getInitials } from "@/lib/utils";

interface TeamMemberRemoveDialogProps {
  member: User;
  teamName: string;
  organizationId: string;
  teamId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function TeamMemberRemoveDialog({
  member,
  teamName,
  organizationId,
  teamId,
  open,
  onOpenChange,
  onSuccess,
}: TeamMemberRemoveDialogProps) {
  const queryKeysToInvalidate: QueryKey[] = [
    v1OrganizationTeamMembersGetOptions({
      path: {
        id: organizationId,
        team_id: teamId,
      },
    }).queryKey,
  ];

  const deleteMutation = useDeleteMutation({
    mutationOptions: v1OrganizationTeamMemberRemoveMutation(),
    successMessage: "Member removed",
    successDescription: "Member removed from team successfully",
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
        team_id: teamId,
        user_id: member.id,
      },
    });
  };

  const fullName = `${member.first_name} ${member.last_name}`;

  return (
    <DeleteConfirmationDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Remove ${fullName} from ${teamName}?`}
      description={`Are you sure you want to remove ${fullName} from the ${teamName} team?`}
      consequences={[
        "The member will no longer inherit grants held by this team",
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
