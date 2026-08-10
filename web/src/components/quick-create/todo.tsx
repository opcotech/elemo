import { zodResolver } from "@hookform/resolvers/zod";
import { PlusIcon } from "lucide-react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { MoreProperties } from "@/components/quick-create/more-properties";
import type { QuickCreateKindProps } from "@/components/quick-create/types";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { DatePicker } from "@/components/ui/date-picker";
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
import { useAuth } from "@/hooks/use-auth";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { v1TodosGetOptions } from "@/lib/api/query-options";
import { v1TodosCreate } from "@/lib/api/sdk";
import type { Options, TodoCreate, V1TodosCreateData } from "@/lib/api/types";
import { zTodoCreate } from "@/lib/client/zod.gen";
import { createFormSchema, normalizeFormData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

const todoQuickCreateSchema = createFormSchema(
  zTodoCreate.omit({ owned_by: true })
);

type TodoQuickCreateValues = z.infer<typeof todoQuickCreateSchema>;

const defaultValues: TodoQuickCreateValues = {
  title: "",
  description: "",
  priority: "normal",
  due_date: null,
};

export function TodoQuickCreate({
  onCancel,
  onComplete,
}: QuickCreateKindProps) {
  const { user } = useAuth();
  const form = useForm<TodoQuickCreateValues>({
    resolver: zodResolver(todoQuickCreateSchema),
    defaultValues,
  });

  const mutation = useFormMutation<
    unknown,
    Options<V1TodosCreateData>,
    TodoQuickCreateValues
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
        todoQuickCreateSchema,
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
              name="description"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Description</FieldLabel>
                  <FieldControl>
                    <Textarea
                      placeholder="Add context (optional)"
                      rows={3}
                      {...field}
                      value={getDefaultValue(field.value)}
                    />
                  </FieldControl>
                  <FieldError />
                </Field>
              )}
            />

            <div className="flex gap-4">
              <ControlledField
                control={form.control}
                name="priority"
                render={({ field }) => (
                  <Field className="w-1/3">
                    <FieldLabel>Priority</FieldLabel>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      items={{
                        normal: "Normal",
                        important: "Important",
                        urgent: "Urgent",
                        critical: "Critical",
                      }}
                    >
                      <FieldControl>
                        <SelectTrigger className="w-full">
                          <SelectValue placeholder="Select a priority" />
                        </SelectTrigger>
                      </FieldControl>
                      <SelectContent>
                        <SelectItem value="normal">Normal</SelectItem>
                        <SelectItem value="important">Important</SelectItem>
                        <SelectItem value="urgent">Urgent</SelectItem>
                        <SelectItem value="critical">Critical</SelectItem>
                      </SelectContent>
                    </Select>
                    <FieldError />
                  </Field>
                )}
              />

              <ControlledField
                control={form.control}
                name="due_date"
                render={({ field }) => (
                  <Field className="w-2/3">
                    <FieldLabel>Due Date</FieldLabel>
                    <FieldControl>
                      <DatePicker
                        date={field.value ? new Date(field.value) : null}
                        onDateChange={(date) => {
                          field.onChange(date ? date.toISOString() : null);
                        }}
                        placeholder="Due date (optional)"
                        disabledDays={[
                          {
                            before: new Date(new Date().setHours(0, 0, 0, 0)),
                          },
                        ]}
                      />
                    </FieldControl>
                    <FieldError />
                  </Field>
                )}
              />
            </div>
          </MoreProperties>
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
