import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";

import { DatePicker } from "@/components/ui/date-picker";
import {
  EntityMultiSelect,
  SearchableEntitySelect,
} from "@/components/ui/entity-select";
import type { EntitySelectOption } from "@/components/ui/entity-select";
import { Input } from "@/components/ui/input";
import { InputGroupInput } from "@/components/ui/input-group";
import { propertyControlClassName } from "@/components/ui/property-list";
import {
  RemovableInputGroup,
  RemovableInputGroupRemove,
} from "@/components/ui/removable-input-group";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { v1SearchGet } from "@/lib/api/sdk";
import type {
  CustomFieldDefinition,
  CustomFieldOption,
  CustomFieldValue,
  OrganizationMember,
} from "@/lib/api/types";
import {
  customFieldValueFromDraft,
  customFieldValuesEqual,
  datetimeLocalToUtcIso,
  decimalInputStep,
  isMultiValued,
  parseIntegerDraft,
  resourceIdsFromValue,
  textDraftFromValue,
  userIdsFromValue,
  utcIsoToDatetimeLocal,
} from "@/lib/custom-fields/value";
import { SEARCH_RESOURCE_TYPES } from "@/lib/search/result";
import type { SearchResourceType } from "@/lib/search/result";
import { cn, getInitials } from "@/lib/utils";
import { personDisplayName } from "@/lib/work/resolve-work-people";

const booleanItems = { yes: "Yes", no: "No" };
const emptyPlaceholder = "—";
const emptySelectKey = "__empty__";

function withEmptySelectItem(
  required: boolean,
  items: Record<string, string>
): Record<string, string> {
  if (required) {
    return items;
  }
  return { [emptySelectKey]: emptyPlaceholder, ...items };
}

