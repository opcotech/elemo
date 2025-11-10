import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { DialogForm } from "@/components/ui/dialog-form";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  v1OrganizationMembersGetOptions,
  v1OrganizationMembersInviteMutation,
  v1OrganizationRolesGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

const inviteFormSchema = z.object({
  email: z.string().email("Invalid email address"),
  role_id: z.string().optional(),
});

type InviteFormValues = z.infer<typeof inviteFormSchema>;

interface OrganizationMemberInviteDialogProps {
  organizationId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function OrganizationMemberInviteDialog({
  organizationId,
  open,
  onOpenChange,
  onSuccess,
}: OrganizationMemberInviteDialogProps) {
  const queryClient = useQueryClient();

  const { data: roles, isLoading: isLoadingRoles } = useQuery(
    v1OrganizationRolesGetOptions({
      path: { id: organizationId },
    })
  );

  const form = useForm<InviteFormValues>({
    resolver: zodResolver(inviteFormSchema),
    defaultValues: {
      email: "",
      role_id: undefined,
    },
  });

  const mutation = useMutation(v1OrganizationMembersInviteMutation());

  const onSubmit = (values: InviteFormValues) => {
    mutation.mutate(
      {
        path: {
          id: organizationId,
        },
        body: {
          email: values.email,
          role_id:
            values.role_id && values.role_id !== ""
              ? values.role_id
              : undefined,
        },
      },
      {
        onSuccess: () => {
          showSuccessToast(
            "Invitation sent",
            "Invitation email sent successfully"
          );
          form.reset();
          onOpenChange(false);

          // Invalidate queries to refresh the members list
          queryClient.invalidateQueries({
            queryKey: v1OrganizationMembersGetOptions({
              path: { id: organizationId },
            }).queryKey,
          });

          onSuccess?.();
        },
        onError: (error) => {
          showErrorToast("Failed to send invitation", error.message);
        },
      }
    );
  };

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Invite Member"
      onSubmit={form.handleSubmit(onSubmit)}
      isPending={mutation.isPending}
      error={mutation.error as Error | null}
      submitButtonText="Send Invitation"
      onReset={() => form.reset()}
      className="sm:max-w-[500px]"
    >
      <FormField
        control={form.control}
        name="email"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Email Address</FormLabel>
            <FormControl>
              <Input type="email" placeholder="user@example.com" {...field} />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />

      <FormField
        control={form.control}
        name="role_id"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Role (Optional)</FormLabel>
            <Select
              value={field.value || ""}
              onValueChange={(value) => {
                // Empty string means no role selected
                field.onChange(value === "" ? undefined : value);
              }}
            >
              <FormControl>
                <SelectTrigger>
                  <SelectValue placeholder="Select a role (optional)" />
                </SelectTrigger>
              </FormControl>
              <SelectContent>
                {isLoadingRoles ? (
                  <div className="px-2 py-1.5">
                    <Skeleton className="h-4 w-32" />
                  </div>
                ) : (
                  roles?.map((role) => (
                    <SelectItem key={role.id} value={role.id}>
                      {role.name}
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
            <FormMessage />
          </FormItem>
        )}
      />
    </DialogForm>
  );
}
