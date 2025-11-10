import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { DialogForm } from "@/components/ui/dialog-form";
import { EntitySelect } from "@/components/ui/entity-select";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Skeleton } from "@/components/ui/skeleton";
import {
  v1OrganizationMembersGetOptions,
  v1OrganizationRoleMembersAddMutation,
  v1OrganizationRoleMembersGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
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
  onSuccess?: () => void;
}

export function RoleMemberAddDialog({
  organizationId,
  roleId,
  open,
  onOpenChange,
  onSuccess,
}: RoleMemberAddDialogProps) {
  const queryClient = useQueryClient();
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

  const mutation = useMutation(v1OrganizationRoleMembersAddMutation());

  useEffect(() => {
    if (open) {
      // Clear error when dialog opens
      setError(null);
    }
  }, [open]);

  const onSubmit = (values: MemberFormValues) => {
    // Clear previous error when submitting again
    setError(null);

    mutation.mutate(
      {
        path: {
          id: organizationId,
          role_id: roleId,
        },
        body: {
          user_id: values.userId,
        },
      },
      {
        onSuccess: () => {
          setError(null);
          showSuccessToast("Member added", "Member added to role successfully");
          form.reset();
          onOpenChange(false);

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
        },
        onError: (err) => {
          setError(err as Error);
          showErrorToast("Failed to add member", err.message);
        },
      }
    );
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
        <FormField
          control={form.control}
          name="userId"
          render={({ field }) => {
            return (
              <FormItem>
                <FormLabel>Select Member</FormLabel>
                <FormControl>
                  <EntitySelect
                    options={memberOptions}
                    value={field.value}
                    onValueChange={field.onChange}
                    placeholder="Choose a member to add"
                    disabled={mutation.isPending}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            );
          }}
        />
      )}
    </DialogForm>
  );
}
