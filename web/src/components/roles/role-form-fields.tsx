import { zodResolver } from "@hookform/resolvers/zod";
import type React from "react";
import { useForm } from "react-hook-form";
import type { UseFormReturn } from "react-hook-form";
import type { z } from "zod";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import { zRoleCreate } from "@/lib/client/zod.gen";
import { createFormSchema } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

const roleFormSchema = createFormSchema(zRoleCreate);

type RoleFormValues = z.infer<typeof roleFormSchema>;

interface RoleFormFieldsProps {
  isPending?: boolean;
  errorMessage?: string;
  onCancel?: () => void;
  submitButtonText?: string;
  onSubmit:
    | ((values: RoleFormValues) => void | Promise<void>)
    | ((e?: React.BaseSyntheticEvent<object, any, any>) => Promise<void>);
  defaultValues?: Partial<RoleFormValues>;
  form?: UseFormReturn<RoleFormValues>;
}

const defaultFormValues: RoleFormValues = {
  name: "",
  description: "",
};

export function RoleFormFields({
  isPending = false,
  errorMessage,
  onCancel,
  submitButtonText = "Create Role",
  onSubmit,
  defaultValues = defaultFormValues,
  form: providedForm,
}: RoleFormFieldsProps) {
  const internalForm = useForm<RoleFormValues>({
    resolver: zodResolver(roleFormSchema),
    defaultValues,
  });
  const form = providedForm || internalForm;

  return (
    <Form {...form}>
      <form
        onSubmit={
          providedForm
            ? (onSubmit as (
                e?: React.BaseSyntheticEvent<object, any, any>
              ) => Promise<void>)
            : form.handleSubmit(
                onSubmit as (values: RoleFormValues) => void | Promise<void>
              )
        }
        className="flex flex-col gap-y-6"
      >
        {errorMessage && (
          <Alert variant="destructive">
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{errorMessage}</AlertDescription>
          </Alert>
        )}

        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Name</FormLabel>
              <FormControl>
                <Input
                  placeholder="Enter role name"
                  {...field}
                  disabled={isPending}
                />
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
                  placeholder="Enter role description (optional)"
                  {...field}
                  value={getDefaultValue(field.value)}
                  rows={4}
                  disabled={isPending}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="flex justify-end gap-2">
          {onCancel && (
            <Button
              type="button"
              variant="outline"
              onClick={onCancel}
              disabled={isPending}
            >
              Cancel
            </Button>
          )}
          <Button type="submit" disabled={isPending}>
            {isPending ? (
              <>
                <Spinner size="xs" className="mr-0.5 text-white" />
                <span>Saving...</span>
              </>
            ) : (
              submitButtonText
            )}
          </Button>
        </div>
      </form>
    </Form>
  );
}

export { roleFormSchema, type RoleFormValues };
