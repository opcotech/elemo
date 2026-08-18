import type { Control, FieldPath, FieldValues } from "react-hook-form";

import { SearchableEntitySelect } from "@/components/ui/entity-select";
import type { EntitySelectOption } from "@/components/ui/entity-select";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import type { LibraryFolderPickerOption } from "@/lib/documents/library";

export function FolderPickerField<TFieldValues extends FieldValues>({
  control,
  name,
  options,
  disabled,
}: {
  control: Control<TFieldValues>;
  name: FieldPath<TFieldValues>;
  options: readonly LibraryFolderPickerOption[];
  disabled?: boolean;
}) {
  const selectOptions: EntitySelectOption[] = options.map((option) => ({
    value: option.value,
    title: option.title,
    searchText: option.searchText,
  }));

  return (
    <ControlledField
      control={control}
      name={name}
      render={({ field }) => (
        <Field>
          <FieldLabel>Folder</FieldLabel>
          <FieldControl>
            <SearchableEntitySelect
              options={selectOptions}
              value={field.value}
              onValueChange={field.onChange}
              placeholder="Select a folder"
              searchPlaceholder="Search folders…"
              emptyMessage="No folders found."
              aria-label="Folder"
              disabled={disabled}
            />
          </FieldControl>
          <FieldError />
        </Field>
      )}
    />
  );
}
