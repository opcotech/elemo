import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import type { z } from "zod";

import { TodoFormFields } from "./todo-form-fields";

import { DialogForm } from "@/components/ui/dialog-form";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { zTodoCreate, zTodoPatch } from "@/lib/api/schemas";
import { v1TodoUpdate } from "@/lib/api/sdk";
import type { Options, TodoPriority, V1TodoUpdateData } from "@/lib/api/types";
import { createFormSchema, normalizePatchData } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

const todoEditFormSchema = createFormSchema(
  zTodoPatch.extend({
    title: zTodoCreate.def.shape.title,
  })
);

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

  const mutation = useFormMutation<
    unknown,
    Options<V1TodoUpdateData>,
    TodoEditFormValues
  >({
    mutationFn: async (variables) => {
      const { data } = await v1TodoUpdate({
        ...variables,
        throwOnError: true,
      });
      return data;
    },
    form,
    successMessage: "Todo updated successfully",
    errorMessagePrefix: "Failed to update todo",
    resetFormOnSuccess: true,
    transformValues: (values) => {
      if (!todo) {
        throw new Error("Todo is required");
      }
      const normalizedBody = normalizePatchData(todoEditFormSchema, values, {
        title: todo.title,
        description: todo.description,
        priority: todo.priority,
        due_date: todo.due_date,
      });
      return {
        path: { id: todo.id },
        body: normalizedBody,
      };
    },
    onSuccess: () => {
      onOpenChange(false);
      onSuccess?.();
    },
  });

  useEffect(() => {
    if (todo && open) {
      form.reset({
        title: todo.title,
        description: getDefaultValue(todo.description),
        priority: todo.priority,
        due_date: todo.due_date,
      });
    }
  }, [todo, open, form]);

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Edit Todo"
      data-section="todo-edit-form"
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Update todo"
      onReset={() => form.reset()}
      className="sm:max-w-xl"
    >
      <TodoFormFields control={form.control} />
    </DialogForm>
  );
}
