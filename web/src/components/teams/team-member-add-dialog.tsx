import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";
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
import { useFormMutation } from "@/hooks/use-form-mutation";
import { collectedListQuery, cursorPageQuery } from "@/lib/api/cursor-pages";
import {
  v1OrganizationMembersGetOptions,
  v1OrganizationTeamMembersGetOptions,
} from "@/lib/api/query-options";
import {
  v1OrganizationMembersGet,
  v1OrganizationTeamMembersAdd,
  v1OrganizationTeamMembersGet,
} from "@/lib/api/sdk";
import type { OrganizationMember, User } from "@/lib/api/types";
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

  const mutation = useFormMutation({
    mutationFn: async (variables: {
      path: { id: string; team_id: string };
      body: { user_id: string };
    }) => {
      const { data } = await v1OrganizationTeamMembersAdd({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Member added",
    successDescription: "Member added to team successfully",
    errorMessagePrefix: "Failed to add member",
    queryKeysToInvalidate: [
      v1OrganizationTeamMembersGetOptions({
        path: {
          id: organizationId,
          team_id: teamId,
        },
      }).queryKey,
    ],
    resetFormOnSuccess: true,
    transformValues: (values) => ({
      path: {
        id: organizationId,
        team_id: teamId,
      },
      body: {
        user_id: values.userId,
      },
    }),
    onSuccess: async () => {
      onOpenChange(false);
      await onSuccess?.();
    },
  });

  useEffect(() => {
    if (open) {
      form.reset({ userId: "" });
    }
  }, [open, form]);

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
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error}
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
