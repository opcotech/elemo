import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
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
import type { TodoCreate } from "@/lib/api";
import { v1TodosCreateMutation } from "@/lib/client/@tanstack/react-query.gen";
import { zTodoCreate } from "@/lib/client/zod.gen";
import {
  createFormSchema,
  getFieldValue,
  normalizeFormData,
} from "@/lib/forms";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

// Use form schema with empty string handling for optional fields
// We need to modify it slightly for the form since we don't want to require owned_by in the form
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

  const mutation = useMutation(v1TodosCreateMutation());

  const onSubmit = (values: TodoFormValues) => {
    const normalizedBody = normalizeFormData(
      todoFormSchema,
      values
    ) as TodoCreate;

    mutation.mutate(
      {
        body: {
          ...normalizedBody,
          owned_by: user!.id,
        },
      },
      {
        onSuccess: () => {
          if (!createMore) onOpenChange(false);
          onSuccess?.();
          form.reset(defaultValues);
          showSuccessToast(
            "Todo added successfully",
            `Todo "${values.title}" with priority "${values.priority}" added successfully`
          );
        },
        onError: (error) => {
          showErrorToast("Failed to add todo", error.message);
        },
      }
    );
  };

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
      onSubmit={form.handleSubmit(onSubmit)}
      isPending={mutation.isPending}
      error={mutation.error as Error | null}
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
                value={getFieldValue(field.value)}
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
