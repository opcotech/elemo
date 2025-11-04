import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import type { RoleCreate } from "@/lib/api";
import {
  v1OrganizationRolesCreateMutation,
  v1OrganizationRolesGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { zRoleCreate } from "@/lib/client/zod.gen";
import {
  createFormSchema,
  getFieldValue,
  normalizeFormData,
} from "@/lib/forms";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

const roleFormSchema = createFormSchema(zRoleCreate);

type RoleFormValues = z.infer<typeof roleFormSchema>;

const defaultValues: RoleFormValues = {
  name: "",
  description: "",
};

interface OrganizationRoleCreateFormProps {
  organizationId: string;
}

export function OrganizationRoleCreateForm({
  organizationId,
}: OrganizationRoleCreateFormProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const form = useForm<RoleFormValues>({
    resolver: zodResolver(roleFormSchema),
    defaultValues,
  });

  const mutation = useMutation(v1OrganizationRolesCreateMutation());

  const onSubmit = (values: RoleFormValues) => {
    const normalizedBody = normalizeFormData(
      roleFormSchema,
      values
    ) as RoleCreate;

    mutation.mutate(
      {
        path: {
          id: organizationId,
        },
        body: normalizedBody,
      },
      {
        onSuccess: () => {
          showSuccessToast("Role created", "Role created successfully");

          // Invalidate roles list query
          queryClient.invalidateQueries({
            queryKey: v1OrganizationRolesGetOptions({
              path: { id: organizationId },
            }).queryKey,
          });

          navigate({
            to: "/settings/organizations/$organizationId",
            params: { organizationId },
          });
        },
        onError: (error) => {
          showErrorToast("Failed to create role", error.message);
        },
      }
    );
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create Role</CardTitle>
        <CardDescription>
          Enter the details below to create a new role.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col gap-y-6"
          >
            {mutation.isError && (
              <Alert variant="destructive">
                <AlertTitle>Failed to create role</AlertTitle>
                <AlertDescription>{mutation.error.message}</AlertDescription>
              </Alert>
            )}

            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Enter role name" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Enter role description (optional)"
                      {...field}
                      value={getFieldValue(field.value)}
                      rows={4}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() =>
                  navigate({
                    to: "/settings/organizations/$organizationId",
                    params: { organizationId },
                  })
                }
                disabled={mutation.isPending}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={mutation.isPending}>
                {mutation.isPending ? (
                  <>
                    <Spinner size="xs" className="mr-0.5 text-white" />
                    <span>Creating...</span>
                  </>
                ) : (
                  "Create Role"
                )}
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
