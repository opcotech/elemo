import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ActivityIcon, ArrowRightIcon, FileTextIcon } from "lucide-react";
import { useMemo, useState } from "react";

import { DocumentCreateDialog } from "@/components/documents/document-create-dialog";
import { DocumentDeleteDialog } from "@/components/documents/document-delete-dialog";
import { DocumentLinkDialog } from "@/components/documents/document-link-dialog";
import {
  DocumentList,
  DocumentListToolbar,
} from "@/components/documents/document-list";
import { DocumentRenameDialog } from "@/components/documents/document-rename-dialog";
import { DocumentSummaryList } from "@/components/documents/document-summary-list";
import { ContentWidth } from "@/components/layout/content-width";
import { openQuickCreate } from "@/components/quick-create/open";
import { ActivityFeed } from "@/components/shared/activity-feed";
import { MockDataAlert } from "@/components/shared/app-feedback";
import {
  CursorPaginator,
  cursorPaginatorProps,
} from "@/components/shared/cursor-paginator";
import { EntityHeader, PageActions } from "@/components/shared/entity-header";
import { RelationList } from "@/components/shared/relation-list";
import { Button } from "@/components/ui/button";
import {
  AddButton,
  CreateButton,
  LinkButton,
} from "@/components/ui/create-button";
import { EmptyState } from "@/components/ui/empty-state";
import { InternalLink } from "@/components/ui/internal-link";
import { ListSkeleton } from "@/components/ui/list-skeleton";
import { Progress } from "@/components/ui/progress";
import { PropertyList } from "@/components/ui/property-list";
import { Section } from "@/components/ui/section";
import { StatusIndicator } from "@/components/ui/status-indicator";
import { CompactWorkList } from "@/components/work/work-list";
import { useCursorPageNav } from "@/hooks/use-cursor-page-nav";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1ProjectsDocumentsGetOptions } from "@/lib/api/query-options";
import {
  v1ProjectsDocumentsCreate,
  v1ProjectsDocumentsRelate,
  v1ProjectsDocumentsUnrelate,
} from "@/lib/api/sdk";
import type { Namespace, PartialDocument, Project } from "@/lib/api/types";
import { Action, can } from "@/lib/auth/permissions";
import { documentListQueryKey } from "@/lib/documents/create";
import {
  ALL_DOCUMENT_CREATORS,
  documentCreators,
  visibleDocuments,
} from "@/lib/documents/document-list";
import type { DocumentListSort } from "@/lib/documents/document-list";
import { invalidateDocumentQueries } from "@/lib/documents/document-queries";
import { internalPath } from "@/lib/internal-url";
import {
  selectActivity,
  selectRelations,
  selectWorkItems,
} from "@/lib/mock-data";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

