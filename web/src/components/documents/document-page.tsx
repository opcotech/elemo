import { CheckIcon, FileQuestionIcon } from "lucide-react";
import { useEffect } from "react";

import { ContentWidth } from "@/components/layout/content-width";
import { AppEmptyState, MockDataAlert } from "@/components/shared/app-feedback";
import { EntityHeader, PageActions } from "@/components/shared/entity-header";
import { RelationList } from "@/components/shared/relation-list";
import { Section } from "@/components/shared/section";
import { Separator } from "@/components/ui/separator";
import { internalPath } from "@/lib/internal-url";
import { getDocumentBody, selectRelations } from "@/lib/mock-data";
import type { DocumentBlock } from "@/lib/mock-data/types";
import { uiActions } from "@/lib/ui-store";
import { cn } from "@/lib/utils";

function renderDocumentBlock(block: DocumentBlock) {
  if (block.type === "heading") {
    const className =
      block.level === 1
        ? "text-2xl font-semibold tracking-tight"
        : block.level === 2
          ? "text-xl font-semibold tracking-tight"
          : "text-base font-semibold";
    return (
      <div key={block.id} className={className}>
        {block.text}
      </div>
    );
  }
  if (block.type === "paragraph") {
    return (
      <p key={block.id} className="text-foreground/90 leading-7">
        {block.text}
      </p>
    );
  }
  if (block.type === "checklist") {
    return (
      <ul key={block.id} className="space-y-2">
        {block.items.map((item) => (
          <li key={item.text} className="flex items-start gap-2">
            <span
              className={cn(
                "mt-0.5 flex size-5 items-center justify-center rounded border",
                item.checked && "bg-primary text-primary-foreground"
              )}
            >
              {item.checked && <CheckIcon className="size-3.5" />}
            </span>
            <span
              className={item.checked ? "text-muted-foreground" : undefined}
            >
              {item.text}
            </span>
          </li>
        ))}
      </ul>
    );
  }
  if (block.type === "callout") {
    return (
      <div
        key={block.id}
        className={cn(
          "rounded-lg border-l-4 px-4 py-3 text-sm",
          block.tone === "warning" && "border-warning bg-warning/10",
          block.tone === "success" && "border-success bg-success/10",
          block.tone === "info" && "border-info bg-info/10"
        )}
      >
        {block.text}
      </div>
    );
  }
  return (
    <pre
      key={block.id}
      className="bg-surface-sunken overflow-x-auto rounded-lg border p-4 text-xs"
    >
      <code>{block.code}</code>
    </pre>
  );
}

export function DocumentPage({ documentId }: { documentId: string }) {
  const document = getDocumentBody(documentId);

  useEffect(() => {
    if (!document) return;
    uiActions.rememberRecentEntity({
      id: document.documentId,
      type: "document",
      label: document.title,
      href: internalPath(`/documents/${document.documentId}`),
      namespaceId: document.namespaceId,
    });
  }, [document]);

  if (!document) {
    return (
      <ContentWidth width="document" className="space-y-6">
        <EntityHeader
          type="document"
          title="Document body unavailable"
          description={`Document summary ID: ${documentId}`}
          showIcon={false}
        />
        <MockDataAlert title="Document body unavailable">
          A short summary exists for this document, but the full body is not
          available to display yet.
        </MockDataAlert>
        <AppEmptyState
          icon={<FileQuestionIcon />}
          title="No readable body"
          description="Return to the source namespace or project document list."
        />
      </ContentWidth>
    );
  }

  const relations = selectRelations({
    entity: { id: document.documentId, type: "document" },
  });
  return (
    <ContentWidth width="document" className="space-y-7">
      <EntityHeader
        type="document"
        title={document.title}
        description={`Updated ${new Intl.DateTimeFormat(undefined, {
          month: "short",
          day: "numeric",
          year: "numeric",
        }).format(new Date(document.updatedAt))}`}
        showIcon={false}
        actions={
          <PageActions
            secondary={[
              {
                label: "View relationships",
                href: `/relations/document/${document.documentId}`,
              },
              { label: "Copy link" },
            ]}
          />
        }
      />
      <MockDataAlert title="Illustrative document body">
        The body and relationships shown here are illustrative examples. Live
        documents currently provide summaries only.
      </MockDataAlert>
      <article className="space-y-6 text-base">
        {document.blocks.map(renderDocumentBlock)}
      </article>
      <Separator />
      <Section title="Relations">
        <RelationList
          relations={relations}
          entity={{ id: document.documentId, type: "document" }}
        />
      </Section>
    </ContentWidth>
  );
}
