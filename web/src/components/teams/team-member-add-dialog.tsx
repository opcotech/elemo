import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { DialogForm } from "@/components/ui/dialog-form";
import { EntitySelect } from "@/components/ui/entity-select";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Skeleton } from "@/components/ui/skeleton";
import { collectedListQuery, cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1OrganizationTeamMembersAddMutation } from "@/lib/api/mutation-options";
import {
  v1OrganizationMembersGetOptions,
  v1OrganizationTeamMembersGetOptions,
} from "@/lib/api/query-options";
import {
  v1OrganizationMembersGet,
  v1OrganizationTeamMembersGet,
} from "@/lib/api/sdk";
import type { OrganizationMember, User } from "@/lib/api/types";
import { runMutationSuccessWorkflow } from "@/lib/mutation-workflow";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import { getInitials } from "@/lib/utils";

const memberFormSchema = z.object({
  userId: z.string().min(1, "User is required"),
});

type MemberFormValues = z.infer<typeof memberFormSchema>;

interface TeamMemberAddDialogProps {
  organizationId: string;
  teamId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void | Promise<void>;
}

export function TeamMemberAddDialog({
  organizationId,
  teamId,
  open,
  onOpenChange,
  onSuccess,
}: TeamMemberAddDialogProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const [error, setError] = useState<Error | null>(null);

  const form = useForm<MemberFormValues>({
    resolver: zodResolver(memberFormSchema),
    defaultValues: {
      userId: "",
    },
  });

  const membersOptions = v1OrganizationMembersGetOptions({
    path: { id: organizationId },
  });
  const { data: organizationMembersPage, isLoading: isLoadingMembers } =
    useQuery({
      ...collectedListQuery<OrganizationMember>(
        membersOptions,
        async (pageToken, signal) => {
          const { data } = await v1OrganizationMembersGet({
            path: { id: organizationId },
            query: cursorPageQuery(pageToken),
            signal,
            throwOnError: true,
          });
          return data;
        }
      ),
      enabled: open,
    });

  const teamMembersOptions = v1OrganizationTeamMembersGetOptions({
    path: {
      id: organizationId,
      team_id: teamId,
    },
  });
  const { data: teamMembersPage, isLoading: isLoadingTeamMembers } = useQuery({
    ...collectedListQuery<User>(
      teamMembersOptions,
      async (pageToken, signal) => {
        const { data } = await v1OrganizationTeamMembersGet({
          path: {
            id: organizationId,
            team_id: teamId,
          },
          query: cursorPageQuery(pageToken),
          signal,
          throwOnError: true,
        });
        return data;
      }
    ),
    enabled: open,
  });

  const teamMembersQueryKey = v1OrganizationTeamMembersGetOptions({
    path: {
      id: organizationId,
      team_id: teamId,
    },
  }).queryKey;
  const mutation = useMutation({
    ...v1OrganizationTeamMembersAddMutation(),
    onSuccess: () =>
      runMutationSuccessWorkflow({
        invalidateQueries: [
          () =>
            queryClient.invalidateQueries({
              queryKey: teamMembersQueryKey,
            }),
          () =>
            queryClient.invalidateQueries({
              queryKey: v1OrganizationTeamMembersGetOptions({
                path: { id: organizationId, team_id: teamId },
              }).queryKey,
            }),
        ],
        invalidateRouter: () => router.invalidate(),
        callbacks: [
          async () => {
            setError(null);
            showSuccessToast(
              "Member added",
              "Member added to team successfully"
            );
            form.reset();
            onOpenChange(false);
            await onSuccess?.();
          },
        ],
      }),
    onError: (err) => {
      setError(new Error(err.message));
      showErrorToast("Failed to add member", err.message);
    },
  });

  useEffect(() => {
    if (open) {
      setError(null);
    }
  }, [open]);

  const onSubmit = (values: MemberFormValues) => {
    setError(null);
    mutation.mutate({
      path: {
        id: organizationId,
        team_id: teamId,
      },
      body: {
        user_id: values.userId,
      },
    });
  };

  const organizationMembers: OrganizationMember[] =
    organizationMembersPage?.items ?? [];
  const teamMembers: User[] = teamMembersPage?.items ?? [];

  const availableMembers = organizationMembers.filter(
    (member) => !teamMembers.some((teamMember) => teamMember.id === member.id)
  );

  const memberOptions = availableMembers.map((member) => ({
    value: member.id,
    title: `${member.first_name} ${member.last_name}`,
    description: member.email,
    avatarSrc: member.picture || null,
    avatarFallback: getInitials(member.first_name, member.last_name),
  }));

  const isLoading = isLoadingMembers || isLoadingTeamMembers;

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Add Member to Team"
      onSubmit={form.handleSubmit(onSubmit)}
      isPending={mutation.isPending}
      error={error}
      submitButtonText="Add Member"
      onReset={() => form.reset()}
      className="sm:max-w-125"
    >
      {isLoading ? (
        <div className="space-y-4">
          <div className="space-y-2">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-10 w-full" />
          </div>
        </div>
      ) : availableMembers.length === 0 ? (
        <Alert>
          <AlertDescription>
            {teamMembers.length > 0
              ? "All organization members are already on this team."
              : "No organization members available to assign."}
          </AlertDescription>
        </Alert>
      ) : (
        <ControlledField
          control={form.control}
          name="userId"
          render={({ field }) => {
            return (
              <Field>
                <FieldLabel>Select Member</FieldLabel>
                <FieldControl>
                  <EntitySelect
                    options={memberOptions}
                    value={field.value}
                    onValueChange={field.onChange}
                    placeholder="Choose a member to add"
                    disabled={mutation.isPending}
                  />
                </FieldControl>
                <FieldError />
              </Field>
            );
          }}
        />
      )}
    </DialogForm>
  );
}
