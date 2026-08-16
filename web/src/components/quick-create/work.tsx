import { zodResolver } from "@hookform/resolvers/zod";
import { PlusIcon } from "lucide-react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

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
import { v1ProjectsIssuesGetOptions } from "@/lib/api/query-options";
import { v1ProjectsIssuesCreate } from "@/lib/api/sdk";
import type {
  IssueCreate,
  Options,
  V1ProjectsIssuesCreateData,
} from "@/lib/api/types";
import { zIssueCreate } from "@/lib/client/zod.gen";
import { createFormSchema, normalizeFormData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

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
            path: { id: navigation.projectId! },
            query: { page_size: 100 },
          }).queryKey,
        ]
      : [],
    transformValues: (values) => {
      const normalizedBody = normalizeFormData(
        workQuickCreateSchema,
        values
      ) as Pick<IssueCreate, "title" | "description" | "kind">;
      return {
        path: { id: navigation.projectId! },
        body: {
          kind: normalizedBody.kind,
          title: normalizedBody.title,
          description: normalizedBody.description || undefined,
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
      <form onSubmit={mutation.handleSubmit}>
        <FieldGroup className="my-5">
          {errorMessage && (
            <Alert variant="destructive">
              <AlertTitle>Failed to save</AlertTitle>
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
          )}

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
          <Button type="submit" disabled={mutation.isPending}>
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
