import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
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
import type {
  Options,
  Organization,
  V1OrganizationUpdateData,
} from "@/lib/api";
import { v1OrganizationUpdate } from "@/lib/client/sdk.gen";
import { zOrganizationCreate, zOrganizationPatch } from "@/lib/client/zod.gen";
import { createFormSchema, normalizePatchData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

// Create a schema without logo and status fields for the form
// TODO: Add logo field when implementing image upload
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
    navigateOnSuccess: {
      to: "/settings/organizations/$organizationId",
      params: { organizationId },
    },
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
        path: {
          id: organizationId,
        },
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
      onCancel={() =>
        navigate({
          to: "/settings/organizations/$organizationId",
          params: { organizationId },
        })
      }
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Save Changes"
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
