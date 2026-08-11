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
import { v1OrganizationRoleMembersAddMutation } from "@/lib/api/mutation-options";
import {
  v1OrganizationMembersGetOptions,
  v1OrganizationRoleMembersGetOptions,
} from "@/lib/api/query-options";
import { runMutationSuccessWorkflow } from "@/lib/mutation-workflow";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import { getInitials } from "@/lib/utils";

const memberFormSchema = z.object({
  userId: z.string().min(1, "User is required"),
});

type MemberFormValues = z.infer<typeof memberFormSchema>;

interface RoleMemberAddDialogProps {
  organizationId: string;
  roleId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void | Promise<void>;
}

export function RoleMemberAddDialog({
  organizationId,
  roleId,
  open,
  onOpenChange,
  onSuccess,
}: RoleMemberAddDialogProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const [error, setError] = useState<Error | null>(null);

  const form = useForm<MemberFormValues>({
    resolver: zodResolver(memberFormSchema),
    defaultValues: {
      userId: "",
    },
  });

  const { data: organizationMembers, isLoading: isLoadingMembers } = useQuery(
    v1OrganizationMembersGetOptions({
      path: { id: organizationId },
    })
  );

  const { data: roleMembers, isLoading: isLoadingRoleMembers } = useQuery(
    v1OrganizationRoleMembersGetOptions({
      path: {
        id: organizationId,
        role_id: roleId,
      },
    })
  );

  const roleMembersQueryKey = v1OrganizationRoleMembersGetOptions({
    path: {
      id: organizationId,
      role_id: roleId,
    },
  }).queryKey;
  const mutation = useMutation({
    ...v1OrganizationRoleMembersAddMutation(),
    onSuccess: () =>
      runMutationSuccessWorkflow({
        invalidateQueries: [
          () =>
            queryClient.invalidateQueries({
              queryKey: roleMembersQueryKey,
            }),
        ],
        invalidateRouter: () => router.invalidate(),
        callbacks: [
          async () => {
            setError(null);
            showSuccessToast(
              "Member added",
              "Member added to role successfully"
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
      // Clear error when dialog opens
      setError(null);
    }
  }, [open]);

  const onSubmit = (values: MemberFormValues) => {
    // Clear previous error when submitting again
    setError(null);

    mutation.mutate({
      path: {
        id: organizationId,
        role_id: roleId,
      },
      body: {
        user_id: values.userId,
      },
    });
  };

  // Filter out members who are already in the role
  const availableMembers =
    organizationMembers && roleMembers
      ? organizationMembers.filter(
          (member) =>
            !roleMembers.some((roleMember) => roleMember.id === member.id)
        )
      : [];

  const memberOptions = availableMembers.map((member) => ({
    value: member.id,
    title: `${member.first_name} ${member.last_name}`,
    description: member.email,
    avatarSrc: member.picture || null,
    avatarFallback: getInitials(member.first_name, member.last_name),
  }));

  const isLoading = isLoadingMembers || isLoadingRoleMembers;

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Add Member to Role"
      onSubmit={form.handleSubmit(onSubmit)}
      isPending={mutation.isPending}
      error={error}
      submitButtonText="Add Member"
      onReset={() => form.reset()}
      className="sm:max-w-[500px]"
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
            {roleMembers && roleMembers.length > 0
              ? "All organization members are already assigned to this role."
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
