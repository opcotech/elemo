import { zodResolver } from "@hookform/resolvers/zod";
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
import { useFormMutation } from "@/hooks/use-form-mutation";
import type { Options, V1OrganizationsCreateData } from "@/lib/api";
import { v1OrganizationsCreate } from "@/lib/client/sdk.gen";
import { zOrganizationCreate } from "@/lib/client/zod.gen";
import { createFormSchema, normalizeFormData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

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

  const mutation = useFormMutation<
    { id: string },
    Options<V1OrganizationsCreateData>,
    OrganizationFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1OrganizationsCreate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Organization created",
    errorMessagePrefix: "Failed to create organization",
    transformValues: (values) => {
      const normalizedBody = normalizeFormData(
        organizationFormSchema,
        values
      ) as {
        name: string;
        email: string;
        website?: string;
      };
      return {
        body: normalizedBody,
      };
    },
    onSuccess: (data) => {
      navigate({
        to: "/settings/organizations/$organizationId",
        params: { organizationId: data.id },
      });
    },
  });

  return (
    <FormCard
      data-section="organization-create-form"
      description="Enter the details below to create a new organization."
      onSubmit={mutation.handleSubmit}
      onCancel={() => navigate({ to: "/settings/organizations" })}
      isPending={mutation.isPending}
      error={mutation.error || null}
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
                  value={getDefaultValue(field.value)}
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
