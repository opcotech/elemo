import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { FileTextIcon } from "lucide-react";
import { useState } from "react";

import { DocumentCreateDialog } from "@/components/documents/document-create-dialog";
import { DocumentLinkDialog } from "@/components/documents/document-link-dialog";
import { DocumentSummaryList } from "@/components/documents/document-summary-list";
import { AddButton, LinkButton } from "@/components/ui/create-button";
import { EmptyState } from "@/components/ui/empty-state";
import { ListSkeleton } from "@/components/ui/list-skeleton";
import { Section } from "@/components/ui/section";
import {
  v1IssueGetOptions,
  v1IssuesDocumentsGetOptions,
  v1NamespacesIssuesKeyGetOptions,
} from "@/lib/api/query-options";
import { namespaceRefPath } from "@/lib/api/refs";
import {
  v1IssuesDocumentsCreate,
  v1IssuesDocumentsRelate,
  v1IssuesDocumentsUnrelate,
} from "@/lib/api/sdk";
import type { PartialDocument } from "@/lib/api/types";
import { documentListQueryKey } from "@/lib/documents/create";
import { invalidateDocumentQueries } from "@/lib/documents/document-queries";
import { showErrorToast, showSuccessToast } from "@/lib/toast";

export function IssueDocumentsSection({
  issueId,
  organizationId,
  namespaceId,
  issueKey,
  documentCount,
  canCreate,
}: {
  issueId: string;
  organizationId: string;
  namespaceId: string;
  issueKey: string;
  documentCount?: number | null;
  canCreate: boolean;
}) {
  const [createOpen, setCreateOpen] = useState(false);
  const [linkOpen, setLinkOpen] = useState(false);
  const { data: documentsPage, isLoading } = useQuery(
    v1IssuesDocumentsGetOptions({
      path: { id: issueId },
      query: { page_size: 20 },
    })
  );
  const documents = documentsPage?.items ?? [];
  const relatedIds = new Set(documents.map((document) => document.id));
  const queryClient = useQueryClient();
  const unlinkMutation = useMutation({
    mutationFn: async (document: PartialDocument) => {
      await v1IssuesDocumentsUnrelate({
        path: { id: issueId, documentId: document.id },
        throwOnError: true,
      });
      return document;
    },
    onSuccess: async (document) => {
      await invalidateDocumentQueries(queryClient, document.id);
      await queryClient.invalidateQueries({
        queryKey: v1IssueGetOptions({ path: { id: issueId } }).queryKey,
      });
      await queryClient.invalidateQueries({
        queryKey: v1NamespacesIssuesKeyGetOptions({
          path: {
            ...namespaceRefPath(organizationId, namespaceId),
            key: issueKey,
          },
        }).queryKey,
      });
      showSuccessToast(
        "Document unlinked",
        `${document.title} is no longer related to this work`
      );
    },
    onError: (error) => {
      showErrorToast(
        "Failed to unlink document",
        error instanceof Error ? error.message : "Unknown error occurred"
      );
    },
  });
  const title =
    documentCount != null
      ? `Linked documents (${documentCount})`
      : "Linked documents";

  const actions = canCreate ? (
    <div className="flex items-center gap-2">
      <LinkButton onClick={() => setLinkOpen(true)} />
      <AddButton onClick={() => setCreateOpen(true)} />
    </div>
  ) : undefined;

  return (
    <div data-section="issue-documents">
      <Section title={title} action={actions}>
        {isLoading ? (
          <ListSkeleton rows={4} />
        ) : documents.length > 0 ? (
          <DocumentSummaryList
            documents={documents}
            onUnlink={
              canCreate
                ? (document) => {
                    void unlinkMutation.mutateAsync(document);
                  }
                : undefined
            }
          />
        ) : (
          <EmptyState
            compact
            icon={<FileTextIcon />}
            title="No linked documents"
            description="Specifications and decisions related to this work will appear here."
          />
        )}
      </Section>
      {canCreate ? (
        <>
          <DocumentCreateDialog
            open={createOpen}
            onOpenChange={setCreateOpen}
            create={async (body) => {
              const { data } = await v1IssuesDocumentsCreate({
                path: { id: issueId },
                body,
                throwOnError: true,
              });
              return data;
            }}
            queryKeysToInvalidate={[
              documentListQueryKey("issue", issueId),
              v1IssueGetOptions({ path: { id: issueId } }).queryKey,
              v1NamespacesIssuesKeyGetOptions({
                path: {
                  ...namespaceRefPath(organizationId, namespaceId),
                  key: issueKey,
                },
              }).queryKey,
            ]}
          />
          <DocumentLinkDialog
            organizationId={organizationId}
            namespaceId={namespaceId}
            relatedIds={relatedIds}
            relatedLabel="this work"
            open={linkOpen}
            onOpenChange={setLinkOpen}
            onLink={async (documentId) => {
              await v1IssuesDocumentsRelate({
                path: { id: issueId, documentId },
                throwOnError: true,
              });
            }}
          />
        </>
      ) : null}
    </div>
  );
}
