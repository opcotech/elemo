import type { Control } from "react-hook-form";

import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import type { DocumentCreateFormValues } from "@/lib/documents/create";

export function DocumentCreateFields({
  control,
  autoFocus = true,
}: {
  control: Control<DocumentCreateFormValues>;
  autoFocus?: boolean;
}) {
  return (
    <ControlledField
      control={control}
      name="title"
      render={({ field }) => (
        <Field>
          <FieldLabel>Title</FieldLabel>
          <FieldControl>
            <Input
              autoFocus={autoFocus}
              placeholder="Untitled document"
              {...field}
            />
          </FieldControl>
          <FieldError />
        </Field>
      )}
    />
  );
}
