import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
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
import { v1OrganizationsCreateMutation } from "@/lib/client/@tanstack/react-query.gen";
import { zOrganizationCreate } from "@/lib/client/zod.gen";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

// Create a schema without logo field for the form
// TODO: Add logo field back in when implementing image upload
const organizationFormSchema = zOrganizationCreate.omit({ logo: true });

type OrganizationFormValues = z.infer<typeof organizationFormSchema>;

const defaultValues: OrganizationFormValues = {
  name: "",
  email: "",
  website: undefined,
};

export function OrganizationCreateForm() {
  const navigate = useNavigate();

  const form = useForm<OrganizationFormValues>({
    resolver: zodResolver(organizationFormSchema),
    defaultValues,
  });

  const mutation = useMutation(v1OrganizationsCreateMutation());

  const onSubmit = (values: OrganizationFormValues) => {
    const body: {
      name: string;
      email: string;
      website?: string;
    } = {
      name: values.name,
      email: values.email,
    };

    if (values.website) {
      body.website = values.website;
    }

    mutation.mutate(
      {
        body,
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
    <Card>
      <CardHeader>
        <CardTitle>Create Organization</CardTitle>
        <CardDescription>
          Enter the details below to create a new organization.
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
                onClick={() => navigate({ to: "/settings/organizations" })}
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
                  "Create Organization"
                )}
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
