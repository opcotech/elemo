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
  FieldDescription,
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
import { isConflict, throwIfApiFailed } from "@/lib/api/errors";
import {
  v1NamespaceGetOptions,
  v1NamespacesProjectsGetOptions,
} from "@/lib/api/query-options";
import { namespaceRefPath } from "@/lib/api/refs";
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
  organizationSlug: string;
  namespaceId: string;
  namespaceSlug: string;
}

export function ProjectCreateForm({
  organizationId,
  organizationSlug,
  namespaceId,
  namespaceSlug,
}: ProjectCreateFormProps) {
  const navigate = useNavigate();
  const namespaceRoute = {
    to: "/settings/organizations/$organizationSlug/namespaces/$namespaceSlug",
    params: { organizationSlug, namespaceSlug },
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
      return throwIfApiFailed(
        await v1NamespacesProjectsCreate({
          ...variables,
          throwOnError: false,
        })
      );
    },
    form,
    successMessage: "Project created",
    errorMessagePrefix: "Failed to create project",
    queryKeysToInvalidate: [
      v1NamespaceGetOptions({
        path: namespaceRefPath(organizationId, namespaceId),
      }).queryKey,
      v1NamespacesProjectsGetOptions({
        path: namespaceRefPath(organizationId, namespaceId),
      }).queryKey,
      accessibleNamespacesQueryKey,
    ],
    onError: (error) => {
      if (isConflict(error)) {
        form.setError("key", {
          type: "conflict",
          message: "This project key is already in use in this namespace",
        });
      }
    },
    transformValues: (values) => {
      const normalized = normalizeFormData(projectCreateFormSchema, {
        ...values,
        key: normalizeProjectKey(values.key),
      }) as ProjectCreate;
      return {
        path: namespaceRefPath(organizationId, namespaceId),
        body: normalized,
      };
    },
    navigateOnSuccess: (navigateTo) =>
      navigateTo({
        to: "/settings/organizations/$organizationSlug/namespaces/$namespaceSlug/projects/$projectKey",
        params: {
          organizationSlug,
          namespaceSlug,
          projectKey: normalizeProjectKey(form.getValues("key")),
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
                    placeholder="PLAT"
                    {...field}
                    onChange={(event) =>
                      field.onChange(normalizeProjectKey(event.target.value))
                    }
                    disabled={mutation.isPending}
                  />
                </FieldControl>
                <FieldDescription>
                  2–6 letters, unique in this namespace. Cannot be changed
                  later.
                </FieldDescription>
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
