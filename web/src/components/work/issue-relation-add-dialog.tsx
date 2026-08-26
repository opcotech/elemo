import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { issueToSelectOption } from "./issue-select-option";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { DialogForm } from "@/components/ui/dialog-form";
import {
  EntitySelect,
  SearchableEntitySelect,
} from "@/components/ui/entity-select";
import {
  ControlledField,
  Field,
  FieldControl,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Skeleton } from "@/components/ui/skeleton";
import { v1IssueRelationsCreateMutation } from "@/lib/api/mutation-options";
import type { PartialIssue } from "@/lib/api/types";
import { runMutationSuccessWorkflow } from "@/lib/mutation-workflow";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import {
  editableIssueRelationKinds,
  filterAvailableRelatedIssues,
  issueRelationInvalidationKeys,
  issueRelationKindLabel,
  relatedIssueCatalogQueryOptions,
} from "@/lib/work/issue-relations";

const relationFormSchema = z.object({
  relatedId: z.string().min(1, "Issue is required"),
  kind: z.enum(editableIssueRelationKinds),
});

type RelationFormValues = z.infer<typeof relationFormSchema>;

const kindOptions = editableIssueRelationKinds.map((kind) => ({
  value: kind,
  title: issueRelationKindLabel(kind),
}));

interface IssueRelationAddDialogProps {
  issueId: string;
  issueKey: string;
  organizationId: string;
  namespaceId: string;
  relatedIds: ReadonlySet<string>;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function IssueRelationAddDialog({
  issueId,
  issueKey,
  organizationId,
  namespaceId,
  relatedIds,
  open,
  onOpenChange,
}: IssueRelationAddDialogProps) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const [error, setError] = useState<Error | null>(null);

  const form = useForm<RelationFormValues>({
    resolver: zodResolver(relationFormSchema),
    defaultValues: {
      relatedId: "",
      kind: "related to",
    },
  });

  const { data: issuesPage, isLoading } = useQuery({
    ...relatedIssueCatalogQueryOptions(organizationId, namespaceId),
    enabled: open && Boolean(organizationId) && Boolean(namespaceId),
  });

  const availableIssues = filterAvailableRelatedIssues<PartialIssue>(
    issuesPage?.items ?? [],
    issueId,
    relatedIds
  );

  const issueOptions = availableIssues.map(issueToSelectOption);

  const mutation = useMutation({
    ...v1IssueRelationsCreateMutation(),
    onSuccess: (created) =>
      runMutationSuccessWorkflow({
        invalidateQueries: issueRelationInvalidationKeys({
          issueId,
          organizationId,
          namespaceId,
          issueKey,
          related: created.related,
        }).map((queryKey) => () => queryClient.invalidateQueries({ queryKey })),
        invalidateRouter: () => router.invalidate(),
        callbacks: [
          () => {
            setError(null);
            showSuccessToast("Relation added", "The related issue was linked");
            form.reset();
            onOpenChange(false);
          },
        ],
      }),
    onError: (err) => {
      setError(new Error(err.message));
      showErrorToast("Failed to add relation", err.message);
    },
  });

  useEffect(() => {
    if (open) {
      setError(null);
    }
  }, [open]);

  const onSubmit = (values: RelationFormValues) => {
    setError(null);
    mutation.mutate({
      path: { id: issueId },
      body: {
        related_id: values.relatedId,
        kind: values.kind,
      },
    });
  };

  return (
    <DialogForm
      form={form}
      open={open}
      onOpenChange={onOpenChange}
      title="Add relation"
      data-section="issue-relation-add-form"
      onSubmit={form.handleSubmit(onSubmit)}
      isPending={mutation.isPending}
      error={error}
      submitButtonText="Add relation"
      onReset={() => form.reset()}
      className="sm:max-w-125"
    >
      {isLoading ? (
        <div className="space-y-4">
          <div className="space-y-2">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-10 w-full" />
          </div>
          <div className="space-y-2">
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-10 w-full" />
          </div>
        </div>
      ) : (
        <>
          <ControlledField
            control={form.control}
            name="kind"
            render={({ field }) => (
              <Field>
                <FieldLabel>Kind</FieldLabel>
                <FieldControl>
                  <EntitySelect
                    options={kindOptions}
                    value={field.value}
                    onValueChange={field.onChange}
                    placeholder="Choose a relation kind"
                    disabled={mutation.isPending}
                  />
                </FieldControl>
                <FieldError />
              </Field>
            )}
          />
          {availableIssues.length === 0 ? (
            <Alert>
              <AlertDescription>
                {relatedIds.size > 0
                  ? "All other issues in this namespace are already related."
                  : "No other issues in this namespace are available to relate."}
              </AlertDescription>
            </Alert>
          ) : (
            <ControlledField
              control={form.control}
              name="relatedId"
              render={({ field }) => (
                <Field>
                  <FieldLabel>Issue</FieldLabel>
                  <FieldControl>
                    <SearchableEntitySelect
                      options={issueOptions}
                      value={field.value}
                      onValueChange={field.onChange}
                      placeholder="Choose an issue"
                      searchPlaceholder="Search issues…"
                      emptyMessage="No issues found."
                      disabled={mutation.isPending}
                    />
                  </FieldControl>
                  <FieldError />
                </Field>
              )}
            />
          )}
        </>
      )}
    </DialogForm>
  );
}
