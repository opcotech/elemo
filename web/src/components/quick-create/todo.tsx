import { zodResolver } from "@hookform/resolvers/zod";
import { PlusIcon } from "lucide-react";
import { useForm } from "react-hook-form";

import { MoreProperties } from "@/components/quick-create/more-properties";
import type { QuickCreateKindProps } from "@/components/quick-create/types";
import {
  TodoFormFields,
  todoCreateFormSchema,
  todoFormDefaultValues,
} from "@/components/todo/todo-form-fields";
import type { TodoCreateFormValues } from "@/components/todo/todo-form-fields";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { DialogFooter } from "@/components/ui/dialog";
import { FieldGroup, FieldProvider } from "@/components/ui/field";
import { Spinner } from "@/components/ui/spinner";
import { useAuth } from "@/hooks/use-auth";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { v1TodosGetOptions } from "@/lib/api/query-options";
import { v1TodosCreate } from "@/lib/api/sdk";
import type { Options, TodoCreate, V1TodosCreateData } from "@/lib/api/types";
import { normalizeFormData } from "@/lib/forms";

export function TodoQuickCreate({
  onCancel,
  onComplete,
}: QuickCreateKindProps) {
  const { user } = useAuth();
  const form = useForm<TodoCreateFormValues>({
    resolver: zodResolver(todoCreateFormSchema),
    defaultValues: todoFormDefaultValues,
  });

  const mutation = useFormMutation<
    unknown,
    Options<V1TodosCreateData>,
    TodoCreateFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1TodosCreate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Todo added successfully",
    errorMessagePrefix: "Failed to add todo",
    resetFormOnSuccess: true,
    queryKeysToInvalidate: [v1TodosGetOptions().queryKey],
    transformValues: (values) => {
      const normalizedBody = normalizeFormData(
        todoCreateFormSchema,
        values
      ) as TodoCreate;
      return {
        body: {
          ...normalizedBody,
          owned_by: user!.id,
        },
      };
    },
    onSuccess: () => {
      onComplete();
    },
  });

  const errorMessage = mutation.error?.message;

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

          <TodoFormFields
            control={form.control}
            titlePlaceholder="What needs to be done?"
            descriptionPlaceholder="Add context"
            descriptionRows={3}
            autoFocusTitle
            wrapExtras={(fields) => <MoreProperties>{fields}</MoreProperties>}
          />
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
                Create todo
              </>
            )}
          </Button>
        </DialogFooter>
      </form>
    </FieldProvider>
  );
}
