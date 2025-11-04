import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { DialogForm } from "@/components/ui/dialog-form";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
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

  const form = useForm<MemberFormValues>({
    resolver: zodResolver(memberFormSchema),
    defaultValues: {
      userId: "",
    },
  });

  // Get organization members
  const { data: organizationMembers, isLoading: isLoadingMembers } = useQuery(
    v1OrganizationMembersGetOptions({
      path: { id: organizationId },
    })
  );

  // Get current role members to filter them out
  const { data: roleMembers, isLoading: isLoadingRoleMembers } = useQuery(
    v1OrganizationRoleMembersGetOptions({
      path: {
        id: organizationId,
        role_id: roleId,
      },
    })
  );

  const mutation = useMutation(v1OrganizationRoleMembersAddMutation());

  const onSubmit = (values: MemberFormValues) => {
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
        onError: (error) => {
          showErrorToast("Failed to add member", error.message);
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

  const isLoading = isLoadingMembers || isLoadingRoleMembers;

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Add Member to Role"
      onSubmit={form.handleSubmit(onSubmit)}
      isPending={mutation.isPending}
      error={mutation.error as Error | null}
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
            const selectedMember = availableMembers.find(
              (member) => member.id === field.value
            );
            return (
              <FormItem>
                <FormLabel>Select Member</FormLabel>
                <Select value={field.value} onValueChange={field.onChange}>
                  <FormControl>
                    <SelectTrigger className="w-full">
                      {selectedMember ? (
                        <div className="flex items-center gap-3">
                          <Avatar className="h-6 w-6">
                            <AvatarImage
                              src={selectedMember.picture || undefined}
                              alt={`${selectedMember.first_name} ${selectedMember.last_name}`}
                            />
                            <AvatarFallback>
                              {getInitials(
                                selectedMember.first_name,
                                selectedMember.last_name
                              )}
                            </AvatarFallback>
                          </Avatar>
                          <span className="font-medium">{`${selectedMember.first_name} ${selectedMember.last_name}`}</span>
                        </div>
                      ) : (
                        <span className="text-muted-foreground">
                          Choose a member to add
                        </span>
                      )}
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {availableMembers.map((member) => {
                      const fullName = `${member.first_name} ${member.last_name}`;
                      return (
                        <SelectItem
                          key={member.id}
                          value={member.id}
                          className="py-2"
                        >
                          <div className="flex items-center gap-3">
                            <Avatar className="h-8 w-8">
                              <AvatarImage
                                src={member.picture || undefined}
                                alt={fullName}
                              />
                              <AvatarFallback>
                                {getInitials(
                                  member.first_name,
                                  member.last_name
                                )}
                              </AvatarFallback>
                            </Avatar>
                            <div className="flex flex-col">
                              <span className="font-medium">{fullName}</span>
                              <span className="text-muted-foreground text-sm">
                                {member.email}
                              </span>
                            </div>
                          </div>
                        </SelectItem>
                      );
                    })}
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            );
          }}
        />
      )}
    </DialogForm>
  );
}
