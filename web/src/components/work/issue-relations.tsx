import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { ChevronDownIcon, Link2Icon, PlusIcon, XIcon } from "lucide-react";
import { useState } from "react";

import { IssueRelationAddDialog } from "./issue-relation-add-dialog";
import { IssueSelectDetails } from "./issue-select-option";

import { AppList, EntityIcon } from "@/components/shared/entity-link";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { EmptyState } from "@/components/ui/empty-state";
import { InternalLink } from "@/components/ui/internal-link";
import {
  Item,
  ItemActions,
  ItemContent,
  ItemMedia,
  ItemTitle,
} from "@/components/ui/item";
import { Section } from "@/components/ui/section";
import { Skeleton } from "@/components/ui/skeleton";
import {
  v1IssueRelationDeleteMutation,
  v1IssueRelationUpdateMutation,
} from "@/lib/api/mutation-options";
import { v1IssueRelationsGetOptions } from "@/lib/api/query-options";
import type { IssueRelation } from "@/lib/api/types";
import { internalPath } from "@/lib/internal-url";
import { runMutationSuccessWorkflow } from "@/lib/mutation-workflow";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import type { EditableIssueRelationKind } from "@/lib/work/issue-relations";
import {
  ISSUE_RELATIONS_PAGE_SIZE,
  ISSUE_RELATIONS_PREVIEW_PAGE_SIZE,
  issueRelationDisplayKind,
  issueRelationInvalidationKeys,
  issueRelationKindLabel,
  issueRelationKindSelectValues,
  relatedIssueIds,
  relatedIssueWorkPath,
  relationKindPatch,
  visibleIssueRelations,
} from "@/lib/work/issue-relations";

function useIssueRelationQueries(issueId: string, pageSize: number) {
  return useQuery(
    v1IssueRelationsGetOptions({
      path: { id: issueId },
      query: { page_size: pageSize },
    })
  );
}

function useIssueRelationMutations({
  issueId,
  organizationId,
  namespaceId,
  issueKey,
}: {
  issueId: string;
  organizationId: string;
  namespaceId: string;
  issueKey?: string;
}) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const updateMutation = useMutation({ ...v1IssueRelationUpdateMutation() });
  const deleteMutation = useMutation({ ...v1IssueRelationDeleteMutation() });

  const invalidateFor = (related?: IssueRelation["related"] | null) =>
    runMutationSuccessWorkflow({
      invalidateQueries: issueRelationInvalidationKeys({
        issueId,
        organizationId,
        namespaceId,
        issueKey,
        related,
      }).map((queryKey) => () => queryClient.invalidateQueries({ queryKey })),
      invalidateRouter: () => router.invalidate(),
    });

  const updateKind = async (
    relation: IssueRelation,
    kind: EditableIssueRelationKind
  ) => {
    if (updateMutation.isPending || deleteMutation.isPending) {
      return;
    }

    try {
      const updated = await updateMutation.mutateAsync({
        path: { id: issueId, relation_id: relation.id },
        body: { kind },
      });
      await invalidateFor(updated.related);
      showSuccessToast("Relation updated", "Relation kind was changed");
    } catch (error) {
      showErrorToast(
        "Failed to update relation",
        error instanceof Error ? error.message : "Unknown error occurred"
      );
    }
  };

  const removeRelation = async (relation: IssueRelation) => {
    if (updateMutation.isPending || deleteMutation.isPending) {
      return;
    }

    try {
      await deleteMutation.mutateAsync({
        path: { id: issueId, relation_id: relation.id },
      });
      await invalidateFor(relation.related);
      showSuccessToast("Relation removed", "The related issue was unlinked");
    } catch (error) {
      showErrorToast(
        "Failed to remove relation",
        error instanceof Error ? error.message : "Unknown error occurred"
      );
    }
  };

  return {
    updateKind,
    removeRelation,
    isPending: updateMutation.isPending || deleteMutation.isPending,
  };
}

function IssueRelationItem({
  relation,
  organizationSlug,
  namespaceSlug,
  onUpdateKind,
  onRemove,
  disabled = false,
}: {
  relation: IssueRelation;
  organizationSlug: string;
  namespaceSlug: string;
  onUpdateKind?: (kind: EditableIssueRelationKind) => void;
  onRemove?: () => void;
  disabled?: boolean;
}) {
  const displayKind = issueRelationDisplayKind(
    relation.kind,
    relation.direction
  );
  const kindLabel = issueRelationKindLabel(displayKind);
  const href = relatedIssueWorkPath(relation.related, {
    organizationSlug,
    namespaceSlug,
  });

  return (
    <Item
      role="listitem"
      size="sm"
      className="group/entity min-w-0 flex-nowrap p-0"
    >
      <InternalLink
        to={internalPath(href)}
        className="text-foreground flex min-w-0 flex-1 items-center gap-2.5 px-3 py-2.5 hover:no-underline"
      >
        <ItemMedia
          variant="icon"
          className="bg-muted text-muted-foreground size-8 rounded-lg"
        >
          <EntityIcon type="work-item" />
        </ItemMedia>
        <ItemContent className="min-w-0">
          <ItemTitle className="group-hover/entity:text-primary block max-w-full truncate">
            {relation.related.key} {relation.related.title}
          </ItemTitle>
          <IssueSelectDetails issue={relation.related} />
        </ItemContent>
      </InternalLink>
      <ItemActions className="shrink-0 pr-3">
        {onUpdateKind ? (
          <DropdownMenu disabled={disabled}>
            <DropdownMenuTrigger
              render={
                <Button
                  variant="ghost"
                  size="xs"
                  aria-label={`Relation kind for ${relation.related.key}`}
                />
              }
            >
              {kindLabel}
              <ChevronDownIcon className="size-3" />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuRadioGroup
                value={displayKind}
                onValueChange={(next: string) => {
                  const kind = relationKindPatch(displayKind, next);
                  if (!kind) {
                    return;
                  }
                  onUpdateKind(kind);
                }}
              >
                {issueRelationKindSelectValues().map((kind) => (
                  <DropdownMenuRadioItem key={kind} value={kind}>
                    {issueRelationKindLabel(kind)}
                  </DropdownMenuRadioItem>
                ))}
              </DropdownMenuRadioGroup>
            </DropdownMenuContent>
          </DropdownMenu>
        ) : (
          <span className="text-muted-foreground text-xs">{kindLabel}</span>
        )}
        {onRemove ? (
          <Button
            type="button"
            variant="ghost"
            size="icon-xs"
            disabled={disabled}
            aria-label={`Remove relation to ${relation.related.key}`}
            title="Remove relation"
            className="hover:text-destructive hover:bg-destructive/10 hover:ring-0"
            onClick={onRemove}
          >
            <XIcon />
          </Button>
        ) : null}
      </ItemActions>
    </Item>
  );
}

