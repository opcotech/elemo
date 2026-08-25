import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
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
import { ImmutableIdentifierField } from "@/components/ui/immutable-identifier-field";
import { Input } from "@/components/ui/input";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { accessibleNamespacesQueryKey } from "@/lib/api/accessible-namespaces";
import {
  v1OrganizationGetOptions,
  v1OrganizationsGetOptions,
} from "@/lib/api/query-options";
import { organizationRefPath } from "@/lib/api/refs";
import { zOrganizationCreate, zOrganizationPatch } from "@/lib/api/schemas";
import { v1OrganizationUpdate } from "@/lib/api/sdk";
import type {
  Options,
  Organization,
  V1OrganizationUpdateData,
} from "@/lib/api/types";
import { createFormSchema, normalizePatchData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

const organizationEditFormSchema = createFormSchema(
  zOrganizationPatch
    .omit({
      logo: true,
      status: true,
    })
    .extend({
      name: zOrganizationCreate.def.shape.name,
      email: zOrganizationCreate.def.shape.email,
    })
);

type OrganizationEditFormValues = z.infer<typeof organizationEditFormSchema>;

interface OrganizationEditFormProps {
  organization: Organization;
  organizationId: string;
}

export function OrganizationEditForm({
  organization,
  organizationId,
}: OrganizationEditFormProps) {
  const navigate = useNavigate();
  const organizationRoute = {
    to: "/settings/organizations/$organizationSlug",
    params: { organizationSlug: organization.slug },
  } as const;

  const form = useForm<OrganizationEditFormValues>({
    resolver: zodResolver(organizationEditFormSchema),
    defaultValues: {
      name: organization.name,
      email: organization.email,
      website: getDefaultValue(organization.website),
    },
  });

  useEffect(() => {
    if (!form.formState.isDirty) {
      form.reset({
        name: organization.name,
        email: organization.email,
        website: getDefaultValue(organization.website),
      });
    }
  }, [organization.name, organization.email, organization.website, form]);

  const mutation = useFormMutation<
    Organization,
    Options<V1OrganizationUpdateData>,
    OrganizationEditFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1OrganizationUpdate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Organization updated",
    errorMessagePrefix: "Failed to update organization",
    queryKeysToInvalidate: [
      v1OrganizationGetOptions({
        path: organizationRefPath(organizationId),
      }).queryKey,
      v1OrganizationGetOptions({
        path: organizationRefPath(organization.slug),
      }).queryKey,
      v1OrganizationsGetOptions().queryKey,
      accessibleNamespacesQueryKey,
    ],
    navigateOnSuccess: (navigateTo) => navigateTo(organizationRoute),
    transformValues: (values) => {
      const normalizedBody = normalizePatchData(
        organizationEditFormSchema,
        values,
        {
          name: organization.name,
          email: organization.email,
          website: organization.website,
        }
      );
      return {
        path: organizationRefPath(organizationId),
        body: normalizedBody,
      };
    },
  });

  return (
    <FormCard
      data-section="organization-edit-form"
      title="Edit Organization"
      description="Update the organization details below."
      onSubmit={mutation.handleSubmit}
      onCancel={() => navigate(organizationRoute)}
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Save Changes"
    >
      <FieldProvider {...form}>
        <FieldGroup>
          <ImmutableIdentifierField label="Slug" value={organization.slug} />

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
