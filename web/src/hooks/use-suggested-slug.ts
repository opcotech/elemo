import { useEffect, useRef } from "react";
import type { FieldValues, Path, UseFormReturn } from "react-hook-form";

import { suggestSlug } from "@/lib/slug";

/** Keep `slug` in sync with `name` until the user edits the slug field. */
export function useSuggestedSlug<TFieldValues extends FieldValues>(
  form: UseFormReturn<TFieldValues>,
  nameField: Path<TFieldValues> = "name" as Path<TFieldValues>,
  slugField: Path<TFieldValues> = "slug" as Path<TFieldValues>
) {
  const name = form.watch(nameField);
  const slug = form.watch(slugField);
  const dirtyFields = form.formState.dirtyFields as unknown as Partial<
    Record<Path<TFieldValues>, boolean>
  >;
  const slugDirty = Boolean(dirtyFields[slugField]);
  const lastSuggestionRef = useRef("");

  useEffect(() => {
    if (typeof name !== "string") {
      return;
    }

    const suggested = suggestSlug(name);
    const current = typeof slug === "string" ? slug : "";

    if (slugDirty) {
      lastSuggestionRef.current = current;
      return;
    }

    // Keep a user-typed slug even when RHF has not marked the field dirty.
    // Empty values still receive a fresh suggestion from the name.
    if (current !== "" && current !== lastSuggestionRef.current) {
      return;
    }

    if (suggested === current) {
      lastSuggestionRef.current = suggested;
      return;
    }

    lastSuggestionRef.current = suggested;
    form.setValue(slugField, suggested as TFieldValues[typeof slugField], {
      shouldDirty: false,
      shouldValidate: false,
    });
  }, [form, name, slug, slugDirty, slugField]);
}
