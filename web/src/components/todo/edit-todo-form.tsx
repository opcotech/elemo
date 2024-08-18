import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { Button } from "@/components/ui/button";
import { DatePicker } from "@/components/ui/date-picker";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Form,
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
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import type { TodoPriority } from "@/lib/api";
import { v1TodoUpdateMutation } from "@/lib/client/@tanstack/react-query.gen";
import { zTodoPatch } from "@/lib/client/zod.gen";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

// Use the existing generated schema for todo updates
const todoEditFormSchema = zTodoPatch;

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
      description: undefined,
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
        description: todo.description || undefined,
        priority: todo.priority,
        due_date: todo.due_date,
      });
    }
  }, [todo, open, form]);

  const onSubmit = (values: TodoEditFormValues) => {
    if (!todo) return;

    // Only include fields that have values
    const updateData: TodoEditFormValues = {};
    if (values.title !== undefined) updateData.title = values.title;
    if (values.description !== undefined)
      updateData.description = values.description;
    if (values.priority !== undefined) updateData.priority = values.priority;
    if (values.due_date !== undefined) updateData.due_date = values.due_date;

    mutation.mutate(
      {
        path: { id: todo.id },
        body: updateData,
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

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      // Reset form when closing
      form.reset();
    }
    onOpenChange(newOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="flex flex-col gap-y-6"
          >
            <DialogHeader>
              <DialogTitle>Edit Todo</DialogTitle>
            </DialogHeader>

            {mutation.isError && (
              <div className="text-destructive text-sm">
                <p>{mutation.error.message}</p>
              </div>
            )}
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
                      value={field.value || undefined}
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
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
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

            <DialogFooter>
              <Button type="submit" disabled={mutation.isPending}>
                {mutation.isPending ? (
                  <>
                    <Spinner size="xs" className="mr-0.5 text-white" />
                    <span>Updating todo...</span>
                  </>
                ) : (
                  "Update todo"
                )}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
