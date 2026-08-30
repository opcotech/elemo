import type {
  CustomFieldDefinition,
  CustomFieldKind,
  CustomFieldSchema,
  CustomFieldValue,
  CustomFieldWrite,
  ResourceType,
} from "@/lib/api/types";
import { SEARCH_RESOURCE_TYPES } from "@/lib/search/result";

export const customFieldKinds: readonly CustomFieldKind[] = [
  "text",
  "integer",
  "decimal",
  "boolean",
  "date",
  "datetime",
  "url",
  "single_select",
  "multi_select",
  "user_reference",
  "resource_reference",
];

export const customFieldKindMeta: Record<
  CustomFieldKind,
  {
    label: string;
    allowsRange: boolean;
    allowsFullText: boolean;
  }
> = {
  text: { label: "Text", allowsRange: false, allowsFullText: true },
  integer: { label: "Integer", allowsRange: true, allowsFullText: false },
  decimal: { label: "Decimal", allowsRange: true, allowsFullText: false },
  boolean: { label: "Boolean", allowsRange: false, allowsFullText: false },
  date: { label: "Date", allowsRange: true, allowsFullText: false },
  datetime: {
    label: "Date and time",
    allowsRange: true,
    allowsFullText: false,
  },
  url: { label: "URL", allowsRange: false, allowsFullText: false },
  single_select: {
    label: "Single select",
    allowsRange: false,
    allowsFullText: false,
  },
  multi_select: {
    label: "Multi select",
    allowsRange: false,
    allowsFullText: false,
  },
  user_reference: {
    label: "User",
    allowsRange: false,
    allowsFullText: false,
  },
  resource_reference: {
    label: "Resource",
    allowsRange: false,
    allowsFullText: false,
  },
};

export const customFieldKindLabels: Record<CustomFieldKind, string> =
  Object.fromEntries(
    customFieldKinds.map((kind) => [kind, customFieldKindMeta[kind].label])
  ) as Record<CustomFieldKind, string>;

export function customFieldKindAllowsRange(kind: CustomFieldKind): boolean {
  return customFieldKindMeta[kind].allowsRange;
}

export function customFieldKindAllowsFullText(kind: CustomFieldKind): boolean {
  return customFieldKindMeta[kind].allowsFullText;
}

/** HTML `step` for a decimal input. Scale 2 → `"0.01"`; unset scale → `"any"`. */
export function decimalInputStep(scale: number | undefined): string {
  if (scale == null || scale < 0) {
    return "any";
  }
  if (scale === 0) {
    return "1";
  }
  return (10 ** -scale).toFixed(scale);
}

// Resource-reference targets must be searchable so the picker can resolve them.
export function nodeResourceTypes(): readonly ResourceType[] {
  return SEARCH_RESOURCE_TYPES;
}

function padDatePart(value: number): string {
  return String(value).padStart(2, "0");
}

/** Convert a UTC ISO timestamp to a `datetime-local` wall-clock string. */
export function utcIsoToDatetimeLocal(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return `${date.getFullYear()}-${padDatePart(date.getMonth() + 1)}-${padDatePart(date.getDate())}T${padDatePart(date.getHours())}:${padDatePart(date.getMinutes())}`;
}

/** Convert a `datetime-local` wall-clock string to a UTC ISO timestamp. */
export function datetimeLocalToUtcIso(local: string): string {
  if (!local) {
    return "";
  }
  const date = new Date(local);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return date.toISOString();
}

export function isSafeHttpUrl(url: string): boolean {
  try {
    const parsed = new URL(url);
    return parsed.protocol === "http:" || parsed.protocol === "https:";
  } catch {
    return false;
  }
}

/** Parse a decimal-digit integer draft; rejects scientific notation and unsafe magnitudes. */
export function parseIntegerDraft(raw: string): number | undefined {
  const trimmed = raw.trim();
  if (!/^-?\d+$/.test(trimmed)) {
    return undefined;
  }
  const integer = Number(trimmed);
  if (!Number.isSafeInteger(integer)) {
    return undefined;
  }
  return integer;
}

