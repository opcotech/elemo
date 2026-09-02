import { PersonAvatarStack } from "@/components/ui/person-avatar-stack";
import { formatDateTime, formatTargetDate } from "@/components/work/utils";
import type {
  CustomFieldDefinition,
  CustomFieldValue,
  OrganizationMember,
} from "@/lib/api/types";
import {
  customFieldOptionLabel,
  formatCustomFieldValue,
  isSafeHttpUrl,
  resourceIdsFromValue,
  userIdsFromValue,
} from "@/lib/custom-fields/value";
import { resolveWorkPeople } from "@/lib/work/resolve-work-people";

function EmptyValue() {
  return <span className="text-muted-foreground">—</span>;
}

export function CustomFieldValueDisplay({
  definition,
  value,
  members,
}: {
  definition: CustomFieldDefinition;
  value?: CustomFieldValue;
  members: OrganizationMember[];
}) {
  if (!value) {
    return <EmptyValue />;
  }

  switch (value.kind) {
    case "boolean":
      return value.boolean ? "Yes" : "No";
    case "date":
      return formatTargetDate(value.date);
    case "datetime":
      return formatDateTime(value.datetime);
    case "url":
      return isSafeHttpUrl(value.url) ? (
        <a
          href={value.url}
          className="text-primary block truncate underline-offset-4 hover:underline"
          target="_blank"
          rel="noopener noreferrer"
        >
          {value.url}
        </a>
      ) : (
        value.url
      );
    case "single_select":
      return customFieldOptionLabel(definition.schema, value.option_key);
    case "multi_select": {
      if (value.option_keys.length === 0) {
        return <EmptyValue />;
      }
      return value.option_keys
        .map((key) => customFieldOptionLabel(definition.schema, key))
        .join(", ");
    }
    case "user_reference":
      return (
        <PersonAvatarStack
          people={resolveWorkPeople(userIdsFromValue(value), members)}
          size="sm"
          showNames
          emptyLabel="—"
        />
      );
    case "resource_reference": {
      const ids = resourceIdsFromValue(value);
      return ids.length > 0 ? ids.join(", ") : <EmptyValue />;
    }
    default: {
      const formatted = formatCustomFieldValue(value);
      return formatted ? formatted : <EmptyValue />;
    }
  }
}