function IssueRelationList({
  relations,
  organizationSlug,
  namespaceSlug,
  onUpdateKind,
  onRemove,
  disabled,
}: {
  relations: readonly IssueRelation[];
  organizationSlug: string;
  namespaceSlug: string;
  onUpdateKind?: (
    relation: IssueRelation,
    kind: EditableIssueRelationKind
  ) => void;
  onRemove?: (relation: IssueRelation) => void;
  disabled?: boolean;
}) {
  return (
    <AppList>
      {relations.map((item) => (
        <IssueRelationItem
          key={item.id}
          relation={item}
          organizationSlug={organizationSlug}
          namespaceSlug={namespaceSlug}
          onUpdateKind={
            onUpdateKind
              ? (kind) => {
                  onUpdateKind(item, kind);
                }
              : undefined
          }
          onRemove={
            onRemove
              ? () => {
                  onRemove(item);
                }
              : undefined
          }
          disabled={disabled}
        />
      ))}
    </AppList>
  );
}

interface IssueRelationsProps {
  issueId: string;
  issueKey: string;
  organizationId: string;
  organizationSlug: string;
  namespaceId: string;
  namespaceSlug: string;
  disabled?: boolean;
}

export function IssueRelations({
  issueId,
  issueKey,
  organizationId,
  organizationSlug,
  namespaceId,
  namespaceSlug,
  disabled = false,
}: IssueRelationsProps) {
  const [addOpen, setAddOpen] = useState(false);
  const { data, isLoading } = useIssueRelationQueries(
    issueId,
    ISSUE_RELATIONS_PAGE_SIZE
  );
  const { updateKind, removeRelation, isPending } = useIssueRelationMutations({
    issueId,
    organizationId,
    namespaceId,
    issueKey,
  });

  const relations = visibleIssueRelations(data?.items ?? []);
  const relatedIds = relatedIssueIds(data?.items ?? []);
  const controlsDisabled = disabled || isPending;

  const addButton = (
    <Button
      type="button"
      variant="ghost"
      size="xs"
      disabled={controlsDisabled}
      onClick={() => setAddOpen(true)}
    >
      <PlusIcon />
      Add
    </Button>
  );

  return (
    <Section
      title="Relations"
      data-section="issue-relations"
      action={addButton}
    >
      {isLoading ? (
        <div className="space-y-2">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </div>
      ) : relations.length === 0 ? (
        <EmptyState
          compact
          icon={<Link2Icon />}
          title="No relations"
          description="Related work will appear here."
        />
      ) : (
        <IssueRelationList
          relations={relations}
          organizationSlug={organizationSlug}
          namespaceSlug={namespaceSlug}
          onUpdateKind={(relation, kind) => {
            void updateKind(relation, kind);
          }}
          onRemove={(relation) => {
            void removeRelation(relation);
          }}
          disabled={controlsDisabled}
        />
      )}
      <IssueRelationAddDialog
        issueId={issueId}
        issueKey={issueKey}
        organizationId={organizationId}
        namespaceId={namespaceId}
        relatedIds={relatedIds}
        open={addOpen}
        onOpenChange={setAddOpen}
      />
    </Section>
  );
}

interface IssueRelationsPreviewProps {
  issueId: string;
  organizationSlug: string;
  namespaceSlug: string;
}

export function IssueRelationsPreview({
  issueId,
  organizationSlug,
  namespaceSlug,
}: IssueRelationsPreviewProps) {
  const { data, isLoading } = useIssueRelationQueries(
    issueId,
    ISSUE_RELATIONS_PREVIEW_PAGE_SIZE
  );
  const relations = visibleIssueRelations(data?.items ?? []).slice(
    0,
    ISSUE_RELATIONS_PREVIEW_PAGE_SIZE
  );

  if (isLoading) {
    return (
      <Section title="Relations" data-section="issue-relations">
        <div className="space-y-2">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </div>
      </Section>
    );
  }

  return (
    <Section title="Relations" data-section="issue-relations">
      {relations.length > 0 ? (
        <IssueRelationList
          relations={relations}
          organizationSlug={organizationSlug}
          namespaceSlug={namespaceSlug}
        />
      ) : (
        <EmptyState
          compact
          icon={<Link2Icon />}
          title="No relations"
          description="Related work will appear here."
        />
      )}
    </Section>
  );
}