export function defaultCustomFieldSchema(
  kind: CustomFieldKind
): CustomFieldSchema {
  switch (kind) {
    case "text":
      return { kind: "text" };
    case "integer":
      return { kind: "integer" };
    case "decimal":
      return { kind: "decimal" };
    case "boolean":
      return { kind: "boolean" };
    case "date":
      return { kind: "date" };
    case "datetime":
      return { kind: "datetime" };
    case "url":
      return { kind: "url", allowed_schemes: ["https"] };
    case "single_select":
      return {
        kind: "single_select",
        options: [{ key: "option_1", label: "Option 1", disabled: false }],
      };
    case "multi_select":
      return {
        kind: "multi_select",
        options: [{ key: "option_1", label: "Option 1", disabled: false }],
      };
    case "user_reference":
      return { kind: "user_reference", multiple: false };
    case "resource_reference":
      return {
        kind: "resource_reference",
        allowed_types: ["Issue"],
        multiple: false,
      };
  }
}

export function isMultiValued(
  kind: CustomFieldKind,
  schema: CustomFieldSchema
): boolean {
  if (kind === "multi_select") {
    return true;
  }
  if (kind === "user_reference" && schema.kind === "user_reference") {
    return Boolean(schema.multiple);
  }
  if (kind === "resource_reference" && schema.kind === "resource_reference") {
    return Boolean(schema.multiple);
  }
  return false;
}

export function userIdsFromValue(
  value: CustomFieldValue | undefined
): string[] {
  if (value?.kind !== "user_reference") {
    return [];
  }
  if (value.user_ids && value.user_ids.length > 0) {
    return value.user_ids;
  }
  return value.user_id ? [value.user_id] : [];
}

export function resourceIdsFromValue(
  value: CustomFieldValue | undefined
): string[] {
  if (value?.kind !== "resource_reference") {
    return [];
  }
  if (value.resource_ids && value.resource_ids.length > 0) {
    return value.resource_ids;
  }
  return value.resource_id ? [value.resource_id] : [];
}

function customFieldValueKey(value: CustomFieldValue): string {
  switch (value.kind) {
    case "text":
      return `text:${value.text}`;
    case "integer":
      return `integer:${value.integer}`;
    case "decimal":
      return `decimal:${value.decimal}`;
    case "boolean":
      return `boolean:${value.boolean}`;
    case "date":
      return `date:${value.date}`;
    case "datetime":
      return `datetime:${value.datetime}`;
    case "url":
      return `url:${value.url}`;
    case "single_select":
      return `single_select:${value.option_key}`;
    case "multi_select":
      return `multi_select:${value.option_keys.join("\0")}`;
    case "user_reference":
      return `user_reference:${userIdsFromValue(value).join("\0")}`;
    case "resource_reference":
      return `resource_reference:${resourceIdsFromValue(value).join("\0")}`;
  }
}

export function customFieldValuesEqual(
  left: CustomFieldValue | undefined,
  right: CustomFieldValue | undefined
): boolean {
  if (left === right) {
    return true;
  }
  if (!left || !right) {
    return false;
  }
  return customFieldValueKey(left) === customFieldValueKey(right);
}

export function customFieldOptionLabel(
  schema: CustomFieldSchema,
  key: string
): string {
  if (schema.kind !== "single_select" && schema.kind !== "multi_select") {
    return key;
  }
  return schema.options.find((option) => option.key === key)?.label ?? key;
}