export function ProjectOverviewPage({
  namespace,
  project,
}: {
  namespace: Namespace;
  project: Project;
}) {
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Project, project.id)
  );
  const mayCreateWork = can(permissions, Action.IssueCreate);
  const mayCreateDocument = can(permissions, Action.DocumentCreate);
  const [createOpen, setCreateOpen] = useState(false);
  const [linkOpen, setLinkOpen] = useState(false);
  const scopedWork = selectWorkItems({
    scope: {
      type: "project",
      namespaceId: namespace.id,
      projectId: project.id,
    },
  });
  const work = scopedWork;
  const projectRelations = selectRelations({
    entity: { id: work[0]?.id ?? "", type: "work-item" },
  });
  const issueCount = project.issue_count ?? 0;
  const { data: documentsPage, isLoading: isDocumentsLoading } = useQuery(
    v1ProjectsDocumentsGetOptions({
      path: { id: project.id },
      query: { page_size: 5 },
    })
  );
  const documents = documentsPage?.items ?? [];
  const relatedIds = new Set(documents.map((document) => document.id));
  const documentCount = project.document_count ?? 0;
  const progress = issueCount ? Math.min(92, 28 + issueCount * 9) : 0;

  return (
    <ContentWidth width="overview" className="space-y-7">
      <EntityHeader
        type="project"
        eyebrow={`${namespace.name} / ${project.key}`}
        title={project.name}
        description={
          project.description || "No project outcome has been described yet."
        }
        showIcon={false}
        actions={
          <PageActions
            primary={
              mayCreateWork ? (
                <CreateButton onClick={() => openQuickCreate("work")}>
                  Work
                </CreateButton>
              ) : (
                <Button disabled size="sm" title="Write permission required">
                  Create work
                </Button>
              )
            }
            secondary={[
              {
                label: "Open project work",
                href: `/namespaces/${namespace.id}/projects/${project.id}/work`,
              },
              {
                label: "View relationships",
                href: `/relations/project/${project.id}`,
              },
            ]}
          />
        }
      />

      <div className="grid gap-8 lg:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]">
        <div className="space-y-8">
          <Section title="Current state">
            <div className="rounded-xl border p-4">
              <div className="flex items-center gap-3">
                <Progress value={progress} className="flex-1" />
                <span className="text-muted-foreground text-sm tabular-nums">
                  {issueCount
                    ? `${progress}% illustrative`
                    : "No issue progress"}
                </span>
              </div>
              <p className="text-muted-foreground mt-3 text-sm">
                {issueCount
                  ? `${issueCount} linked issues are associated with this project.`
                  : "No linked issues are associated with this project yet."}
              </p>
              {issueCount > 0 && (
                <MockDataAlert title="Illustrative progress" className="mt-3">
                  Progress is estimated from linked issues and does not reflect
                  verified completion.
                </MockDataAlert>
              )}
            </div>
          </Section>

          <Section
            title="Active work"
            action={
              <Button
                variant="ghost"
                size="sm"
                render={
                  <InternalLink
                    to={internalPath(
                      `/namespaces/${namespace.id}/projects/${project.id}/work`
                    )}
                  />
                }
              >
                View all <ArrowRightIcon />
              </Button>
            }
          >
            <MockDataAlert title="Illustrative project work" className="mb-3">
              {scopedWork.length
                ? "Work detail is provided by centralized fixtures."
                : "No fixture work matches this project. Unrelated examples are not shown."}
            </MockDataAlert>
            <CompactWorkList items={work} limit={5} />
          </Section>

          <Section
            title="Documents"
            data-section="documents"
            action={
              <div className="flex items-center gap-2">
                {mayCreateDocument ? (
                  <>
                    <LinkButton size="sm" onClick={() => setLinkOpen(true)} />
                    <AddButton size="sm" onClick={() => setCreateOpen(true)} />
                  </>
                ) : null}
                <Button
                  variant="ghost"
                  size="sm"
                  render={
                    <InternalLink
                      to={internalPath(
                        `/namespaces/${namespace.id}/projects/${project.id}/documents`
                      )}
                    />
                  }
                >
                  View all <ArrowRightIcon />
                </Button>
              </div>
            }
          >
            {isDocumentsLoading ? (
              <ListSkeleton rows={4} />
            ) : documents.length > 0 ? (
              <DocumentSummaryList documents={documents} />
            ) : (
              <EmptyState
                compact
                icon={<FileTextIcon />}
                title="No related documents"
                description="Documents related to this project will appear here."
              />
            )}
          </Section>
          {mayCreateDocument ? (
            <>
              <DocumentCreateDialog
                open={createOpen}
                onOpenChange={setCreateOpen}
                create={async (body) => {
                  const { data } = await v1ProjectsDocumentsCreate({
                    path: { id: project.id },
                    body,
                    throwOnError: true,
                  });
                  return data;
                }}
                queryKeysToInvalidate={[
                  documentListQueryKey("project", project.id),
                ]}
              />
              <DocumentLinkDialog
                namespaceId={namespace.id}
                relatedIds={relatedIds}
                relatedLabel="this project"
                open={linkOpen}
                onOpenChange={setLinkOpen}
                onLink={async (documentId) => {
                  await v1ProjectsDocumentsRelate({
                    path: { id: project.id, documentId },
                    throwOnError: true,
                  });
                }}
              />
            </>
          ) : null}
        </div>

        <div className="space-y-8">
          <Section title="Details">
            <PropertyList
              items={[
                { label: "Key", value: project.key },
                {
                  label: "State",
                  value: <StatusIndicator status={project.status} />,
                },
                {
                  label: "Namespace",
                  value: (
                    <InternalLink
                      to={internalPath(`/namespaces/${namespace.id}`)}
                      className="text-primary hover:underline"
                    >
                      {namespace.name}
                    </InternalLink>
                  ),
                },
                {
                  label: "Issues",
                  value: `${issueCount} linked issues`,
                },
                {
                  label: "Documents",
                  value: documentCount,
                },
              ]}
            />
          </Section>
          <Section title="Relations">
            <MockDataAlert title="Illustrative relationships" className="mb-3">
              Project relationships are illustrative examples linked to the work
              shown above.
            </MockDataAlert>
            <RelationList
              relations={projectRelations}
              entity={{
                id: work[0]?.id ?? project.id,
                type: "work-item",
              }}
            />
          </Section>
        </div>
      </div>
    </ContentWidth>
  );
}

