import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { Checkbox } from "@/components/ui/checkbox";
import { DatePicker } from "@/components/ui/date-picker";
import { DialogForm } from "@/components/ui/dialog-form";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { useAuth } from "@/hooks/use-auth";
import { useFormMutation } from "@/hooks/use-form-mutation";
import type { Options, TodoCreate, V1TodosCreateData } from "@/lib/api";
import { v1TodosCreate } from "@/lib/client/sdk.gen";
import { zTodoCreate } from "@/lib/client/zod.gen";
import { createFormSchema, normalizeFormData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

const todoFormSchema = createFormSchema(zTodoCreate.omit({ owned_by: true }));

type TodoFormValues = z.infer<typeof todoFormSchema>;

const defaultValues: TodoFormValues = {
  title: "",
  description: "",
  priority: "normal",
  due_date: null,
};

interface AddTodoFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function AddTodoForm({
  open,
  onOpenChange,
  onSuccess,
}: AddTodoFormProps) {
  const { user } = useAuth();
  const [createMore, setCreateMore] = useState(false);

  const form = useForm<TodoFormValues>({
    resolver: zodResolver(todoFormSchema),
    defaultValues,
  });

  const mutation = useFormMutation<
    unknown,
    Options<V1TodosCreateData>,
    TodoFormValues
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
    resetFormOnSuccess: false, // We'll handle reset manually based on createMore
    transformValues: (values) => {
      const normalizedBody = normalizeFormData(
        todoFormSchema,
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
      if (!createMore) {
        onOpenChange(false);
      }
      onSuccess?.();
      form.reset(defaultValues);
      setCreateMore(false);
    },
  });

  const handleReset = () => {
    form.reset(defaultValues);
    setCreateMore(false);
  };

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Add Todo"
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Add todo"
      onReset={handleReset}
    >
      <FormField
        control={form.control}
        name="title"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Title</FormLabel>
            <FormControl>
              <Input placeholder="Enter todo title" {...field} />
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
                placeholder="Enter todo description (optional)"
                className="min-h-40 resize-y"
                rows={6}
                {...field}
                value={getDefaultValue(field.value)}
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />

      <div className="flex gap-4">
        <FormField
          control={form.control}
          name="priority"
          render={({ field }) => (
            <FormItem className="w-1/3">
              <FormLabel>Priority</FormLabel>
              <Select value={field.value} onValueChange={field.onChange}>
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select a priority" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value="normal">Normal</SelectItem>
                  <SelectItem value="important">Important</SelectItem>
                  <SelectItem value="urgent">Urgent</SelectItem>
                  <SelectItem value="critical">Critical</SelectItem>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="due_date"
          render={({ field }) => (
            <FormItem className="w-2/3">
              <FormLabel>Due Date</FormLabel>
              <FormControl>
                <DatePicker
                  date={field.value ? new Date(field.value) : null}
                  onDateChange={(date) => {
                    field.onChange(date ? date.toISOString() : null);
                  }}
                  placeholder="Due date (optional)"
                  disabledDays={[
                    { before: new Date(new Date().setHours(0, 0, 0, 0)) },
                  ]}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </div>

      <div className="flex items-center gap-2">
        <Checkbox
          id="createMore"
          checked={createMore}
          onCheckedChange={(checked) => setCreateMore(!!checked)}
        />
        <Label htmlFor="createMore" className="font-normal">
          Create more
        </Label>
      </div>
    </DialogForm>
  );
}
