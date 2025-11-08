import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
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
import type { Organization } from "@/lib/api";
import { v1OrganizationUpdateMutation } from "@/lib/client/@tanstack/react-query.gen";
import { zOrganizationPatch } from "@/lib/client/zod.gen";
import {
  createFormSchema,
  getFieldValue,
  normalizePatchData,
} from "@/lib/forms";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

// Create a schema without logo and status fields for the form
// TODO: Add logo field when implementing image upload
const organizationEditFormSchema = createFormSchema(
  zOrganizationPatch.omit({
    logo: true,
    status: true,
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
      website: getFieldValue(organization.website),
    },
  });

  // Update form when organization data changes (but not while user is editing)
  useEffect(() => {
    // Only reset if the form is not being actively edited (isDirty check)
    if (!form.formState.isDirty) {
      form.reset({
        name: organization.name,
        email: organization.email,
        website: getFieldValue(organization.website),
      });
    }
  }, [organization.name, organization.email, organization.website]);

  const mutation = useMutation(v1OrganizationUpdateMutation());

  const onSubmit = (values: OrganizationEditFormValues) => {
    // Normalize patch data: converts empty strings to null for cleared optional fields
    const normalizedBody = normalizePatchData(
      organizationEditFormSchema,
      values,
      {
        name: organization.name,
        email: organization.email,
        website: organization.website,
      }
    );

    mutation.mutate(
      {
        path: {
          id: organizationId,
        },
        body: normalizedBody,
      },
      {
        onSuccess: () => {
          showSuccessToast(
            "Organization updated",
            "Organization updated successfully"
          );
          navigate({
            to: "/settings/organizations/$organizationId",
            params: { organizationId },
          });
        },
        onError: (error) => {
          showErrorToast("Failed to update organization", error.message);
        },
      }
    );
  };

  return (
    <FormCard
      onSubmit={form.handleSubmit(onSubmit)}
      onCancel={() =>
        navigate({
          to: "/settings/organizations/$organizationId",
          params: { organizationId },
        })
      }
      isPending={mutation.isPending}
      error={mutation.error as Error}
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
