import { describe, expect, it } from "vitest";

import type { CustomFieldDefinition, CustomFieldSchema } from "@/lib/api/types";
import {
  customFieldKindAllowsFullText,
  customFieldKindAllowsRange,
  customFieldOptionLabel,
  customFieldValueFromDraft,
  customFieldValuesEqual,
  customFieldWritesFromValues,
  datetimeLocalToUtcIso,
  decimalInputStep,
  defaultCustomFieldSchema,
  formatCustomFieldValue,
  isSafeHttpUrl,
  missingRequiredCustomFieldNames,
  nodeResourceTypes,
  parseIntegerDraft,
  resourceIdsFromValue,
  textDraftFromValue,
  userIdsFromValue,
  utcIsoToDatetimeLocal,
} from "@/lib/custom-fields/value";

function definition(
  overrides: Partial<CustomFieldDefinition> & {
    kind: CustomFieldDefinition["kind"];
    schema: CustomFieldSchema;
  }
): CustomFieldDefinition {
  return {
    id: "def-1",
    key: "story_points",
    name: "Story points",
    scope_id: "project-1",
    scope_type: "Project",
    target_type: "Issue",
    required: false,
    archived: false,
    index_exact: false,
    index_range: false,
    index_fulltext: false,
    order: 0,
    owner_user_id: "user-1",
    ...overrides,
  };
}

