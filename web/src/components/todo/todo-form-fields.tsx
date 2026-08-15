import type { ReactNode } from "react";
import type { Control, FieldValues, Path } from "react-hook-form";
import type { z } from "zod";

import { todoPriorities, todoPriorityLabels } from "./priority";
import { TodoPriorityRibbon } from "./todo-priority-ribbon";

import { DatePicker } from "@/components/ui/date-picker";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import type { TodoPriority } from "@/lib/api/types";
import { zTodoCreate } from "@/lib/client/zod.gen";
import { createFormSchema } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

export const todoCreateFormSchema = createFormSchema(
  zTodoCreate.omit({ owned_by: true })
);

export type TodoCreateFormValues = z.infer<typeof todoCreateFormSchema>;

export const todoFormDefaultValues: TodoCreateFormValues = {
  title: "",
  description: "",
  priority: "normal",
  due_date: null,
};

interface TodoFormFieldsProps<TFieldValues extends FieldValues> {
  control: Control<TFieldValues>;
  titlePlaceholder?: string;
  descriptionPlaceholder?: string;
  descriptionRows?: number;
  autoFocusTitle?: boolean;
  wrapExtras?: (fields: ReactNode) => ReactNode;
}

function pastDisabledDays() {
  return [{ before: new Date(new Date().setHours(0, 0, 0, 0)) }];
}

export function TodoFormFields<TFieldValues extends FieldValues>({
  control,
  titlePlaceholder = "Enter todo title",
  descriptionPlaceholder = "Enter todo description",
  descriptionRows = 6,
  autoFocusTitle = false,
  wrapExtras = (fields) => fields,
}: TodoFormFieldsProps<TFieldValues>) {
  const extras = (
    <>
      <ControlledField
        control={control}
        name={"description" as Path<TFieldValues>}
        render={({ field }) => (
          <Field>
            <FieldLabel>Description</FieldLabel>
            <FieldControl>
              <Textarea
                placeholder={descriptionPlaceholder}
                className={
                  descriptionRows >= 6 ? "min-h-40 resize-y" : undefined
                }
                rows={descriptionRows}
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
          control={control}
          name={"priority" as Path<TFieldValues>}
          render={({ field }) => (
            <Field className="w-1/3">
              <FieldLabel>Priority</FieldLabel>
              <Select
                value={field.value}
                onValueChange={field.onChange}
                items={todoPriorityLabels}
              >
                <FieldControl>
                  <SelectTrigger className="w-full" aria-label="Priority">
                    <SelectValue placeholder="Select a priority">
                      {field.value ? (
                        <TodoPriorityRibbon
                          priority={field.value as TodoPriority}
                        />
                      ) : null}
                    </SelectValue>
                  </SelectTrigger>
                </FieldControl>
                <SelectContent align="start" alignItemWithTrigger={false}>
                  {todoPriorities.map((priority) => (
                    <SelectItem key={priority} value={priority}>
                      <TodoPriorityRibbon priority={priority} />
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FieldError />
            </Field>
          )}
        />

        <ControlledField
          control={control}
          name={"due_date" as Path<TFieldValues>}
          render={({ field }) => (
            <Field className="w-2/3">
              <FieldLabel>Due Date</FieldLabel>
              <FieldControl>
                <DatePicker
                  date={field.value ? new Date(field.value) : null}
                  onDateChange={(date) => {
                    field.onChange(date ? date.toISOString() : null);
                  }}
                  placeholder="Due date"
                  disabledDays={pastDisabledDays()}
                />
              </FieldControl>
              <FieldError />
            </Field>
          )}
        />
      </div>
    </>
  );

  return (
    <>
      <ControlledField
        control={control}
        name={"title" as Path<TFieldValues>}
        render={({ field }) => (
          <Field>
            <FieldLabel>Title</FieldLabel>
            <FieldControl>
              <Input
                autoFocus={autoFocusTitle}
                placeholder={titlePlaceholder}
                {...field}
              />
            </FieldControl>
            <FieldError />
          </Field>
        )}
      />
      {wrapExtras(extras)}
    </>
  );
}
