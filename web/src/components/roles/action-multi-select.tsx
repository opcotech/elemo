import type { Control, FieldValues, Path } from "react-hook-form";

import { Checkbox } from "@/components/ui/checkbox";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { inspectableActions } from "@/lib/auth/permissions";

export function ActionMultiSelect<TFieldValues extends FieldValues>({
  control,
  name = "actions" as Path<TFieldValues>,
  isPending = false,
  description = "Select the exact actions this bundle grants. There is no wildcard.",
}: {
  control: Control<TFieldValues>;
  name?: Path<TFieldValues>;
  isPending?: boolean;
  description?: string;
}) {
  return (
    <ControlledField
      control={control}
      name={name}
      render={({ field }) => {
        const selected = new Set<string>(
          Array.isArray(field.value) ? field.value : []
        );

        return (
          <Field>
            <FieldLabel>Actions</FieldLabel>
            <FieldDescription>{description}</FieldDescription>
            <FieldControl>
              <div
                data-slot="checkbox-group"
                className="grid max-h-64 gap-2 overflow-y-auto rounded-lg border p-3 sm:grid-cols-2"
              >
                {inspectableActions.map((action) => {
                  const checked = selected.has(action);
                  return (
                    <label
                      key={action}
                      className="flex cursor-pointer items-center gap-2 text-sm"
                    >
                      <Checkbox
                        checked={checked}
                        disabled={isPending}
                        onCheckedChange={(value) => {
                          const next = new Set(selected);
                          if (value) {
                            next.add(action);
                          } else {
                            next.delete(action);
                          }
                          field.onChange([...next]);
                        }}
                      />
                      <span className="font-mono text-xs">{action}</span>
                    </label>
                  );
                })}
              </div>
            </FieldControl>
            <FieldError />
          </Field>
        );
      }}
    />
  );
}