export function formatCustomFieldValue(
  value: CustomFieldValue | undefined
): string {
  if (!value) {
    return "";
  }
  switch (value.kind) {
    case "text":
      return value.text;
    case "integer":
      return String(value.integer);
    case "decimal":
      return value.decimal;
    case "boolean":
      return value.boolean ? "Yes" : "No";
    case "date":
      return value.date;
    case "datetime":
      return value.datetime;
    case "url":
      return value.url;
    case "single_select":
      return value.option_key;
    case "multi_select":
      return value.option_keys.join(", ");
    case "user_reference":
      return value.user_ids?.join(", ") || value.user_id || "";
    case "resource_reference":
      return value.resource_ids?.join(", ") || value.resource_id || "";
  }
}

export function textDraftFromValue(
  value: CustomFieldValue | undefined,
  kind: CustomFieldKind
): string {
  if (!value || value.kind !== kind) {
    return "";
  }
  switch (value.kind) {
    case "text":
      return value.text;
    case "integer":
      return String(value.integer);
    case "decimal":
      return value.decimal;
    case "date":
      return value.date;
    case "datetime":
      return value.datetime;
    case "url":
      return value.url;
    case "single_select":
      return value.option_key;
    case "user_reference":
      return userIdsFromValue(value)[0] ?? "";
    case "resource_reference":
      return resourceIdsFromValue(value)[0] ?? "";
    default:
      return "";
  }
}

export function customFieldValueFromDraft(
  kind: CustomFieldKind,
  draft: {
    text?: string;
    integer?: string;
    decimal?: string;
    boolean?: boolean;
    date?: string;
    datetime?: string;
    url?: string;
    optionKey?: string;
    optionKeys?: string[];
    userId?: string;
    userIds?: string[];
    resourceId?: string;
    resourceIds?: string[];
  }
): CustomFieldValue | undefined {
  switch (kind) {
    case "text":
      return draft.text ? { kind: "text", text: draft.text } : undefined;
    case "integer": {
      const integer = parseIntegerDraft(draft.integer ?? "");
      return integer === undefined ? undefined : { kind: "integer", integer };
    }
    case "decimal":
      return draft.decimal
        ? { kind: "decimal", decimal: draft.decimal }
        : undefined;
    case "boolean":
      return { kind: "boolean", boolean: Boolean(draft.boolean) };
    case "date":
      return draft.date ? { kind: "date", date: draft.date } : undefined;
    case "datetime":
      return draft.datetime
        ? { kind: "datetime", datetime: draft.datetime }
        : undefined;
    case "url":
      return draft.url ? { kind: "url", url: draft.url } : undefined;
    case "single_select":
      return draft.optionKey
        ? { kind: "single_select", option_key: draft.optionKey }
        : undefined;
    case "multi_select":
      return draft.optionKeys && draft.optionKeys.length > 0
        ? { kind: "multi_select", option_keys: draft.optionKeys }
        : undefined;
    case "user_reference":
      if (draft.userIds && draft.userIds.length > 0) {
        return { kind: "user_reference", user_ids: draft.userIds };
      }
      return draft.userId
        ? { kind: "user_reference", user_id: draft.userId }
        : undefined;
    case "resource_reference":
      if (draft.resourceIds && draft.resourceIds.length > 0) {
        return {
          kind: "resource_reference",
          resource_ids: draft.resourceIds,
        };
      }
      return draft.resourceId
        ? { kind: "resource_reference", resource_id: draft.resourceId }
        : undefined;
  }
}

export function customFieldWritesFromValues(
  definitions: readonly CustomFieldDefinition[],
  values: Record<string, CustomFieldValue | undefined>
): CustomFieldWrite[] {
  return definitions.flatMap((definition) => {
    const value = values[definition.id];
    if (!value) {
      return [];
    }
    return [{ definition_id: definition.id, value }];
  });
}

export function missingRequiredCustomFieldNames(
  definitions: readonly CustomFieldDefinition[],
  values: Record<string, CustomFieldValue | undefined>
): string[] {
  return definitions
    .filter(
      (definition) =>
        definition.required && !definition.archived && !values[definition.id]
    )
    .map((definition) => definition.name);
}
