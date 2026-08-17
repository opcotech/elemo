import { useSuspenseQuery } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { DocumentChangeLibraryDialog } from "./document-change-library-dialog";
import { DocumentDeleteDialog } from "./document-delete-dialog";
import { DocumentEditor } from "./document-editor";
import { DocumentInlineExcerpt } from "./document-inline-excerpt";
import { DocumentInlineTitle } from "./document-inline-title";
import { DocumentLocation } from "./document-location";
import { DocumentMoveDialog } from "./document-move-dialog";
import { useDocumentUpdate } from "./use-document-update";

import { PageActions } from "@/components/shared/entity-header";
import { Button } from "@/components/ui/button";
import type { RichTextMentionItem } from "@/components/ui/rich-text-extensions";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { v1DocumentGetOptions } from "@/lib/api/query-options";
import type { Document, DocumentPatch } from "@/lib/api/types";
import {
  parseDocumentContent,
  parseDocumentTitle,
  resolveDocumentExcerpt,
} from "@/lib/documents/document-edit";
import { documentLibraryKindFromType } from "@/lib/documents/library";
import { formatDate } from "@/lib/format-date";
import { internalPath } from "@/lib/internal-url";
import { showErrorToast, showSuccessToast } from "@/lib/toast";
import { uiActions } from "@/lib/ui-store";
import { getDefaultValue } from "@/lib/utils";
import { useAccessibleOrganizationMembers } from "@/lib/work/use-organization-members-for-namespace";

function documentUrl(documentId: string): string {
  const path = `/documents/${documentId}`;
  if (typeof window === "undefined") {
    return path;
  }
  return new URL(path, window.location.origin).href;
}

function timestampLabel(document: Document): string {
  const created = formatDate(document.created_at);
  const updated = formatDate(document.updated_at ?? document.created_at);
  return `Created ${created} · Updated ${updated}`;
}

function normalizedExcerpt(value: string | null | undefined): string | null {
  return value?.trim() || null;
}

function saveMessage(patch: DocumentPatch): string {
  const changed = [
    patch.title !== undefined,
    patch.excerpt !== undefined,
    patch.content !== undefined,
  ].filter(Boolean).length;
  if (changed !== 1) {
    return "Your changes were saved";
  }
  if (patch.title !== undefined) {
    return "Title updated";
  }
  if (patch.excerpt !== undefined) {
    return "Excerpt updated";
  }
  return "Content updated";
}

async function writeClipboardText(value: string): Promise<void> {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value);
      return;
    } catch {
      // Fall through to execCommand.
    }
  }

  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  document.body.appendChild(textarea);
  textarea.select();
  const copied = document.execCommand("copy");
  textarea.remove();
  if (!copied) {
    throw new Error("Clipboard unavailable");
  }
}

export function DocumentPageSkeleton() {
  return (
    <div
      className="document-editor flex min-h-full flex-col"
      role="status"
      aria-busy="true"
    >
      <span className="sr-only">Loading page</span>
      <div className="bg-background border-b">
        <div className="flex min-h-10 items-center gap-2 px-2 py-1.5">
          <Skeleton className="h-6 flex-1" />
          <Skeleton className="h-6 w-24" />
          <Skeleton className="size-7 rounded-md" />
        </div>
      </div>
      <div className="flex flex-1 justify-center gap-6 px-4 py-8 sm:px-8 sm:py-10 lg:px-12 lg:py-12">
        <div className="document-editor__paper bg-card w-full max-w-5xl min-w-0 space-y-8">
          <Skeleton className="h-12 w-2/3" />
          <Skeleton className="h-6 w-1/2" />
          <Skeleton className="h-80 w-full" />
        </div>
      </div>
    </div>
  );
}

