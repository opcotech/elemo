import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { useState } from "react";
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
import { useAuth } from "@/hooks/use-auth";
import { v1TodosCreateMutation } from "@/lib/client/@tanstack/react-query.gen";
import { zTodoCreate } from "@/lib/client/zod.gen";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

// Use the existing generated schema for todo creation
// We need to modify it slightly for the form since we don't want to require owned_by in the form
const todoFormSchema = zTodoCreate.omit({ owned_by: true });

type TodoFormValues = z.infer<typeof todoFormSchema>;

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
    defaultValues: {
      priority: "normal",
      due_date: null,
    },
  });

  const mutation = useMutation(v1TodosCreateMutation());

  const onSubmit = (values: TodoFormValues) => {
    mutation.mutate(
      {
        body: {
          title: values.title,
          description: values.description,
          priority: values.priority,
          due_date: values.due_date,
          owned_by: user!.id,
        },
      },
      {
        onSuccess: () => {
          if (!createMore) onOpenChange(false);
          onSuccess?.();
          form.reset();
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
              <DialogTitle>Add Todo</DialogTitle>
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

            <DialogFooter className="flex flex-col items-start gap-2 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="createMore"
                  name="createMore"
                  className="h-4 w-4"
                  checked={createMore}
                  onChange={(e) => setCreateMore(e.target.checked)}
                />
                <label htmlFor="createMore" className="text-sm">
                  Create more
                </label>
              </div>
            
              <Button type="submit" disabled={mutation.isPending}>
                {mutation.isPending ? (
                  <>
                    <Spinner size="xs" className="mr-0.5 text-white" />
                    <span>Adding todo...</span>
                  </>
                ) : (
                  "Add todo"
                )}
              </Button>
            </DialogFooter>

          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