export function ProjectDocumentsPage({
  namespace,
  project,
}: {
  namespace: Namespace;
  project: Project;
}) {
  const [query, setQuery] = useState("");
  const [sort, setSort] = useState<DocumentListSort>("updated-desc");
  const [creatorId, setCreatorId] = useState(ALL_DOCUMENT_CREATORS);
  const [renamingDocument, setRenamingDocument] =
    useState<PartialDocument | null>(null);
  const [deletingDocument, setDeletingDocument] =
    useState<PartialDocument | null>(null);
  const [linkOpen, setLinkOpen] = useState(false);
  const { data: permissions } = usePermissions(
    withResourceType(ResourceType.Project, project.id)
  );
  const mayWrite =
    can(permissions, Action.DocumentCreate) ||
    can(permissions, Action.DocumentUpdate) ||
    can(permissions, Action.DocumentDelete);
  const queryClient = useQueryClient();
  const unlinkMutation = useMutation({
    mutationFn: async (document: PartialDocument) => {
      await v1ProjectsDocumentsUnrelate({
        path: { id: project.id, documentId: document.id },
        throwOnError: true,
      });
      return document;
    },
    onSuccess: async (document) => {
      await invalidateDocumentQueries(queryClient, document.id);
      showSuccessToast(
        "Document unlinked",
        `${document.title} is no longer related to this project`
      );
    },
    onError: (error) => {
      showErrorToast(
        "Failed to unlink document",
        error instanceof Error ? error.message : "Unknown error occurred"
      );
    },
  });
  const pageNav = useCursorPageNav({
    resetKey: `${query}|${sort}|${creatorId}`,
  });
  const { data: documentsPage, isLoading } = useQuery(
    v1ProjectsDocumentsGetOptions({
      path: { id: project.id },
      query: cursorPageQuery(pageNav.pageToken),
    })
  );
  const documents = useMemo(
    () => visibleDocuments(documentsPage?.items ?? [], query, sort, creatorId),
    [creatorId, documentsPage?.items, query, sort]
  );
  const relatedIds = useMemo(
    () => new Set((documentsPage?.items ?? []).map((document) => document.id)),
    [documentsPage?.items]
  );
  const creators = useMemo(
    () => documentCreators(documentsPage?.items ?? []),
    [documentsPage?.items]
  );

  return (
    <ContentWidth
      width="overview"
      className="space-y-6"
      data-section="documents"
    >
      <EntityHeader
        type="document"
        eyebrow={`${namespace.name} / ${project.name}`}
        title="Documents"
        description="Documents related to this project."
        showIcon={false}
        actions={
          mayWrite ? (
            <div className="flex items-center gap-2">
              <LinkButton size="sm" onClick={() => setLinkOpen(true)} />
              <CreateButton onClick={() => openQuickCreate("document")} />
            </div>
          ) : undefined
        }
      />
      <DocumentListToolbar
        query={query}
        onQueryChange={setQuery}
        sort={sort}
        onSortChange={setSort}
        creators={creators}
        creatorId={creatorId}
        onCreatorChange={setCreatorId}
      />
      {isLoading ? (
        <ListSkeleton />
      ) : documents.length > 0 ? (
        <>
          <DocumentList
            documents={documents}
            onRename={
              mayWrite
                ? (document) => {
                    setRenamingDocument(document);
                  }
                : undefined
            }
            onUnlink={
              mayWrite
                ? (document) => {
                    void unlinkMutation.mutateAsync(document);
                  }
                : undefined
            }
            onDelete={
              mayWrite
                ? (document) => {
                    setDeletingDocument(document);
                  }
                : undefined
            }
          />
          <CursorPaginator {...cursorPaginatorProps(documentsPage, pageNav)} />
        </>
      ) : (
        <EmptyState
          icon={<FileTextIcon />}
          title={query ? "No document matches" : "No related documents"}
          description={
            query
              ? "Try a different title or summary."
              : "Documents related to this project will appear here."
          }
          action={
            query ? (
              <Button variant="outline" onClick={() => setQuery("")}>
                Clear search
              </Button>
            ) : mayWrite ? (
              <CreateButton onClick={() => openQuickCreate("document")} />
            ) : undefined
          }
        />
      )}
      {renamingDocument ? (
        <DocumentRenameDialog
          document={renamingDocument}
          open
          onOpenChange={(open) => {
            if (!open) {
              setRenamingDocument(null);
            }
          }}
        />
      ) : null}
      {deletingDocument ? (
        <DocumentDeleteDialog
          document={deletingDocument}
          open
          onOpenChange={(open) => {
            if (!open) {
              setDeletingDocument(null);
            }
          }}
          navigateOnSuccess={false}
        />
      ) : null}
      {mayWrite ? (
        <DocumentLinkDialog
          namespaceId={namespace.id}
          relatedIds={relatedIds}
          relatedLabel="this project"
          open={linkOpen}
          onOpenChange={setLinkOpen}
          onLink={async (documentId) => {
            await v1ProjectsDocumentsRelate({
              path: { id: project.id, documentId },
              throwOnError: true,
            });
          }}
        />
      ) : null}
    </ContentWidth>
  );
}

export function ProjectActivityPage({
  namespace,
  project,
}: {
  namespace: Namespace;
  project: Project;
}) {
  const activity = selectActivity({ limit: 20 });
  return (
    <ContentWidth width="overview" className="space-y-6">
      <EntityHeader
        type="project"
        eyebrow={`${namespace.name} / ${project.name}`}
        title="Activity"
        description="A bounded history of meaningful project changes."
        showIcon={false}
      />
      <MockDataAlert title="Illustrative activity">
        Activity entries below are illustrative examples and may not reflect
        this project’s live history.
      </MockDataAlert>
      {activity.length > 0 ? (
        <ActivityFeed entries={activity} />
      ) : (
        <EmptyState
          icon={<ActivityIcon />}
          title="No activity"
          description="Meaningful changes will appear here as activity becomes available."
        />
      )}
    </ContentWidth>
  );
}
