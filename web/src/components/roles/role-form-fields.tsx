import type { Control, FieldValues, Path } from "react-hook-form";
import type { z } from "zod";

import { ActionMultiSelect } from "./action-multi-select";

import {
  ControlledField,
  Field,
  FieldControl,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { NameDescriptionFields } from "@/components/ui/name-description-fields";
import { zRoleCreate } from "@/lib/api/schemas";
import { createFormSchema } from "@/lib/forms";
import { getDefaultValue } from "@/lib/utils";

export const roleFormSchema = createFormSchema(zRoleCreate);

export type RoleFormValues = z.infer<typeof roleFormSchema>;

export function RoleFormFields<TFieldValues extends FieldValues>({
  control,
  isPending = false,
  showKey = true,
  keyDisabled = false,
  showActions = true,
}: {
  control: Control<TFieldValues>;
  isPending?: boolean;
  showKey?: boolean;
  keyDisabled?: boolean;
  showActions?: boolean;
}) {
  return (
    <FieldGroup>
      <NameDescriptionFields
        control={control}
        isPending={isPending}
        namePlaceholder="Enter role name"
        descriptionPlaceholder="Enter role description"
      />
      {showKey ? (
        <ControlledField
          control={control}
          name={"key" as Path<TFieldValues>}
          render={({ field }) => (
            <Field>
              <FieldLabel>Key</FieldLabel>
              <FieldDescription>
                Stable identifier for this role bundle, for example org-member.
              </FieldDescription>
              <FieldControl>
                <Input
                  placeholder="org-member"
                  {...field}
                  value={getDefaultValue(field.value)}
                  disabled={isPending || keyDisabled}
                />
              </FieldControl>
              <FieldError />
            </Field>
          )}
        />
      ) : null}
      {showActions ? (
        <ActionMultiSelect control={control} isPending={isPending} />
      ) : null}
    </FieldGroup>
  );
}
