import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import type { TodoPriority } from "@/lib/api";
import { v1TodoUpdateMutation } from "@/lib/client/@tanstack/react-query.gen";
import { zTodoPatch } from "@/lib/client/zod.gen";
import {
  createFormSchema,
  getFieldValue,
  normalizePatchData,
} from "@/lib/forms";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

// Use form schema with empty string handling for optional fields
const todoEditFormSchema = createFormSchema(zTodoPatch);

type TodoEditFormValues = z.infer<typeof todoEditFormSchema>;

interface TodoItem {
  id: string;
  title: string;
  description: string;
  priority: TodoPriority;
  completed: boolean;
  due_date: string | null;
  created_at: string;
}

interface EditTodoFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  todo: TodoItem | null;
}

export function EditTodoForm({
  open,
  onOpenChange,
  onSuccess,
  todo,
}: EditTodoFormProps) {
  const form = useForm<TodoEditFormValues>({
    resolver: zodResolver(todoEditFormSchema),
    defaultValues: {
      title: "",
      description: "",
      priority: "normal",
      due_date: null,
    },
  });

  const mutation = useMutation(v1TodoUpdateMutation());

  // Update form values when todo changes
  useEffect(() => {
    if (todo && open) {
      form.reset({
        title: todo.title,
        description: getFieldValue(todo.description),
        priority: todo.priority,
        due_date: todo.due_date,
      });
    }
  }, [todo, open, form]);

  const onSubmit = (values: TodoEditFormValues) => {
    if (!todo) return;

    // Normalize patch data: converts empty strings to null for cleared optional fields
    const normalizedBody = normalizePatchData(todoEditFormSchema, values, {
      title: todo.title,
      description: todo.description,
      priority: todo.priority,
      due_date: todo.due_date,
    });

    mutation.mutate(
      {
        path: { id: todo.id },
        body: normalizedBody,
      },
      {
        onSuccess: () => {
          onOpenChange(false);
          onSuccess?.();
          form.reset();
          showSuccessToast(
            "Todo updated successfully",
            `Todo "${values.title || todo.title}" has been updated`
          );
        },
        onError: (error) => {
          showErrorToast("Failed to update todo", error.message);
        },
      }
    );
  };

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Edit Todo"
      onSubmit={form.handleSubmit(onSubmit)}
      isPending={mutation.isPending}
      error={mutation.error as Error | null}
      submitButtonText="Update todo"
      onReset={() => form.reset()}
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
              <Select onValueChange={field.onChange} defaultValue={field.value}>
                <FormControl>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select priority" />
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
    </DialogForm>
  );
}
