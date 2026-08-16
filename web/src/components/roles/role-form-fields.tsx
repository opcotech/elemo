import type { Control, FieldValues } from "react-hook-form";
import type { z } from "zod";

import { FieldGroup } from "@/components/ui/field";
import { NameDescriptionFields } from "@/components/ui/name-description-fields";
import { zRoleCreate } from "@/lib/client/zod.gen";
import { createFormSchema } from "@/lib/forms";

export const roleFormSchema = createFormSchema(zRoleCreate);

export type RoleFormValues = z.infer<typeof roleFormSchema>;

export function RoleFormFields<TFieldValues extends FieldValues>({
  control,
  isPending = false,
}: {
  control: Control<TFieldValues>;
  isPending?: boolean;
}) {
  return (
    <FieldGroup>
      <NameDescriptionFields
        control={control}
        isPending={isPending}
        namePlaceholder="Enter role name"
        descriptionPlaceholder="Enter role description"
      />
    </FieldGroup>
  );
}
