import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "@tanstack/react-router";
import { useForm } from "react-hook-form";

import {
  normalizeProjectKey,
  projectCreateFormSchema,
} from "@/components/projects/project-form-schema";
import type { ProjectCreateFormValues } from "@/components/projects/project-form-schema";
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
import { useFormMutation } from "@/hooks/use-form-mutation";
import { accessibleNamespacesQueryKey } from "@/lib/api/accessible-namespaces";
import {
  v1NamespaceGetOptions,
  v1NamespacesProjectsGetOptions,
} from "@/lib/api/query-options";
import { v1NamespacesProjectsCreate } from "@/lib/api/sdk";
import type {
  Options,
  ProjectCreate,
  V1NamespacesProjectsCreateData,
} from "@/lib/api/types";
import { normalizeFormData } from "@/lib/forms";

const defaultValues: ProjectCreateFormValues = {
  key: "",
  name: "",
  description: "",
};

interface ProjectCreateFormProps {
  organizationId: string;
  namespaceId: string;
}

export function ProjectCreateForm({
  organizationId,
  namespaceId,
}: ProjectCreateFormProps) {
  const navigate = useNavigate();
  const namespaceRoute = {
    to: "/settings/organizations/$organizationId/namespaces/$namespaceId",
    params: { organizationId, namespaceId },
  } as const;

  const form = useForm<ProjectCreateFormValues>({
    resolver: zodResolver(projectCreateFormSchema),
    defaultValues,
  });

  const mutation = useFormMutation<
    { id: string },
    Options<V1NamespacesProjectsCreateData>,
    ProjectCreateFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1NamespacesProjectsCreate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Project created",
    errorMessagePrefix: "Failed to create project",
    queryKeysToInvalidate: [
      v1NamespaceGetOptions({
        path: { id: namespaceId },
      }).queryKey,
      v1NamespacesProjectsGetOptions({
        path: { id: namespaceId },
      }).queryKey,
      accessibleNamespacesQueryKey,
    ],
    transformValues: (values) => {
      const normalized = normalizeFormData(projectCreateFormSchema, {
        ...values,
        key: normalizeProjectKey(values.key),
      }) as ProjectCreate;
      return {
        path: {
          id: namespaceId,
        },
        body: normalized,
      };
    },
    navigateOnSuccess: (navigateTo, data) =>
      navigateTo({
        to: "/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId",
        params: {
          organizationId,
          namespaceId,
          projectId: data.id,
        },
      }),
  });

  return (
    <FormCard
      data-section="project-create-form"
      description="Enter the details below to create a new project."
      onSubmit={mutation.handleSubmit}
      onCancel={() => navigate(namespaceRoute)}
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Create Project"
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
        </FieldGroup>
      </FieldProvider>
    </FormCard>
  );
}