export function DocumentPage({
  initialDocument,
}: {
  initialDocument: Document;
}) {
  const { data: document } = useSuspenseQuery({
    ...v1DocumentGetOptions({
      path: { id: initialDocument.id },
    }),
    initialData: initialDocument,
  });

  const { updateDocument, isPending } = useDocumentUpdate(document.id);
  const recentHref = internalPath(`/documents/${document.id}`);
  const { members } = useAccessibleOrganizationMembers();
  const [titleDraft, setTitleDraft] = useState(document.title);
  const [excerptDraft, setExcerptDraft] = useState(document.excerpt ?? "");
  const [contentDraft, setContentDraft] = useState(() =>
    getDefaultValue(document.content)
  );
  const [contentDirty, setContentDirty] = useState(false);
  const contentBaselineRef = useRef<string | null>(null);
  const [editorKey, setEditorKey] = useState(0);
  const [titleError, setTitleError] = useState<string | null>(null);
  const [excerptError, setExcerptError] = useState<string | null>(null);
  const [contentError, setContentError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [moveDialogOpen, setMoveDialogOpen] = useState(false);
  const [libraryDialogOpen, setLibraryDialogOpen] = useState(false);

  const titleDirty = titleDraft.trim() !== document.title;
  const excerptDirty =
    normalizedExcerpt(excerptDraft) !== normalizedExcerpt(document.excerpt);
  const dirty = titleDirty || excerptDirty || contentDirty;

  const mentionItems = useMemo<RichTextMentionItem[]>(
    () =>
      members.map((member) => {
        const label =
          `${member.first_name} ${member.last_name}`.trim() || member.email;
        return {
          id: member.id,
          label,
          detail: member.email,
        };
      }),
    [members]
  );

  useEffect(() => {
    uiActions.rememberRecentEntity({
      id: document.id,
      type: "document",
      label: document.title,
      href: recentHref,
    });
  }, [document.id, document.title, recentHref]);

  useEffect(() => {
    if (!titleDirty) {
      setTitleDraft(document.title);
    }
  }, [document.title, titleDirty]);

  useEffect(() => {
    if (!excerptDirty) {
      setExcerptDraft(document.excerpt ?? "");
    }
  }, [document.excerpt, excerptDirty]);

  useEffect(() => {
    if (!contentDirty) {
      setContentDraft(getDefaultValue(document.content));
    }
  }, [document.content, contentDirty]);

  const saveDocument = useCallback(async () => {
    if (saving || isPending || !dirty) {
      return;
    }

    const parsedTitle = parseDocumentTitle(titleDraft);
    const parsedContent = parseDocumentContent(contentDraft);
    const excerptSource = parsedContent.ok
      ? parsedContent.content
      : contentDraft;
    const parsedExcerpt = resolveDocumentExcerpt(excerptDraft, excerptSource);

    setTitleError(parsedTitle.ok ? null : parsedTitle.error);
    setExcerptError(parsedExcerpt.ok ? null : parsedExcerpt.error);
    setContentError(
      contentDirty && !parsedContent.ok ? parsedContent.error : null
    );

    if (
      !parsedTitle.ok ||
      !parsedExcerpt.ok ||
      (contentDirty && !parsedContent.ok)
    ) {
      return;
    }

    const patch: DocumentPatch = {};
    if (parsedTitle.title !== document.title) {
      patch.title = parsedTitle.title;
    }
    if (parsedExcerpt.excerpt !== normalizedExcerpt(document.excerpt)) {
      patch.excerpt = parsedExcerpt.excerpt;
    }
    if (contentDirty && parsedContent.ok) {
      patch.content = parsedContent.content;
    }

    if (
      patch.title === undefined &&
      patch.excerpt === undefined &&
      patch.content === undefined
    ) {
      setTitleDraft(parsedTitle.title);
      setExcerptDraft(parsedExcerpt.excerpt ?? "");
      setContentDirty(false);
      return;
    }

    setSaving(true);
    try {
      const saved = await updateDocument(patch, saveMessage(patch));
      setTitleDraft(saved.title);
      setExcerptDraft(saved.excerpt ?? "");
      contentBaselineRef.current = contentDraft;
      setContentDirty(false);
      setTitleError(null);
      setExcerptError(null);
      setContentError(null);
    } catch {
      setContentError("Could not save document");
    } finally {
      setSaving(false);
    }
  }, [
    contentDirty,
    contentDraft,
    dirty,
    document.excerpt,
    document.title,
    excerptDraft,
    isPending,
    titleDraft,
    saving,
    updateDocument,
  ]);

  const discardChanges = () => {
    setTitleDraft(document.title);
    setExcerptDraft(document.excerpt ?? "");
    setContentDraft(getDefaultValue(document.content));
    contentBaselineRef.current = null;
    setContentDirty(false);
    setTitleError(null);
    setExcerptError(null);
    setContentError(null);
    setEditorKey((key) => key + 1);
  };

  const copyUrl = documentUrl(document.id);
  const relocateDisabled = saving || isPending || dirty;
  const libraryKind = documentLibraryKindFromType(document.library.type);

  return (
    <>
      <DocumentEditor
        key={`${document.id}-${editorKey}`}
        content={getDefaultValue(document.content)}
        disabled={saving || isPending}
        mentionItems={mentionItems}
        trailing={
          <div className="border-border flex shrink-0 items-center gap-1.5 border-l pl-2">
            <TooltipProvider delay={300}>
              <Tooltip>
                <TooltipTrigger
                  render={
                    <p className="text-muted-foreground max-w-24 text-xs leading-none sm:max-w-36" />
                  }
                >
                  {formatDate(document.updated_at ?? document.created_at)}
                </TooltipTrigger>
                <TooltipContent>{timestampLabel(document)}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
            {dirty ? (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                disabled={saving}
                onClick={discardChanges}
              >
                Discard
              </Button>
            ) : null}
            {dirty ? (
              <Button
                type="button"
                size="sm"
                disabled={saving || isPending}
                onClick={() => {
                  void saveDocument();
                }}
              >
                {saving ? "Saving…" : "Save"}
              </Button>
            ) : null}
            <PageActions
              size="icon-sm"
              secondary={[
                {
                  label: "View relationships",
                  href: `/relations/document/${document.id}`,
                },
                {
                  label: "Copy link",
                  onSelect: () => {
                    void writeClipboardText(copyUrl).then(
                      () => {
                        showSuccessToast("Copied", copyUrl);
                      },
                      (error: unknown) => {
                        showErrorToast(
                          "Could not copy",
                          error instanceof Error
                            ? error
                            : "Clipboard unavailable"
                        );
                      }
                    );
                  },
                },
                {
                  label: "Move",
                  disabled: relocateDisabled,
                  onSelect: () => setMoveDialogOpen(true),
                },
                {
                  label: "Change library",
                  disabled: relocateDisabled,
                  onSelect: () => setLibraryDialogOpen(true),
                },
                {
                  label: "Delete",
                  variant: "destructive",
                  onSelect: () => setDeleteDialogOpen(true),
                },
              ]}
            />
          </div>
        }
        banner={
          contentError ? (
            <p className="text-destructive px-3 pb-1.5 text-xs">
              {contentError}
            </p>
          ) : null
        }
        leading={
          <>
            <DocumentInlineTitle
              value={titleDraft}
              disabled={saving || isPending}
              error={titleError}
              onChange={(value) => {
                setTitleDraft(value);
                if (titleError) {
                  setTitleError(null);
                }
              }}
              onReset={() => {
                setTitleDraft(document.title);
                setTitleError(null);
              }}
            />
            <DocumentInlineExcerpt
              value={excerptDraft}
              disabled={saving || isPending}
              error={excerptError}
              onChange={(value) => {
                setExcerptDraft(value);
                if (excerptError) {
                  setExcerptError(null);
                }
              }}
              onReset={() => {
                setExcerptDraft(document.excerpt ?? "");
                setExcerptError(null);
              }}
            />
          </>
        }
        footer={<DocumentLocation document={document} />}
        onDraftChange={(markdown, source) => {
          setContentDraft(markdown);
          if (source === "create") {
            contentBaselineRef.current = markdown;
            setContentDirty(false);
            return;
          }
          setContentDirty(markdown !== contentBaselineRef.current);
          if (contentError) {
            setContentError(null);
          }
        }}
        onSave={() => {
          void saveDocument();
        }}
      />
      <DocumentDeleteDialog
        document={document}
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
      />
      <DocumentMoveDialog
        documentId={document.id}
        documentTitle={document.title}
        kind={libraryKind}
        libraryId={document.library.id}
        currentFolderId={document.folder?.id}
        open={moveDialogOpen}
        onOpenChange={setMoveDialogOpen}
      />
      <DocumentChangeLibraryDialog
        documentId={document.id}
        currentLibrary={document.library}
        open={libraryDialogOpen}
        onOpenChange={setLibraryDialogOpen}
      />
    </>
  );
}
