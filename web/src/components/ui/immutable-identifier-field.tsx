import { useId } from "react";

import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export function ImmutableIdentifierField({
  label,
  value,
  description = "Set on create and cannot be changed.",
}: {
  label: string;
  value: string;
  description?: string;
}) {
  const id = useId();

  return (
    <Field>
      <Label htmlFor={id}>{label}</Label>
      <Input id={id} aria-label={label} value={value} readOnly disabled />
      <p className="text-muted-foreground text-sm leading-normal font-normal">
        {description}
      </p>
    </Field>
  );
}
