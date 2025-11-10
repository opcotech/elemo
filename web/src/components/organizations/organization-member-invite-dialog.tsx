import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";
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
import { useFormMutation } from "@/hooks/use-form-mutation";
import type { Options, V1OrganizationMembersInviteData } from "@/lib/api";
import {
  v1OrganizationMembersGetOptions,
  v1OrganizationRolesGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { v1OrganizationMembersInvite } from "@/lib/client/sdk.gen";

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

  const mutation = useFormMutation<
    unknown,
    Options<V1OrganizationMembersInviteData>,
    InviteFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1OrganizationMembersInvite({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Invitation sent",
    successDescription: "Invitation email sent successfully",
    errorMessagePrefix: "Failed to send invitation",
    queryKeysToInvalidate: [
      v1OrganizationMembersGetOptions({
        path: { id: organizationId },
      }).queryKey,
    ],
    resetFormOnSuccess: true,
    transformValues: (values) => {
      return {
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
      };
    },
    onSuccess: () => {
      onOpenChange(false);
      onSuccess?.();
    },
  });

  useEffect(() => {
    if (open) {
      form.reset({
        email: "",
        role_id: undefined,
      });
    }
  }, [open, form]);

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Invite Member"
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error || null}
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
