import { describe, expect, it } from "vitest";
import { z } from "zod";

import {
  createFormSchema,
  normalizeFormData,
  normalizePatchData,
} from "@/lib/forms";

const resourceSchema = z.object({
  name: z.string().min(1),
  website: z.string().url().optional(),
  note: z.union([z.string().min(1), z.null()]).optional(),
});

describe("form helpers", () => {
  it("accepts empty optional controls without weakening required fields", () => {
    const formSchema = createFormSchema(resourceSchema);

    expect(
      formSchema.parse({ name: "Elemo", website: "", note: null })
    ).toEqual({
      name: "Elemo",
      website: undefined,
      note: undefined,
    });
    expect(formSchema.safeParse({ name: "", website: "" }).success).toBe(false);
  });

  it("omits empty optional values from create payloads", () => {
    expect(
      normalizeFormData(resourceSchema, {
        name: "Elemo",
        website: "",
        note: undefined,
      })
    ).toEqual({ name: "Elemo" });
  });

  it("turns cleared optional values into null for patch payloads", () => {
    expect(
      normalizePatchData(
        resourceSchema,
        { name: "Elemo", website: "" },
        { name: "Elemo", website: "https://elemo.app" }
      )
    ).toEqual({ name: "Elemo", website: null });
  });
});
