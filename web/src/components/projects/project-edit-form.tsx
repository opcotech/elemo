import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useEffect } from "react";
import { useForm } from "react-hook-form";

import { projectEditFormSchema } from "@/components/projects/project-form-schema";
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
import { ImmutableIdentifierField } from "@/components/ui/immutable-identifier-field";
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
  v1NamespacesProjectsKeyGetOptions,
  v1ProjectGetOptions,
} from "@/lib/api/query-options";
import { namespaceRefPath, projectIdPath } from "@/lib/api/refs";
import { v1ProjectUpdate } from "@/lib/api/sdk";
import type { Options, Project, V1ProjectUpdateData } from "@/lib/api/types";
import { normalizePatchData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

interface ProjectEditFormProps {
  project: Project;
  organizationId: string;
  organizationSlug: string;
  namespaceId: string;
  namespaceSlug: string;
}

export function ProjectEditForm({
  project,
  organizationId,
  organizationSlug,
  namespaceId,
  namespaceSlug,
}: ProjectEditFormProps) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const projectDetailRoute = {
    to: "/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey",
    params: {
      organizationSlug,
      namespaceSlug,
      projectKey: project.key,
    },
  } as const;

  const form = useForm<ProjectEditFormValues>({
    resolver: zodResolver(projectEditFormSchema),
    defaultValues: {
      name: project.name,
      description: getDefaultValue(project.description),
      status: project.status,
    },
  });

  useEffect(() => {
    if (!form.formState.isDirty) {
      form.reset({
        name: project.name,
        description: getDefaultValue(project.description),
        status: project.status,
      });
    }
  }, [project.name, project.description, project.status, form]);

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
        path: projectIdPath(project.id),
      }).queryKey,
      v1NamespacesProjectsKeyGetOptions({
        path: {
          ...namespaceRefPath(organizationSlug, namespaceSlug),
          projectKey: project.key,
        },
      }).queryKey,
      v1NamespacesProjectsKeyGetOptions({
        path: {
          ...namespaceRefPath(organizationId, namespaceId),
          projectKey: project.key,
        },
      }).queryKey,
      v1NamespaceGetOptions({
        path: namespaceRefPath(organizationId, namespaceId),
      }).queryKey,
      v1NamespacesProjectsGetOptions({
        path: namespaceRefPath(organizationId, namespaceId),
      }).queryKey,
      accessibleNamespacesQueryKey,
    ],
    navigateOnSuccess: (navigateTo) => navigateTo(projectDetailRoute),
    onSuccess: (data) => {
      const projectCaches = [
        v1ProjectGetOptions({
          path: projectIdPath(data.id),
        }).queryKey,
        v1NamespacesProjectsKeyGetOptions({
          path: {
            ...namespaceRefPath(organizationSlug, namespaceSlug),
            projectKey: data.key,
          },
        }).queryKey,
        v1NamespacesProjectsKeyGetOptions({
          path: {
            ...namespaceRefPath(organizationId, namespaceId),
            projectKey: data.key,
          },
        }).queryKey,
      ];
      for (const queryKey of projectCaches) {
        queryClient.setQueryData(queryKey, data);
      }
    },
    transformValues: (values) => {
      const normalizedBody = normalizePatchData(projectEditFormSchema, values, {
        name: project.name,
        description: project.description,
        status: project.status,
      });
      return {
        path: projectIdPath(project.id),
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
          <ImmutableIdentifierField
            label="Key"
            value={project.key}
            description="Project key is set on create and cannot be changed."
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
