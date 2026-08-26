import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import {
  ControlledField,
  Field,
  FieldControl,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldProvider,
} from "@/components/ui/field";
import { FormCard } from "@/components/ui/form-card";
import { Input } from "@/components/ui/input";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { useSuggestedSlug } from "@/hooks/use-suggested-slug";
import { accessibleNamespacesQueryKey } from "@/lib/api/accessible-namespaces";
import { isConflict, throwIfApiFailed } from "@/lib/api/errors";
import { v1OrganizationsGetOptions } from "@/lib/api/query-options";
import { zOrganizationCreate } from "@/lib/api/schemas";
import { v1OrganizationsCreate } from "@/lib/api/sdk";
import type { Options, V1OrganizationsCreateData } from "@/lib/api/types";
import { createFormSchema, normalizeFormData } from "@/lib/forms";
import { organizationSlugFormSchema } from "@/lib/slug";
import { getDefaultValue } from "@/lib/utils";

const organizationFormSchema = createFormSchema(
  zOrganizationCreate.omit({ logo: true, slug: true })
).extend({
  slug: organizationSlugFormSchema,
});

type OrganizationFormValues = z.infer<typeof organizationFormSchema>;

const defaultValues: OrganizationFormValues = {
  name: "",
  slug: "",
  email: "",
  website: "",
};

export function OrganizationCreateForm() {
  const navigate = useNavigate();

  const form = useForm<OrganizationFormValues>({
    resolver: zodResolver(organizationFormSchema),
    defaultValues,
  });
  useSuggestedSlug(form);

  const mutation = useFormMutation<
    { id: string },
    Options<V1OrganizationsCreateData>,
    OrganizationFormValues
  >({
    mutationFn: async (variables) => {
      return throwIfApiFailed(
        await v1OrganizationsCreate({
          ...variables,
          throwOnError: false,
        })
      );
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
        slug: string;
        name: string;
        email: string;
        website?: string;
      };
      return {
        body: normalizedBody,
      };
    },
    onError: (error) => {
      if (isConflict(error)) {
        form.setError("slug", {
          type: "conflict",
          message: "This slug is already in use",
        });
      }
    },
    navigateOnSuccess: (navigateTo) =>
      navigateTo({
        to: "/settings/organizations/$organizationSlug",
        params: { organizationSlug: form.getValues("slug") },
      }),
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
            name="slug"
            render={({ field }) => (
              <Field>
                <FieldLabel>Slug</FieldLabel>
                <FieldControl>
                  <Input placeholder="acme" {...field} />
                </FieldControl>
                <FieldDescription>
                  Lowercase letters, numbers, and hyphens. Cannot be changed
                  later.
                </FieldDescription>
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
