import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { useForm } from "react-hook-form";

import {
  normalizeProjectKey,
  projectEditFormSchema,
} from "@/components/projects/project-form-schema";
import type { ProjectEditFormValues } from "@/components/projects/project-form-schema";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { useFormMutation } from "@/hooks/use-form-mutation";
import type { Options, Project, V1ProjectUpdateData } from "@/lib/api";
import {
  v1NamespaceGetOptions,
  v1NamespacesProjectsGetOptions,
  v1ProjectGetOptions,
} from "@/lib/client/@tanstack/react-query.gen";
import { v1ProjectUpdate } from "@/lib/client/sdk.gen";
import { normalizePatchData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

interface ProjectEditFormProps {
  project: Project;
  organizationId: string;
  namespaceId: string;
}

export function ProjectEditForm({
  project,
  organizationId,
  namespaceId,
}: ProjectEditFormProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const projectDetailRoute = {
    to: "/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId",
    params: { organizationId, namespaceId, projectId: project.id },
  } as const;

  const form = useForm<ProjectEditFormValues>({
    resolver: zodResolver(projectEditFormSchema),
    defaultValues: {
      key: project.key,
      name: project.name,
      description: getDefaultValue(project.description),
      status: project.status,
    },
  });

  useEffect(() => {
    if (!form.formState.isDirty) {
      form.reset({
        key: project.key,
        name: project.name,
        description: getDefaultValue(project.description),
        status: project.status,
      });
    }
  }, [project.key, project.name, project.description, project.status, form]);

  const mutation = useFormMutation<
    Project,
    Options<V1ProjectUpdateData>,
    ProjectEditFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1ProjectUpdate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Project updated",
    errorMessagePrefix: "Failed to update project",
    queryKeysToInvalidate: [
      v1ProjectGetOptions({
        path: { id: project.id },
      }).queryKey,
      v1NamespaceGetOptions({
        path: { id: namespaceId },
      }).queryKey,
      v1NamespacesProjectsGetOptions({
        path: { id: namespaceId },
      }).queryKey,
    ],
    navigateOnSuccess: projectDetailRoute,
    onSuccess: (data) => {
      // Keep detail page from briefly rendering stale cached project data.
      queryClient.setQueryData(
        v1ProjectGetOptions({
          path: { id: data.id },
        }).queryKey,
        data
      );
    },
    transformValues: (values) => {
      const normalizedBody = normalizePatchData(
        projectEditFormSchema,
        {
          ...values,
          key: normalizeProjectKey(values.key),
        },
        {
          key: project.key,
          name: project.name,
          description: project.description,
          status: project.status,
        }
      );
      return {
        path: {
          id: project.id,
        },
        body: normalizedBody,
      };
    },
  });

  return (
    <FormCard
      data-section="project-edit-form"
      description="Update the project details below."
      onSubmit={mutation.handleSubmit}
      onCancel={() => navigate(projectDetailRoute)}
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Save Changes"
    >
      <Form {...form}>
        <FormField
          control={form.control}
          name="key"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Key</FormLabel>
              <FormControl>
                <Input
                  placeholder="Enter project key"
                  {...field}
                  onChange={(event) =>
                    field.onChange(normalizeProjectKey(event.target.value))
                  }
                  disabled={mutation.isPending}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input
                  placeholder="Enter project name"
                  {...field}
                  disabled={mutation.isPending}
                />
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
                  placeholder="Enter project description (optional)"
                  {...field}
                  value={getDefaultValue(field.value)}
                  rows={4}
                  disabled={mutation.isPending}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="status"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Status</FormLabel>
              <Select
                onValueChange={field.onChange}
                value={field.value}
                disabled={mutation.isPending}
              >
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select status" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="active">Active</SelectItem>
                  <SelectItem value="pending">Pending</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />
      </Form>
    </FormCard>
  );
}
