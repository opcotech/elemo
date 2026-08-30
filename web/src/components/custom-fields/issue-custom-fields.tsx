import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { CustomFieldEditor } from "@/components/custom-fields/field-editor";
import { CustomFieldValueDisplay } from "@/components/custom-fields/value-display";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { PropertyList } from "@/components/ui/property-list";
import { Section } from "@/components/ui/section";
import { Skeleton } from "@/components/ui/skeleton";
import {
  v1ResourceCustomFieldValueDeleteMutation,
  v1ResourceCustomFieldValuePutMutation,
} from "@/lib/api/mutation-options";
import { v1ResourceCustomFieldsGetOptions } from "@/lib/api/query-options";
import type { CustomFieldValue } from "@/lib/api/types";
import { customFieldValuesEqual } from "@/lib/custom-fields/value";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import { useOrganizationMembersForNamespace } from "@/lib/work/use-organization-members-for-namespace";

export function IssueCustomFields({
  issueId,
  namespaceId,
  disabled,
  mode,
}: {
  issueId: string;
  namespaceId?: string;
  disabled?: boolean;
  mode: "edit" | "readonly";
}) {
  const queryClient = useQueryClient();
  const queryOptions = v1ResourceCustomFieldsGetOptions({
    path: { resourceType: "Issue", id: issueId },
  });
  const { data, isPending, isError, error } = useQuery(queryOptions);
  const { members } = useOrganizationMembersForNamespace(namespaceId, {
    enabled: Boolean(namespaceId),
  });
  const putMutation = useMutation(v1ResourceCustomFieldValuePutMutation());
  const deleteMutation = useMutation(
    v1ResourceCustomFieldValueDeleteMutation()
  );
  const [editorReset, setEditorReset] = useState(0);

  const entries = (data ?? []).filter((entry) => !entry.definition.archived);

  const saveValue = async (
    definitionId: string,
    required: boolean,
    current: CustomFieldValue | undefined,
    next: CustomFieldValue | undefined
  ) => {
    if (customFieldValuesEqual(current, next)) {
      return;
    }
    if (!next && required) {
      setEditorReset((currentReset) => currentReset + 1);
      return;
    }
    try {
      if (!next) {
        await deleteMutation.mutateAsync({
          path: {
            resourceType: "Issue",
            id: issueId,
            definitionId,
          },
        });
        showSuccessToast(
          "Custom field cleared",
          "The stored value was removed"
        );
      } else {
        await putMutation.mutateAsync({
          path: {
            resourceType: "Issue",
            id: issueId,
            definitionId,
          },
          body: next,
        });
        showSuccessToast("Custom field updated", "The value was saved");
      }
      await queryClient.invalidateQueries({ queryKey: queryOptions.queryKey });
    } catch (saveError) {
      setEditorReset((currentReset) => currentReset + 1);
      showErrorToast(
        "Failed to update custom field",
        saveError instanceof Error ? saveError : "Unknown error"
      );
    }
  };

  if (isPending) {
    return (
      <Section title="Custom Fields" data-section="issue-custom-fields">
        <Skeleton
          className={mode === "readonly" ? "h-16 w-full" : "h-24 w-full"}
        />
      </Section>
    );
  }

  if (isError) {
    return (
      <Section title="Custom Fields" data-section="issue-custom-fields">
        <Alert variant="destructive">
          <AlertTitle>Failed to load custom fields</AlertTitle>
          <AlertDescription>
            {error instanceof Error
              ? error.message
              : "Custom fields could not be loaded. Please try again later."}
          </AlertDescription>
        </Alert>
      </Section>
    );
  }

  if (entries.length === 0) {
    return null;
  }

  return (
    <Section title="Custom Fields" data-section="issue-custom-fields">
      <PropertyList
        compact={mode === "readonly"}
        items={entries.map((entry) => {
          const readonly = mode === "readonly";
          return {
            id: entry.definition.id,
            label: entry.definition.name,
            value: (
              <div data-custom-field-key={entry.definition.key}>
                {readonly ? (
                  <CustomFieldValueDisplay
                    definition={entry.definition}
                    value={entry.value}
                    members={members}
                  />
                ) : (
                  <CustomFieldEditor
                    key={`${issueId}-${entry.definition.id}-${editorReset}`}
                    definition={entry.definition}
                    value={entry.value}
                    variant="sidebar"
                    disabled={
                      disabled ||
                      putMutation.isPending ||
                      deleteMutation.isPending
                    }
                    members={members}
                    onCommit={(next) => {
                      void saveValue(
                        entry.definition.id,
                        entry.definition.required,
                        entry.value,
                        next
                      );
                    }}
                  />
                )}
              </div>
            ),
          };
        })}
      />
    </Section>
  );
}
