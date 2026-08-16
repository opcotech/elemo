import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";

import {
  TodoFormFields,
  todoCreateFormSchema,
  todoFormDefaultValues,
} from "./todo-form-fields";
import type { TodoCreateFormValues } from "./todo-form-fields";

import { Checkbox } from "@/components/ui/checkbox";
import { DialogForm } from "@/components/ui/dialog-form";
import { Label } from "@/components/ui/label";
import { useAuth } from "@/hooks/use-auth";
import { useFormMutation } from "@/hooks/use-form-mutation";
import { v1TodosCreate } from "@/lib/api/sdk";
import type { Options, TodoCreate, V1TodosCreateData } from "@/lib/api/types";
import { normalizeFormData } from "@/lib/forms";

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
    resetFormOnSuccess: false,
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
      if (!createMore) {
        onOpenChange(false);
      }
      onSuccess?.();
      form.reset(todoFormDefaultValues);
    },
  });

  const handleReset = () => {
    form.reset(todoFormDefaultValues);
    setCreateMore(false);
  };

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Add Todo"
      data-section="todo-add-form"
      onSubmit={mutation.handleSubmit}
      isPending={mutation.isPending}
      error={mutation.error || null}
      submitButtonText="Add todo"
      onReset={handleReset}
      className="sm:max-w-xl"
    >
      <TodoFormFields control={form.control} />
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
