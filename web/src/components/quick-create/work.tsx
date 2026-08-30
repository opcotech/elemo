import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { PlusIcon } from "lucide-react";
import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { CustomFieldEditor } from "@/components/custom-fields/field-editor";
import { QuickCreateContext } from "@/components/quick-create/context-panel";
import { MoreProperties } from "@/components/quick-create/more-properties";
import type { QuickCreateKindProps } from "@/components/quick-create/types";
import { MockDataAlert } from "@/components/shared/app-feedback";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { DialogFooter } from "@/components/ui/dialog";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldProvider,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { useNavigationContext } from "@/hooks/use-navigation-context";
import {
  v1CustomFieldsGetOptions,
  v1ProjectsIssuesGetOptions,
} from "@/lib/api/query-options";
import { projectIdPath } from "@/lib/api/refs";
import { zIssueCreate } from "@/lib/api/schemas";
import { v1ProjectsIssuesCreate } from "@/lib/api/sdk";
import type {
  CustomFieldValue,
  IssueCreate,
  Options,
  V1ProjectsIssuesCreateData,
} from "@/lib/api/types";
import {
  customFieldWritesFromValues,
  missingRequiredCustomFieldNames,
} from "@/lib/custom-fields/value";
import { createFormSchema, normalizeFormData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";
import { useOrganizationMembersForNamespace } from "@/lib/work/use-organization-members-for-namespace";

const workQuickCreateSchema = createFormSchema(
  zIssueCreate.pick({
    title: true,
    description: true,
    kind: true,
  })
);

type WorkQuickCreateValues = z.infer<typeof workQuickCreateSchema>;

const defaultValues: WorkQuickCreateValues = {
  title: "",
  description: "",
  kind: "task",
};

export function WorkQuickCreate({
  onCancel,
  onComplete,
}: QuickCreateKindProps) {
  const navigation = useNavigationContext();
  const canCreateInProject =
    navigation.type === "project" && Boolean(navigation.projectId);
  const [customValues, setCustomValues] = useState<
    Record<string, CustomFieldValue | undefined>
  >({});
  const [customFieldError, setCustomFieldError] = useState<string>();
  const { members } = useOrganizationMembersForNamespace(
    navigation.namespaceId,
    { enabled: canCreateInProject }
  );
  const {
    data: customFieldDefinitions = [],
    isPending: customFieldsPending,
    isError: customFieldsError,
  } = useQuery({
    ...v1CustomFieldsGetOptions({
      query: {
        scope_id: navigation.projectId ?? "",
        scope_type: "Project",
        target_type: "Issue",
        include_archived: false,
      },
    }),
    enabled: canCreateInProject,
    retry: false,
  });
  const activeDefinitions = useMemo(
    () => customFieldDefinitions.filter((definition) => !definition.archived),
    [customFieldDefinitions]
  );
  const requiredDefinitions = activeDefinitions.filter(
    (definition) => definition.required
  );
  const optionalDefinitions = activeDefinitions.filter(
    (definition) => !definition.required
  );
  const customFieldsBlocked = customFieldsPending || customFieldsError;

  const form = useForm<WorkQuickCreateValues>({
    resolver: zodResolver(workQuickCreateSchema),
    defaultValues,
  });

  const mutation = useFormMutation<
    unknown,
    Options<V1ProjectsIssuesCreateData>,
    WorkQuickCreateValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1ProjectsIssuesCreate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Issue created successfully",
    errorMessagePrefix: "Failed to create issue",
    resetFormOnSuccess: true,
    queryKeysToInvalidate: canCreateInProject
      ? [
          v1ProjectsIssuesGetOptions({
            path: projectIdPath(navigation.projectId!),
            query: { page_size: 100 },
          }).queryKey,
        ]
      : [],
    transformValues: (values) => {
      const normalizedBody = normalizeFormData(
        workQuickCreateSchema,
        values
      ) as Pick<IssueCreate, "title" | "description" | "kind">;
      const customFields = customFieldWritesFromValues(
        activeDefinitions,
        customValues
      );
      return {
        path: projectIdPath(navigation.projectId!),
        body: {
          kind: normalizedBody.kind,
          title: normalizedBody.title,
          description: normalizedBody.description || undefined,
          ...(customFields.length > 0 ? { custom_fields: customFields } : {}),
        },
      };
    },
    onSuccess: () => {
      onComplete();
    },
  });

  const errorMessage = mutation.error?.message;

  if (!canCreateInProject) {
    return (
      <FieldProvider {...form}>
        <form
          onSubmit={(event) => {
            event.preventDefault();
          }}
        >
          <FieldGroup className="my-5">
            <ControlledField
              control={form.control}
              name="title"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Title</FieldLabel>
                  <FieldControl>
                    <Input
                      autoFocus
                      placeholder="What needs to be done?"
                      {...field}
                    />
                  </FieldControl>
                  <FieldError />
                </Field>
              )}
            />

            <QuickCreateContext />

            <MockDataAlert title="Project context required">
              Issues can be created from a project Work surface. Open a project
              first, then use quick create.
            </MockDataAlert>
          </FieldGroup>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button type="submit" disabled>
              <PlusIcon />
              Create unavailable
            </Button>
          </DialogFooter>
        </form>
      </FieldProvider>
    );
  }

  return (
    <FieldProvider {...form}>
      <form
        onSubmit={(event) => {
          const missing = missingRequiredCustomFieldNames(
            requiredDefinitions,
            customValues
          );
          if (missing.length > 0) {
            event.preventDefault();
            setCustomFieldError(
              `Fill required custom fields: ${missing.join(", ")}`
            );
            return;
          }
          setCustomFieldError(undefined);
          mutation.handleSubmit(event);
        }}
      >
        <FieldGroup className="my-5">
          {errorMessage && (
            <Alert variant="destructive">
              <AlertTitle>Failed to save</AlertTitle>
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
          )}
          {customFieldsError ? (
            <Alert variant="destructive">
              <AlertTitle>Custom fields unavailable</AlertTitle>
              <AlertDescription>
                Custom field definitions could not be loaded. Try again before
                creating the issue.
              </AlertDescription>
            </Alert>
          ) : null}
          {customFieldError ? (
            <Alert variant="destructive">
              <AlertTitle>Custom fields required</AlertTitle>
              <AlertDescription>{customFieldError}</AlertDescription>
            </Alert>
          ) : null}

          <ControlledField
            control={form.control}
            name="title"
            render={({ field }) => (
              <Field>
                <FieldLabel>Title</FieldLabel>
                <FieldControl>
                  <Input
                    autoFocus
                    placeholder="What needs to be done?"
                    {...field}
                  />
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />

          <MoreProperties>
            <ControlledField
              control={form.control}
              name="kind"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Kind</FieldLabel>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                    items={{
                      task: "Task",
                      story: "Story",
                      bug: "Bug",
                      epic: "Epic",
                    }}
                  >
                    <FieldControl>
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Select a kind" />
                      </SelectTrigger>
                    </FieldControl>
                    <SelectContent>
                      <SelectItem value="task">Task</SelectItem>
                      <SelectItem value="story">Story</SelectItem>
                      <SelectItem value="bug">Bug</SelectItem>
                      <SelectItem value="epic">Epic</SelectItem>
                    </SelectContent>
                  </Select>
                  <FieldError />
                </Field>
              )}
            />

            <ControlledField
              control={form.control}
              name="description"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Description</FieldLabel>
                  <FieldControl>
                    <Textarea
                      placeholder="Add context"
                      rows={3}
                      {...field}
                      value={getDefaultValue(field.value)}
                    />
                  </FieldControl>
                  <FieldError />
                </Field>
              )}
            />
          </MoreProperties>

          {requiredDefinitions.map((definition) => (
            <div key={definition.id} className="space-y-2">
              <Label>
                {definition.name}
                {definition.required ? " *" : ""}
              </Label>
              <CustomFieldEditor
                key={definition.id}
                definition={definition}
                value={customValues[definition.id]}
                members={members}
                onCommit={(next) =>
                  setCustomValues((current) => ({
                    ...current,
                    [definition.id]: next,
                  }))
                }
              />
            </div>
          ))}

          {optionalDefinitions.length > 0 ? (
            <MoreProperties>
              {optionalDefinitions.map((definition) => (
                <div key={definition.id} className="space-y-2">
                  <Label>{definition.name}</Label>
                  <CustomFieldEditor
                    key={definition.id}
                    definition={definition}
                    value={customValues[definition.id]}
                    members={members}
                    onCommit={(next) =>
                      setCustomValues((current) => ({
                        ...current,
                        [definition.id]: next,
                      }))
                    }
                  />
                </div>
              ))}
            </MoreProperties>
          ) : null}

          <QuickCreateContext />
        </FieldGroup>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={onCancel}
            disabled={mutation.isPending}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            disabled={mutation.isPending || customFieldsBlocked}
          >
            {mutation.isPending ? (
              <>
                <Spinner size="xs" className="mr-0.5 text-white" />
                <span>Saving...</span>
              </>
            ) : (
              <>
                <PlusIcon />
                Create issue
              </>
            )}
          </Button>
        </DialogFooter>
      </form>
    </FieldProvider>
  );
}
