import type { Control, FieldValues, Path } from "react-hook-form";

import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { getDefaultValue } from "@/lib/utils";

export function NameDescriptionFields<TFieldValues extends FieldValues>({
  control,
  isPending = false,
  namePlaceholder = "Enter name",
  descriptionPlaceholder = "Enter description",
  descriptionRows = 4,
}: {
  control: Control<TFieldValues>;
  isPending?: boolean;
  namePlaceholder?: string;
  descriptionPlaceholder?: string;
  descriptionRows?: number;
}) {
  return (
    <>
      <ControlledField
        control={control}
        name={"name" as Path<TFieldValues>}
        render={({ field }) => (
          <Field>
            <FieldLabel>Name</FieldLabel>
            <FieldControl>
              <Input
                placeholder={namePlaceholder}
                {...field}
                disabled={isPending}
              />
            </FieldControl>
            <FieldError />
          </Field>
        )}
      />

      <ControlledField
        control={control}
        name={"description" as Path<TFieldValues>}
        render={({ field }) => (
          <Field>
            <FieldLabel>Description</FieldLabel>
            <FieldControl>
              <Textarea
                placeholder={descriptionPlaceholder}
                {...field}
                value={getDefaultValue(field.value)}
                rows={descriptionRows}
                disabled={isPending}
              />
            </FieldControl>
            <FieldError />
          </Field>
        )}
      />
    </>
  );
}