function formatDateInput(date: Date | null): string {
  if (!date) {
    return "";
  }
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function parseDateInput(value: string | undefined): Date | null {
  if (!value) {
    return null;
  }
  const parsed = new Date(`${value}T00:00:00`);
  return Number.isNaN(parsed.getTime()) ? null : parsed;
}

function memberOptions(members: OrganizationMember[]): EntitySelectOption[] {
  return members.map((member) => ({
    value: member.id,
    title: personDisplayName(member),
    avatarSrc: member.picture,
    avatarFallback: getInitials(member.first_name, member.last_name),
  }));
}

function mergeSelectedOptions(
  selectedIds: string[],
  catalog: EntitySelectOption[]
): EntitySelectOption[] {
  const byId = new Map<string, EntitySelectOption>();
  for (const option of catalog) {
    byId.set(option.value, option);
  }
  for (const id of selectedIds) {
    if (id && !byId.has(id)) {
      byId.set(id, {
        value: id,
        title: id,
        avatarFallback: id.slice(0, 2).toUpperCase(),
      });
    }
  }
  const selected = selectedIds
    .map((id) => byId.get(id))
    .filter((option): option is EntitySelectOption => option != null);
  const rest = catalog.filter((option) => !selectedIds.includes(option.value));
  return [...selected, ...rest];
}

function visibleSelectOptions(
  options: CustomFieldOption[],
  storedKeys: readonly string[]
): CustomFieldOption[] {
  const stored = new Set(storedKeys);
  return options.filter((option) => !option.disabled || stored.has(option.key));
}

function ResourceReferenceSelect({
  value,
  multiple,
  allowedTypes,
  disabled,
  size = "default",
  triggerClassName,
  emptyOption,
  onChange,
}: {
  value: string[];
  multiple: boolean;
  allowedTypes: string[];
  disabled?: boolean;
  size?: "sm" | "default";
  triggerClassName?: string;
  emptyOption?: string;
  onChange: (ids: string[]) => void;
}) {
  const searchTypes = allowedTypes.filter((type): type is SearchResourceType =>
    SEARCH_RESOURCE_TYPES.includes(type as SearchResourceType)
  );
  const canSearch = searchTypes.length > 0;
  const { data } = useQuery({
    queryKey: ["custom-field-resource-search", searchTypes],
    queryFn: async ({ signal }) => {
      const { data: searchPage } = await v1SearchGet({
        query: {
          types: [...searchTypes],
          page_size: 20,
        },
        signal,
        throwOnError: true,
      });
      return searchPage;
    },
    enabled: !disabled && canSearch,
  });

  const options: EntitySelectOption[] = useMemo(() => {
    const hits = data?.items ?? [];
    return mergeSelectedOptions(
      value,
      hits.map((hit) => ({
        value: `${hit.type}:${hit.id}`,
        title: hit.key ? `${hit.key} ${hit.title}` : hit.title,
        description: hit.subtitle,
      }))
    );
  }, [data, value]);

  const pickerDisabled = disabled || !canSearch;

  if (multiple) {
    return (
      <EntityMultiSelect
        options={options}
        value={value}
        disabled={pickerDisabled}
        size={size}
        triggerClassName={triggerClassName}
        placeholder={emptyPlaceholder}
        emptyOption={emptyOption}
        onValueChange={onChange}
      />
    );
  }

  return (
    <SearchableEntitySelect
      options={options}
      value={value[0]}
      disabled={pickerDisabled}
      size={size}
      triggerClassName={triggerClassName}
      placeholder={emptyPlaceholder}
      emptyOption={emptyOption}
      onValueChange={(next) => onChange(next ? [next] : [])}
    />
  );
}

export function CustomFieldEditor({
  definition,
  value,
  disabled,
  members = [],
  variant = "default",
  onCommit,
}: {
  definition: CustomFieldDefinition;
  value?: CustomFieldValue;
  disabled?: boolean;
  members?: OrganizationMember[];
  variant?: "default" | "sidebar";
  onCommit: (next: CustomFieldValue | undefined) => void;
}) {
  const [draft, setDraft] = useState(() =>
    textDraftFromValue(value, definition.kind)
  );

  useEffect(() => {
    setDraft(textDraftFromValue(value, definition.kind));
  }, [value, definition.id, definition.kind]);

  const storedSelectKeys =
    value?.kind === "single_select"
      ? [value.option_key]
      : value?.kind === "multi_select"
        ? value.option_keys
        : [];
  const selectOptions =
    definition.schema.kind === "single_select" ||
    definition.schema.kind === "multi_select"
      ? visibleSelectOptions(definition.schema.options, storedSelectKeys)
      : [];
  const multi = isMultiValued(definition.kind, definition.schema);
  const storedUserIds = userIdsFromValue(value);
  const storedResourceIds = resourceIdsFromValue(value);
  const selectedUserIds =
    multi && storedUserIds.length > 0
      ? storedUserIds
      : [draft || storedUserIds[0]].filter((id): id is string => Boolean(id));
  const users = mergeSelectedOptions(selectedUserIds, memberOptions(members));
  const sidebar = variant === "sidebar";
  const size = sidebar ? "sm" : "default";
  const controlClassName = sidebar ? propertyControlClassName : undefined;
  const dateClassName = sidebar
    ? cn(propertyControlClassName, "justify-start")
    : undefined;
  const selectContentAlign = sidebar
    ? ({ align: "start" as const, alignItemWithTrigger: false } as const)
    : undefined;

  const revertDraft = () => {
    setDraft(textDraftFromValue(value, definition.kind));
  };

  const commitValue = (next: CustomFieldValue | undefined) => {
    if (!next && definition.required) {
      revertDraft();
      return;
    }
    if (customFieldValuesEqual(value, next)) {
      return;
    }
    onCommit(next);
  };

  const commitDraft = (
    nextDraft: typeof draft,
    extra?: Parameters<typeof customFieldValueFromDraft>[1]
  ) => {
    setDraft(nextDraft);
    commitValue(
      customFieldValueFromDraft(definition.kind, {
        text: nextDraft,
        integer: nextDraft,
        decimal: nextDraft,
        date: nextDraft,
        datetime: nextDraft,
        url: nextDraft,
        optionKey: nextDraft,
        userId: nextDraft,
        resourceId: nextDraft,
        ...extra,
      })
    );
  };

  const commitIntegerDraft = (nextDraft: string) => {
    setDraft(nextDraft);
    const trimmed = nextDraft.trim();
    if (!trimmed) {
      commitValue(undefined);
      return;
    }
    const integer = parseIntegerDraft(trimmed);
    if (integer === undefined) {
      revertDraft();
      return;
    }
    commitValue({ kind: "integer", integer });
  };

  switch (definition.kind) {
    case "boolean": {
      const selected =
        value?.kind === "boolean" ? (value.boolean ? "yes" : "no") : undefined;
      return (
        <Select
          value={selected}
          onValueChange={(next) => {
            if (!next || next === emptySelectKey) {
              commitValue(undefined);
              return;
            }
            commitValue(
              customFieldValueFromDraft("boolean", {
                boolean: next === "yes",
              })
            );
          }}
          disabled={disabled}
          items={withEmptySelectItem(definition.required, booleanItems)}
        >
          <SelectTrigger
            size={size}
            className={cn("w-full", controlClassName)}
            aria-label={definition.name}
          >
            <SelectValue placeholder={emptyPlaceholder} />
          </SelectTrigger>
          <SelectContent {...selectContentAlign}>
            {definition.required ? null : (
              <SelectItem value={emptySelectKey}>{emptyPlaceholder}</SelectItem>
            )}
            <SelectItem value="yes">Yes</SelectItem>
            <SelectItem value="no">No</SelectItem>
          </SelectContent>
        </Select>
      );
    }
    case "date":
      return (
        <DatePicker
          date={parseDateInput(value?.kind === "date" ? value.date : draft)}
          disabled={disabled}
          clearable={!definition.required}
          placeholder={emptyPlaceholder}
          aria-label={definition.name}
          className={dateClassName}
          onDateChange={(date) => commitDraft(formatDateInput(date))}
        />
      );
    case "datetime": {
      const iso = value?.kind === "datetime" ? value.datetime : draft;
      const local = iso ? utcIsoToDatetimeLocal(iso) : "";
      const clearable = !definition.required;
      const DatetimeInput = clearable ? InputGroupInput : Input;
      const field = (
        <div className={cn("relative", clearable && "min-w-0 flex-1")}>
          <DatetimeInput
            type="datetime-local"
            aria-label={definition.name}
            disabled={disabled}
            className={cn(
              !clearable && controlClassName,
              clearable && "h-full",
              !local &&
                "text-transparent [&::-webkit-calendar-picker-indicator]:opacity-100 [&::-webkit-datetime-edit]:invisible"
            )}
            value={local}
            onChange={(event) => {
              const raw = event.target.value;
              commitDraft(raw ? datetimeLocalToUtcIso(raw) : "");
            }}
          />
          {local ? null : (
            <span
              className={cn(
                "text-muted-foreground pointer-events-none absolute inset-y-0 left-0 flex items-center text-sm font-medium",
                sidebar ? "px-2" : "px-3"
              )}
            >
              {emptyPlaceholder}
            </span>
          )}
        </div>
      );
      if (!clearable) {
        return field;
      }
      return (
        <RemovableInputGroup
          size={size}
          className={cn(size === "sm" ? "h-8" : "h-9", controlClassName)}
        >
          {field}
          {local ? (
            <RemovableInputGroupRemove
              addonClassName="pr-0.5"
              disabled={disabled}
              aria-label={`Clear ${definition.name}`}
              title={`Clear ${definition.name}`}
              onClick={() => commitDraft("")}
            />
          ) : null}
        </RemovableInputGroup>
      );
    }
    case "single_select": {
      const selected =
        value?.kind === "single_select" ? value.option_key : draft || undefined;
      const optionItems = Object.fromEntries(
        selectOptions.map((option) => [option.key, option.label])
      );
      return (
        <Select
          value={selected}
          onValueChange={(next) => {
            if (!next || next === emptySelectKey) {
              commitDraft("");
              return;
            }
            commitDraft(next);
          }}
          disabled={disabled}
          items={withEmptySelectItem(definition.required, optionItems)}
        >
          <SelectTrigger
            size={size}
            className={cn("w-full", controlClassName)}
            aria-label={definition.name}
          >
            <SelectValue placeholder={emptyPlaceholder} />
          </SelectTrigger>
          <SelectContent {...selectContentAlign}>
            {definition.required ? null : (
              <SelectItem value={emptySelectKey}>{emptyPlaceholder}</SelectItem>
            )}
            {selectOptions.map((option) => (
              <SelectItem
                key={option.key}
                value={option.key}
                disabled={option.disabled}
              >
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      );
    }
    case "multi_select":
      return (
        <EntityMultiSelect
          options={selectOptions.map((option) => ({
            value: option.key,
            title: option.label,
          }))}
          value={value?.kind === "multi_select" ? value.option_keys : []}
          disabled={disabled}
          size={size}
          triggerClassName={controlClassName}
          placeholder={emptyPlaceholder}
          emptyOption={definition.required ? undefined : emptyPlaceholder}
          onValueChange={(keys) =>
            commitValue(
              customFieldValueFromDraft("multi_select", { optionKeys: keys })
            )
          }
        />
      );
    case "user_reference":
      if (multi) {
        return (
          <EntityMultiSelect
            options={users}
            value={selectedUserIds}
            disabled={disabled}
            size={size}
            triggerClassName={controlClassName}
            placeholder={emptyPlaceholder}
            emptyOption={definition.required ? undefined : emptyPlaceholder}
            onValueChange={(ids) =>
              commitValue(
                customFieldValueFromDraft("user_reference", { userIds: ids })
              )
            }
          />
        );
      }
      return (
        <SearchableEntitySelect
          options={users}
          value={selectedUserIds[0]}
          disabled={disabled}
          size={size}
          triggerClassName={controlClassName}
          placeholder={emptyPlaceholder}
          emptyOption={definition.required ? undefined : emptyPlaceholder}
          onValueChange={(id) => commitDraft(id ?? "")}
        />
      );
    case "resource_reference": {
      const allowed =
        definition.schema.kind === "resource_reference"
          ? definition.schema.allowed_types
          : [];
      const current =
        multi && storedResourceIds.length > 0
          ? storedResourceIds
          : [draft || storedResourceIds[0]].filter((id): id is string =>
              Boolean(id)
            );
      return (
        <ResourceReferenceSelect
          value={current}
          multiple={multi}
          allowedTypes={allowed}
          disabled={disabled}
          size={size}
          triggerClassName={controlClassName}
          emptyOption={definition.required ? undefined : emptyPlaceholder}
          onChange={(ids) => {
            if (multi) {
              commitValue(
                customFieldValueFromDraft("resource_reference", {
                  resourceIds: ids,
                })
              );
              return;
            }
            commitDraft(ids[0] ?? "");
          }}
        />
      );
    }
    case "integer": {
      const min =
        definition.schema.kind === "integer"
          ? definition.schema.min
          : undefined;
      const max =
        definition.schema.kind === "integer"
          ? definition.schema.max
          : undefined;
      return (
        <Input
          type="number"
          step="1"
          min={min}
          max={max}
          aria-label={definition.name}
          disabled={disabled}
          className={controlClassName}
          placeholder={emptyPlaceholder}
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          onBlur={() => commitIntegerDraft(draft)}
        />
      );
    }
    case "decimal": {
      const min =
        definition.schema.kind === "decimal"
          ? definition.schema.min
          : undefined;
      const max =
        definition.schema.kind === "decimal"
          ? definition.schema.max
          : undefined;
      const scale =
        definition.schema.kind === "decimal"
          ? definition.schema.scale
          : undefined;
      return (
        <Input
          type="number"
          step={decimalInputStep(scale)}
          min={min}
          max={max}
          aria-label={definition.name}
          disabled={disabled}
          className={controlClassName}
          placeholder={emptyPlaceholder}
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          onBlur={() => commitDraft(draft)}
        />
      );
    }
    default:
      return (
        <Input
          aria-label={definition.name}
          disabled={disabled}
          className={controlClassName}
          placeholder={emptyPlaceholder}
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          onBlur={() => commitDraft(draft)}
        />
      );
  }
}
