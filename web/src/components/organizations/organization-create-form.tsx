import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldProvider,
} from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
import { Input } from "@/components/ui/input";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { accessibleNamespacesQueryKey } from "@/lib/api/accessible-namespaces";
import { v1OrganizationsGetOptions } from "@/lib/api/query-options";
import { v1OrganizationsCreate } from "@/lib/api/sdk";
import type { Options, V1OrganizationsCreateData } from "@/lib/api/types";
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
    queryKeysToInvalidate: [
      v1OrganizationsGetOptions().queryKey,
      accessibleNamespacesQueryKey,
    ],
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
      <FieldProvider {...form}>
        <FieldGroup>
          <ControlledField
            control={form.control}
            name="name"
            render={({ field }) => (
              <Field>
                <FieldLabel>Name</FieldLabel>
                <FieldControl>
                  <Input placeholder="Enter organization name" {...field} />
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />

          <ControlledField
            control={form.control}
            name="email"
            render={({ field }) => (
              <Field>
                <FieldLabel>Email</FieldLabel>
                <FieldControl>
                  <Input
                    type="email"
                    placeholder="Enter organization email"
                    {...field}
                  />
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />

          <ControlledField
            control={form.control}
            name="website"
            render={({ field }) => (
              <Field>
                <FieldLabel>Website</FieldLabel>
                <FieldControl>
                  <Input
                    type="url"
                    placeholder="https://example.com"
                    {...field}
                    value={getDefaultValue(field.value)}
                  />
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />
        </FieldGroup>
      </FieldProvider>
    </FormCard>
  );
}
