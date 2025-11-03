import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

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
import type { Organization } from "@/lib/api";
import { v1OrganizationUpdateMutation } from "@/lib/client/@tanstack/react-query.gen";
import { zOrganizationPatch } from "@/lib/client/zod.gen";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

// Create a schema without logo and status fields for the form
// TODO: Add logo field when implementing image upload
const organizationEditFormSchema = zOrganizationPatch.omit({
  logo: true,
  status: true,
});

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
      website: organization.website || undefined,
    },
  });

  // Update form when organization data changes
  useEffect(() => {
    form.reset({
      name: organization.name,
      email: organization.email,
      website: organization.website || undefined,
    });
  }, [organization, form]);

  const mutation = useMutation(v1OrganizationUpdateMutation());

  const onSubmit = (values: OrganizationEditFormValues) => {
    const body: {
      name?: string;
      email?: string;
      website?: string;
    } = {};

    // Only include fields that have changed or are explicitly provided
    if (values.name !== organization.name) {
      body.name = values.name;
    }
    if (values.email !== organization.email) {
      body.email = values.email;
    }
    if (values.website !== (organization.website || undefined)) {
      body.website = values.website;
    }

    // If no changes, don't submit
    if (Object.keys(body).length === 0) {
      showErrorToast(
        "No changes",
        "Please make changes to the organization before saving."
      );
      return;
    }

    mutation.mutate(
      {
        path: {
          id: organizationId,
        },
        body,
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
    <Card>
      <CardHeader>
        <CardTitle>Edit Organization</CardTitle>
        <CardDescription>
          Update the organization details below.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col gap-y-6"
          >
            {mutation.isError && (
              <div className="text-destructive text-sm">
                <p>{mutation.error.message}</p>
              </div>
            )}

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
                      value={field.value || ""}
                      onChange={(e) => {
                        const value = e.target.value;
                        field.onChange(value === "" ? undefined : value);
                      }}
                      onBlur={field.onBlur}
                      name={field.name}
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
                    <span>Saving...</span>
                  </>
                ) : (
                  "Save Changes"
                )}
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
