import { describe, expect, it } from "vitest";

import {
  todoCreateFormSchema,
  todoFormDefaultValues,
} from "./todo-form-fields";

describe("todoCreateFormSchema", () => {
  it("accepts a complete todo payload", () => {
    expect(
      todoCreateFormSchema.parse({
        title: "Write tests",
        description: "Cover the shared fields",
        priority: "normal",
        due_date: null,
      })
    ).toEqual({
      title: "Write tests",
      description: "Cover the shared fields",
      priority: "normal",
      due_date: null,
    });
  });

  it("requires a title", () => {
    expect(() =>
      todoCreateFormSchema.parse({
        ...todoFormDefaultValues,
        title: undefined,
      })
    ).toThrow();
  });
});
