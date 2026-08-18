import Highlight from "@tiptap/extension-highlight";
import Typography from "@tiptap/extension-typography";
import { EditorContent, useEditor } from "@tiptap/react";
import type { Editor } from "@tiptap/react";
import { ListTreeIcon } from "lucide-react";
import { useEffect, useMemo, useRef } from "react";
import type { ReactNode } from "react";

import { DocumentEditorToc, useDocumentTocOpen } from "./document-editor-toc";
import { DocumentEditorToolbar } from "./document-editor-toolbar";

import { Button } from "@/components/ui/button";
import {
  createRichTextExtensions,
  preventLinkNavigation,
} from "@/components/ui/rich-text-extensions";
import type { RichTextMentionItem } from "@/components/ui/rich-text-extensions";
import { cn, getDefaultValue } from "@/lib/utils";
import { markdownToSafeHtml } from "@/lib/work/markdown-html";

function focusEditorNearPointer(
  editor: Editor,
  event: { clientX: number; clientY: number }
) {
  const rect = editor.view.dom.getBoundingClientRect();
  if (rect.width === 0 || rect.height === 0) {
    editor.chain().focus(undefined, { scrollIntoView: false }).run();
    return;
  }

  const left = Math.min(Math.max(event.clientX, rect.left + 1), rect.right - 1);
  const top = Math.min(Math.max(event.clientY, rect.top + 1), rect.bottom - 1);
  const hit = editor.view.posAtCoords({ left, top });
  if (hit) {
    editor.chain().focus(hit.pos, { scrollIntoView: false }).run();
    return;
  }

  editor.chain().focus(undefined, { scrollIntoView: false }).run();
}

export function DocumentEditor({
  content,
  disabled = false,
  mentionItems = [],
  className,
  trailing,
  banner,
  leading,
  footer,
  onDraftChange,
  onSave,
}: {
  content: string;
  disabled?: boolean;
  mentionItems?: readonly RichTextMentionItem[];
  className?: string;
  trailing?: ReactNode;
  banner?: ReactNode;
  leading?: ReactNode;
  footer?: ReactNode;
  onDraftChange: (markdown: string, source: "create" | "update") => void;
  onSave?: () => void;
}) {
  const [tocOpen, setTocOpen] = useDocumentTocOpen();
  const mentionItemsRef = useRef(mentionItems);
  mentionItemsRef.current = mentionItems;
  const onDraftChangeRef = useRef(onDraftChange);
  onDraftChangeRef.current = onDraftChange;
  const onSaveRef = useRef(onSave);
  onSaveRef.current = onSave;

  const seed = getDefaultValue(content);
  const fallbackHtml = seed.trim() ? markdownToSafeHtml(seed) : "";

  const extensions = useMemo(
    () => [
      ...createRichTextExtensions({
        placeholder: "Start writing…",
        getMentionItems: () => mentionItemsRef.current,
      }),
      Highlight.configure({
        HTMLAttributes: {
          class: "document-editor__highlight",
        },
      }),
      Typography,
    ],
    []
  );

  const editor = useEditor({
    extensions,
    content: seed,
    contentType: "markdown",
    editable: !disabled,
    immediatelyRender: false,
    editorProps: {
      attributes: {
        "aria-label": "Document content",
        class:
          "document-editor__content rich-text-editor__content focus:outline-none",
      },
      handleDOMEvents: {
        click: preventLinkNavigation,
        auxclick: (view, event) => {
          if (event.button !== 1) {
            return false;
          }
          return preventLinkNavigation(view, event);
        },
      },
    },
    onCreate: ({ editor: current }) => {
      onDraftChangeRef.current(current.getMarkdown(), "create");
    },
    onUpdate: ({ editor: current }) => {
      onDraftChangeRef.current(current.getMarkdown(), "update");
    },
  });

  useEffect(() => {
    if (!editor) {
      return;
    }
    editor.setEditable(!disabled);
  }, [editor, disabled]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "s") {
        event.preventDefault();
        onSaveRef.current?.();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  return (
    <div
      className={cn("document-editor flex min-h-full flex-col", className)}
      data-section="document-editor"
    >
      <div className="bg-background/95 sticky top-0 z-20 border-b backdrop-blur-sm">
        <div className="flex min-h-10 items-center gap-1 px-1.5 py-2.5 sm:gap-2 sm:px-2">
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className={cn(
              "hidden shrink-0 lg:inline-flex",
              tocOpen && "bg-primary/10 text-primary-on-subtle"
            )}
            aria-pressed={tocOpen}
            aria-label={
              tocOpen ? "Hide table of contents" : "Show table of contents"
            }
            title={
              tocOpen ? "Hide table of contents" : "Show table of contents"
            }
            onClick={() => {
              setTocOpen((open) => !open);
            }}
          >
            <ListTreeIcon />
          </Button>
          <DocumentEditorToolbar
            editor={editor}
            disabled={disabled}
            className="min-w-0 flex-1"
          />
          {trailing}
        </div>
        {banner}
      </div>
      <div className="flex flex-1 justify-center gap-6 px-4 py-8 sm:px-8 sm:py-10 lg:px-12 lg:py-12">
        {tocOpen ? (
          <DocumentEditorToc
            editor={editor}
            onClose={() => {
              setTocOpen(false);
            }}
          />
        ) : null}
        <div className="flex w-full max-w-5xl min-w-0 flex-col">
          <div
            className={cn(
              "document-editor__paper bg-card w-full min-w-0 cursor-text",
              disabled && "pointer-events-none opacity-60"
            )}
            onMouseDown={(event) => {
              if (event.button !== 0 || !editor || disabled) {
                return;
              }
              const target = event.target;
              if (!(target instanceof Element)) {
                return;
              }
              if (
                target.closest(
                  "input, textarea, button, a, [role='dialog'], .document-editor__leading, .ProseMirror"
                )
              ) {
                return;
              }
              event.preventDefault();
              focusEditorNearPointer(editor, event);
            }}
          >
            {leading ? (
              <div className="document-editor__leading">{leading}</div>
            ) : null}
            {editor ? (
              <EditorContent
                editor={editor}
                className="document-editor__surface"
              />
            ) : fallbackHtml ? (
              <div
                className="document-editor__surface document-editor__fallback document-editor__content rich-text-editor__content"
                dangerouslySetInnerHTML={{ __html: fallbackHtml }}
              />
            ) : (
              <div className="document-editor__surface document-editor__fallback">
                <p className="text-muted-foreground">Start writing…</p>
              </div>
            )}
          </div>
          {footer ? (
            <div className="document-editor__footer">{footer}</div>
          ) : null}
        </div>
      </div>
    </div>
  );
}