describe("custom field value conversion", () => {
  it("builds a typed schema for each kind", () => {
    expect(defaultCustomFieldSchema("text")).toEqual({ kind: "text" });
    expect(defaultCustomFieldSchema("url")).toEqual({
      kind: "url",
      allowed_schemes: ["https"],
    });
    expect(defaultCustomFieldSchema("single_select").kind).toBe(
      "single_select"
    );
    expect(defaultCustomFieldSchema("resource_reference")).toEqual({
      kind: "resource_reference",
      allowed_types: ["Issue"],
      multiple: false,
    });
  });

  it("lists searchable resource types and excludes internal graph types", () => {
    const types = nodeResourceTypes();
    expect(types).toContain("Issue");
    expect(types).toContain("Document");
    expect(types).toContain("Project");
    expect(types).not.toContain("Team");
    expect(types).not.toContain("Label");
    expect(types).not.toContain("Permission");
    expect(types).not.toContain("UserToken");
    expect(types).not.toContain("Notification");
    expect(types).not.toContain("CustomFieldDefinition");
  });

  it("converts drafts into closed typed values", () => {
    expect(customFieldValueFromDraft("text", { text: "hello" })).toEqual({
      kind: "text",
      text: "hello",
    });
    expect(customFieldValueFromDraft("integer", { integer: "8" })).toEqual({
      kind: "integer",
      integer: 8,
    });
    expect(
      customFieldValueFromDraft("integer", { integer: "8.5" })
    ).toBeUndefined();
    expect(
      customFieldValueFromDraft("integer", { integer: "1e2" })
    ).toBeUndefined();
    expect(
      customFieldValueFromDraft("integer", {
        integer: String(Number.MAX_SAFE_INTEGER + 1),
      })
    ).toBeUndefined();
    expect(customFieldValueFromDraft("boolean", { boolean: false })).toEqual({
      kind: "boolean",
      boolean: false,
    });
    expect(
      customFieldValueFromDraft("multi_select", { optionKeys: ["a", "b"] })
    ).toEqual({ kind: "multi_select", option_keys: ["a", "b"] });
    expect(customFieldValueFromDraft("text", { text: "" })).toBeUndefined();
  });

  it("formats stored values for display", () => {
    expect(formatCustomFieldValue({ kind: "boolean", boolean: true })).toBe(
      "Yes"
    );
    expect(
      formatCustomFieldValue({ kind: "multi_select", option_keys: ["a", "b"] })
    ).toBe("a, b");
    expect(formatCustomFieldValue(undefined)).toBe("");
    expect(
      formatCustomFieldValue({
        kind: "user_reference",
        user_id: "user-1",
      })
    ).toBe("user-1");
  });

  it("resolves select option labels from the definition schema", () => {
    const schema = {
      kind: "single_select" as const,
      options: [{ key: "alpha", label: "Alpha", disabled: false }],
    };
    expect(customFieldOptionLabel(schema, "alpha")).toBe("Alpha");
    expect(customFieldOptionLabel(schema, "missing")).toBe("missing");
    expect(customFieldOptionLabel({ kind: "text" }, "alpha")).toBe("alpha");
  });

  it("reads user and resource ids from single or multi payloads", () => {
    expect(
      userIdsFromValue({ kind: "user_reference", user_id: "user-1" })
    ).toEqual(["user-1"]);
    expect(
      userIdsFromValue({
        kind: "user_reference",
        user_ids: ["user-1", "user-2"],
      })
    ).toEqual(["user-1", "user-2"]);
    expect(userIdsFromValue({ kind: "text", text: "nope" })).toEqual([]);
    expect(
      resourceIdsFromValue({
        kind: "resource_reference",
        resource_id: "Issue:abc",
      })
    ).toEqual(["Issue:abc"]);
  });

  it("round-trips text drafts from stored values", () => {
    expect(
      textDraftFromValue({ kind: "decimal", decimal: "12.50" }, "decimal")
    ).toBe("12.50");
    expect(textDraftFromValue({ kind: "text", text: "hello" }, "integer")).toBe(
      ""
    );
    expect(
      textDraftFromValue(
        { kind: "user_reference", user_ids: ["user-9"] },
        "user_reference"
      )
    ).toBe("user-9");
  });

  it("treats unchanged values as equal, including user id shapes", () => {
    expect(customFieldValuesEqual(undefined, undefined)).toBe(true);
    expect(
      customFieldValuesEqual(
        { kind: "text", text: "hello" },
        { kind: "text", text: "hello" }
      )
    ).toBe(true);
    expect(
      customFieldValuesEqual(
        { kind: "text", text: "hello" },
        { kind: "text", text: "hello " }
      )
    ).toBe(false);
    expect(
      customFieldValuesEqual({ kind: "text", text: "hello" }, undefined)
    ).toBe(false);
    expect(
      customFieldValuesEqual(
        { kind: "user_reference", user_id: "user-1" },
        { kind: "user_reference", user_ids: ["user-1"] }
      )
    ).toBe(true);
    expect(
      customFieldValuesEqual(
        { kind: "multi_select", option_keys: ["a", "b"] },
        { kind: "multi_select", option_keys: ["b", "a"] }
      )
    ).toBe(false);
  });

  it("builds create writes and reports missing required fields", () => {
    const required = definition({
      id: "req",
      kind: "text",
      schema: { kind: "text" },
      required: true,
      name: "Release note",
    });
    const optional = definition({
      id: "opt",
      kind: "integer",
      schema: { kind: "integer" },
      key: "points",
      name: "Points",
    });

    expect(missingRequiredCustomFieldNames([required, optional], {})).toEqual([
      "Release note",
    ]);
    expect(
      customFieldWritesFromValues([required, optional], {
        req: { kind: "text", text: "shipped" },
      })
    ).toEqual([
      {
        definition_id: "req",
        value: { kind: "text", text: "shipped" },
      },
    ]);
  });

  it("gates index capabilities by kind", () => {
    expect(customFieldKindAllowsRange("integer")).toBe(true);
    expect(customFieldKindAllowsRange("text")).toBe(false);
    expect(customFieldKindAllowsFullText("text")).toBe(true);
    expect(customFieldKindAllowsFullText("integer")).toBe(false);
  });

  it("maps decimal scale to an HTML number input step", () => {
    expect(decimalInputStep(undefined)).toBe("any");
    expect(decimalInputStep(0)).toBe("1");
    expect(decimalInputStep(2)).toBe("0.01");
    expect(decimalInputStep(1)).toBe("0.1");
  });

  it("parses integer drafts without scientific notation or unsafe magnitudes", () => {
    expect(parseIntegerDraft("42")).toBe(42);
    expect(parseIntegerDraft("-3")).toBe(-3);
    expect(parseIntegerDraft("1e2")).toBeUndefined();
    expect(parseIntegerDraft("8.5")).toBeUndefined();
    expect(parseIntegerDraft("")).toBeUndefined();
  });

  it("round-trips datetime-local wall time through UTC ISO", () => {
    const local = "2026-08-20T14:30";
    const iso = datetimeLocalToUtcIso(local);
    expect(iso).toMatch(/Z$/);
    expect(utcIsoToDatetimeLocal(iso)).toBe(local);
    expect(utcIsoToDatetimeLocal("not-a-date")).toBe("");
    expect(datetimeLocalToUtcIso("")).toBe("");
  });

  it("accepts only http(s) URLs for readonly links", () => {
    expect(isSafeHttpUrl("https://example.com/path")).toBe(true);
    expect(isSafeHttpUrl("http://example.com")).toBe(true);
    expect(isSafeHttpUrl("javascript:alert(1)")).toBe(false);
    expect(isSafeHttpUrl("not a url")).toBe(false);
  });
});
