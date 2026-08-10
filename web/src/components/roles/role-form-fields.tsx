import type React from "react";
import type { UseFormReturn } from "react-hook-form";
import type { z } from "zod";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
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
  onSubmit: React.FormEventHandler<HTMLFormElement>;
  form: UseFormReturn<RoleFormValues>;
}

export function RoleFormFields({
  isPending = false,
  errorMessage,
  onCancel,
  submitButtonText = "Create Role",
  onSubmit,
  form,
}: RoleFormFieldsProps) {
  return (
    <FieldProvider {...form}>
      <form onSubmit={onSubmit} className="flex flex-col gap-y-6">
        {errorMessage && (
          <Alert variant="destructive">
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{errorMessage}</AlertDescription>
          </Alert>
        )}

        <FieldGroup>
          <ControlledField
            control={form.control}
            name="name"
            render={({ field }) => (
              <Field>
                <FieldLabel>Name</FieldLabel>
                <FieldControl>
                  <Input
                    placeholder="Enter role name"
                    {...field}
                    disabled={isPending}
                  />
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />

          <ControlledField
            control={form.control}
            name="description"
            render={({ field }) => (
              <Field>
                <FieldLabel>Description</FieldLabel>
                <FieldControl>
                  <Textarea
                    placeholder="Enter role description (optional)"
                    {...field}
                    value={getDefaultValue(field.value)}
                    rows={4}
                    disabled={isPending}
                  />
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />
        </FieldGroup>

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
    </FieldProvider>
  );
}

export { roleFormSchema, type RoleFormValues };
