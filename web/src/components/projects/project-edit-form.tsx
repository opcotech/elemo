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
import { NameDescriptionFields } from "@/components/ui/name-description-fields";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { accessibleNamespacesQueryKey } from "@/lib/api/accessible-namespaces";
import {
  v1NamespaceGetOptions,
  v1NamespacesProjectsGetOptions,
  v1ProjectGetOptions,
} from "@/lib/api/query-options";
import { v1ProjectUpdate } from "@/lib/api/sdk";
import type { Options, Project, V1ProjectUpdateData } from "@/lib/api/types";
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
      accessibleNamespacesQueryKey,
    ],
    navigateOnSuccess: (navigateTo) => navigateTo(projectDetailRoute),
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
      <FieldProvider {...form}>
        <FieldGroup>
          <ControlledField
            control={form.control}
            name="key"
            render={({ field }) => (
              <Field>
                <FieldLabel>Key</FieldLabel>
                <FieldControl>
                  <Input
                    placeholder="Enter project key"
                    {...field}
                    onChange={(event) =>
                      field.onChange(normalizeProjectKey(event.target.value))
                    }
                    disabled={mutation.isPending}
                  />
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />

          <NameDescriptionFields
            control={form.control}
            isPending={mutation.isPending}
            namePlaceholder="Enter project name"
            descriptionPlaceholder="Enter project description"
          />

          <ControlledField
            control={form.control}
            name="status"
            render={({ field }) => (
              <Field>
                <FieldLabel>Status</FieldLabel>
                <Select
                  onValueChange={field.onChange}
                  value={field.value}
                  disabled={mutation.isPending}
                  items={{
                    active: "Active",
                    pending: "Pending",
                  }}
                >
                  <FieldControl>
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select status" />
                    </SelectTrigger>
                  </FieldControl>
                  <SelectContent>
                    <SelectItem value="active">Active</SelectItem>
                    <SelectItem value="pending">Pending</SelectItem>
                  </SelectContent>
                </Select>
                <FieldError />
              </Field>
            )}
          />
        </FieldGroup>
      </FieldProvider>
    </FormCard>
  );
}
