import type { z } from "zod";

import {
  zProjectCreate,
  zProjectPatch,
  zProjectStatus,
} from "@/lib/api/schemas";
import { createFormSchema } from "@/lib/forms";

/** Normalize a project key to uppercase for display and API submission. */
export function normalizeProjectKey(key: string): string {
  return key.toUpperCase();
}

export const projectKeySchema = zProjectCreate.shape.key.regex(
  /^[A-Za-z]+$/,
  "Project key must contain only ASCII letters (A-Z or a-z)"
);

export const projectCreateFormSchema = createFormSchema(
  zProjectCreate.omit({ logo: true, status: true })
).extend({
  key: projectKeySchema,
});

export const projectEditFormSchema = createFormSchema(
  zProjectPatch.omit({ logo: true })
).extend({
  key: projectKeySchema,
  name: zProjectCreate.shape.name,
  status: zProjectStatus,
});

export type ProjectCreateFormValues = z.infer<typeof projectCreateFormSchema>;
export type ProjectEditFormValues = z.infer<typeof projectEditFormSchema>;
