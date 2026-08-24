import type { Control, FieldValues } from "react-hook-form";
import type { z } from "zod";

import { FieldGroup } from "@/components/ui/field";
import { NameDescriptionFields } from "@/components/ui/name-description-fields";
import { zTeamCreate } from "@/lib/api/schemas";
import { createFormSchema } from "@/lib/forms";

export const teamFormSchema = createFormSchema(zTeamCreate);

export type TeamFormValues = z.infer<typeof teamFormSchema>;

export function TeamFormFields<TFieldValues extends FieldValues>({
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
        namePlaceholder="Enter team name"
        descriptionPlaceholder="Enter team description"
      />
    </FieldGroup>
  );
}
