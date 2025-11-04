import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { FormCard } from "@/components/ui/form-card";
import { Input } from "@/components/ui/input";
import { v1OrganizationsCreateMutation } from "@/lib/client/@tanstack/react-query.gen";
import { zOrganizationCreate } from "@/lib/client/zod.gen";
import {
  createFormSchema,
  getFieldValue,
  normalizeFormData,
} from "@/lib/forms";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

// Create a schema without logo field for the form
// TODO: Add logo field back in when implementing image upload
const organizationFormSchema = createFormSchema(
  zOrganizationCreate.omit({ logo: true })
);

type OrganizationFormValues = z.infer<typeof organizationFormSchema>;

const defaultValues: OrganizationFormValues = {
  name: "",
  email: "",
  website: "",
};

export function OrganizationCreateForm() {
  const navigate = useNavigate();

  const form = useForm<OrganizationFormValues>({
    resolver: zodResolver(organizationFormSchema),
    defaultValues,
  });

  const mutation = useMutation(v1OrganizationsCreateMutation());

  const onSubmit = (values: OrganizationFormValues) => {
    // Normalize form data: converts empty strings to undefined for optional fields
    const normalizedBody = normalizeFormData(
      organizationFormSchema,
      values
    ) as {
      name: string;
      email: string;
      website?: string;
    };

    mutation.mutate(
      {
        body: normalizedBody,
      },
      {
        onSuccess: (data) => {
          showSuccessToast(
            "Organization created",
            "Organization created successfully"
          );
          navigate({
            to: "/settings/organizations/$organizationId",
            params: { organizationId: data.id },
          });
        },
        onError: (error) => {
          showErrorToast("Failed to create organization", error.message);
        },
      }
    );
  };

  return (
    <FormCard
      title="Create Organization"
      description="Enter the details below to create a new organization."
      onSubmit={form.handleSubmit(onSubmit)}
      onCancel={() => navigate({ to: "/settings/organizations" })}
      isPending={mutation.isPending}
      error={mutation.error}
      submitButtonText="Create Organization"
    >
      <Form {...form}>
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input placeholder="Enter organization name" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl>
                <Input
                  type="email"
                  placeholder="Enter organization email"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="website"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Website</FormLabel>
              <FormControl>
                <Input
                  type="url"
                  placeholder="https://example.com (optional)"
                  {...field}
                  value={getFieldValue(field.value)}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </Form>
    </FormCard>
  );
}
