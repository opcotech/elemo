import { EditorContent, useEditor, useEditorState } from "@tiptap/react";
import {
  BetweenHorizontalEndIcon,
  BetweenVerticalEndIcon,
  BoldIcon,
  CodeIcon,
  Heading1Icon,
  Heading2Icon,
  Heading3Icon,
  ItalicIcon,
  LinkIcon,
  ListChecksIcon,
  ListIcon,
  ListOrderedIcon,
  MinusIcon,
  QuoteIcon,
  Redo2Icon,
  SmileIcon,
  SquareCodeIcon,
  StrikethroughIcon,
  TableColumnsSplitIcon,
  TableIcon,
  TableRowsSplitIcon,
  UnderlineIcon,
  Undo2Icon,
} from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { LinkAddDialog } from "@/components/ui/link-add-dialog";
import {
  CODE_BLOCK_LANGUAGES,
  createRichTextExtensions,
  editorValueFrom,
  preventLinkNavigation,
} from "@/components/ui/rich-text-extensions";
import type {
  RichTextEditorValue,
  RichTextMentionItem,
} from "@/components/ui/rich-text-extensions";
import {
  applyLinkDraft,
  captureLinkDraft,
  removeLinkDraft,
} from "@/components/ui/rich-text-link";
import type { LinkDialogDraft } from "@/components/ui/rich-text-link";
import { cn } from "@/lib/utils";

export type { RichTextEditorValue, RichTextMentionItem };

interface RichTextEditorProps {
  content?: string;
  editable?: boolean;
  disabled?: boolean;
  placeholder?: string;
  className?: string;
  "aria-label"?: string;
  fallback?: ReactNode;
  autoFocus?: boolean;
  mentionItems?: readonly RichTextMentionItem[];
  onChange?: (value: RichTextEditorValue) => void;
  onReady?: () => void;
}

function ToolbarButton({
  label,
  active,
  disabled,
  onClick,
  children,
}: {
  label: string;
  active?: boolean;
  disabled?: boolean;
  onClick: () => void;
  children: ReactNode;
}) {
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon-xs"
      aria-label={label}
      title={label}
      aria-pressed={active}
      disabled={disabled}
      className={cn(active && "bg-primary/10 text-primary-on-subtle")}
      onClick={onClick}
    >
      {children}
    </Button>
  );
}

function ToolbarDivider() {
  return (
    <span
      aria-hidden
      className="bg-border mx-1 hidden h-4 w-px sm:inline-block"
    />
  );
}

export function RichTextEditor({
  content = "",
  editable = true,
  disabled = false,
  placeholder = "Describe the work…",
  className,
  "aria-label": ariaLabel = "Rich text editor",
  fallback,
  autoFocus = false,
  mentionItems = [],
  onChange,
  onReady,
}: RichTextEditorProps) {
  const mentionItemsRef = useRef(mentionItems);
  mentionItemsRef.current = mentionItems;
  const [linkDialogOpen, setLinkDialogOpen] = useState(false);
  const [linkDraft, setLinkDraft] = useState<LinkDialogDraft | null>(null);

  const extensions = useMemo(
    () =>
      createRichTextExtensions({
        placeholder,
        getMentionItems: () => mentionItemsRef.current,
      }),
    [placeholder]
  );

  const editor = useEditor({
    extensions,
    content,
    contentType: "markdown",
    editable: editable && !disabled,
    // Only mounted after a client interaction (click-to-edit), so sync init is safe
    // and avoids a null first paint that flashes the description box.
    immediatelyRender: true,
    autofocus: autoFocus ? "end" : false,
    editorProps: {
      attributes: {
        "aria-label": ariaLabel,
        class: "rich-text-editor__content focus:outline-none",
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
      onChange?.(editorValueFrom(current));
      onReady?.();
    },
    onUpdate: ({ editor: current }) => {
      onChange?.(editorValueFrom(current));
    },
  });

  useEffect(() => {
    if (!editor) {
      return;
    }
    editor.setEditable(editable && !disabled);
  }, [editor, editable, disabled]);

  const toolbarState = useEditorState({
    editor,
    selector: ({ editor: current }) => {
      if (!current) {
        return {
          bold: false,
          italic: false,
          underline: false,
          strike: false,
          heading1: false,
          heading2: false,
          heading3: false,
          bulletList: false,
          orderedList: false,
          taskList: false,
          blockquote: false,
          code: false,
          codeBlock: false,
          codeLanguage: "plaintext",
          link: false,
          canUndo: false,
          canRedo: false,
          canAddRow: false,
          canDeleteRow: false,
          canAddColumn: false,
          canDeleteColumn: false,
        };
      }
      return {
        bold: current.isActive("bold"),
        italic: current.isActive("italic"),
        underline: current.isActive("underline"),
        strike: current.isActive("strike"),
        heading1: current.isActive("heading", { level: 1 }),
        heading2: current.isActive("heading", { level: 2 }),
        heading3: current.isActive("heading", { level: 3 }),
        bulletList: current.isActive("bulletList"),
        orderedList: current.isActive("orderedList"),
        taskList: current.isActive("taskList"),
        blockquote: current.isActive("blockquote"),
        code: current.isActive("code"),
        codeBlock: current.isActive("codeBlock"),
        codeLanguage:
          (current.getAttributes("codeBlock").language as string | undefined) ??
          "plaintext",
        link: current.isActive("link"),
        canUndo: current.can().undo(),
        canRedo: current.can().redo(),
        canAddRow: current.can().addRowAfter(),
        canDeleteRow: current.can().deleteRow(),
        canAddColumn: current.can().addColumnAfter(),
        canDeleteColumn: current.can().deleteColumn(),
      };
    },
  });

  const controlsDisabled = disabled || !editable;

  const openLinkDialog = () => {
    if (!editor) {
      return;
    }
    setLinkDraft(captureLinkDraft(editor));
    setLinkDialogOpen(true);
  };

  const closeLinkDialog = (open: boolean) => {
    setLinkDialogOpen(open);
    if (!open) {
      setLinkDraft(null);
      editor?.chain().focus().run();
    }
  };

  return (
    <div
      className={cn(
        "rich-text-editor border-border bg-card rounded-xl border",
        disabled && "opacity-60",
        className
      )}
    >
      {editable && (
        <div className="border-border bg-card/95 sticky top-0 z-10 flex min-h-10 flex-wrap items-center gap-0.5 rounded-t-xl border-b px-2 py-1.5 backdrop-blur">
          {editor ? (
            <>
              <ToolbarButton
                label="Undo"
                disabled={controlsDisabled || !toolbarState?.canUndo}
                onClick={() => editor.chain().focus().undo().run()}
              >
                <Undo2Icon />
              </ToolbarButton>
              <ToolbarButton
                label="Redo"
                disabled={controlsDisabled || !toolbarState?.canRedo}
                onClick={() => editor.chain().focus().redo().run()}
              >
                <Redo2Icon />
              </ToolbarButton>
              <ToolbarDivider />
              <ToolbarButton
                label="Heading 1"
                active={toolbarState?.heading1}
                disabled={controlsDisabled}
                onClick={() =>
                  editor.chain().focus().toggleHeading({ level: 1 }).run()
                }
              >
                <Heading1Icon />
              </ToolbarButton>
              <ToolbarButton
                label="Heading 2"
                active={toolbarState?.heading2}
                disabled={controlsDisabled}
                onClick={() =>
                  editor.chain().focus().toggleHeading({ level: 2 }).run()
                }
              >
                <Heading2Icon />
              </ToolbarButton>
              <ToolbarButton
                label="Heading 3"
                active={toolbarState?.heading3}
                disabled={controlsDisabled}
                onClick={() =>
                  editor.chain().focus().toggleHeading({ level: 3 }).run()
                }
              >
                <Heading3Icon />
              </ToolbarButton>
              <ToolbarDivider />
              <ToolbarButton
                label="Bold"
                active={toolbarState?.bold}
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().toggleBold().run()}
              >
                <BoldIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Italic"
                active={toolbarState?.italic}
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().toggleItalic().run()}
              >
                <ItalicIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Underline"
                active={toolbarState?.underline}
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().toggleUnderline().run()}
              >
                <UnderlineIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Strikethrough"
                active={toolbarState?.strike}
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().toggleStrike().run()}
              >
                <StrikethroughIcon />
              </ToolbarButton>
              <ToolbarDivider />
              <ToolbarButton
                label="Bullet list"
                active={toolbarState?.bulletList}
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().toggleBulletList().run()}
              >
                <ListIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Ordered list"
                active={toolbarState?.orderedList}
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().toggleOrderedList().run()}
              >
                <ListOrderedIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Task list"
                active={toolbarState?.taskList}
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().toggleTaskList().run()}
              >
                <ListChecksIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Quote"
                active={toolbarState?.blockquote}
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().toggleBlockquote().run()}
              >
                <QuoteIcon />
              </ToolbarButton>
              <ToolbarDivider />
              <ToolbarButton
                label="Inline code"
                active={toolbarState?.code}
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().toggleCode().run()}
              >
                <CodeIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Code block"
                active={toolbarState?.codeBlock}
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().toggleCodeBlock().run()}
              >
                <SquareCodeIcon />
              </ToolbarButton>
              {toolbarState?.codeBlock ? (
                <label className="text-muted-foreground ml-1 flex items-center gap-1 text-xs">
                  <span className="sr-only">Code language</span>
                  <select
                    className="border-border bg-background h-7 rounded-md border px-1.5 text-xs"
                    value={
                      CODE_BLOCK_LANGUAGES.includes(
                        toolbarState.codeLanguage as (typeof CODE_BLOCK_LANGUAGES)[number]
                      )
                        ? toolbarState.codeLanguage
                        : "plaintext"
                    }
                    disabled={controlsDisabled}
                    onChange={(event) => {
                      editor
                        .chain()
                        .focus()
                        .updateAttributes("codeBlock", {
                          language: event.target.value,
                        })
                        .run();
                    }}
                  >
                    {CODE_BLOCK_LANGUAGES.map((language) => (
                      <option key={language} value={language}>
                        {language}
                      </option>
                    ))}
                  </select>
                </label>
              ) : null}
              <ToolbarButton
                label="Link"
                active={toolbarState?.link}
                disabled={controlsDisabled}
                onClick={openLinkDialog}
              >
                <LinkIcon />
              </ToolbarButton>
              <ToolbarDivider />
              <ToolbarButton
                label="Horizontal rule"
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().setHorizontalRule().run()}
              >
                <MinusIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Insert table"
                disabled={controlsDisabled}
                onClick={() =>
                  editor
                    .chain()
                    .focus()
                    .insertTable({ rows: 3, cols: 3, withHeaderRow: true })
                    .run()
                }
              >
                <TableIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Add row"
                disabled={controlsDisabled || !toolbarState?.canAddRow}
                onClick={() => editor.chain().focus().addRowAfter().run()}
              >
                <BetweenHorizontalEndIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Delete row"
                disabled={controlsDisabled || !toolbarState?.canDeleteRow}
                onClick={() => editor.chain().focus().deleteRow().run()}
              >
                <TableRowsSplitIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Add column"
                disabled={controlsDisabled || !toolbarState?.canAddColumn}
                onClick={() => editor.chain().focus().addColumnAfter().run()}
              >
                <BetweenVerticalEndIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Delete column"
                disabled={controlsDisabled || !toolbarState?.canDeleteColumn}
                onClick={() => editor.chain().focus().deleteColumn().run()}
              >
                <TableColumnsSplitIcon />
              </ToolbarButton>
              <ToolbarButton
                label="Insert emoji"
                disabled={controlsDisabled}
                onClick={() => editor.chain().focus().insertContent(":").run()}
              >
                <SmileIcon />
              </ToolbarButton>
            </>
          ) : null}
        </div>
      )}
      {editor ? (
        <EditorContent editor={editor} className="rich-text-editor__surface" />
      ) : (
        <div className="rich-text-editor__surface rich-text-editor__fallback">
          {fallback}
        </div>
      )}
      <LinkAddDialog
        open={linkDialogOpen}
        onOpenChange={closeLinkDialog}
        url={linkDraft?.href ?? ""}
        onUrlChange={(href) =>
          setLinkDraft((current) => (current ? { ...current, href } : current))
        }
        label={linkDraft?.label ?? ""}
        onLabelChange={(label) =>
          setLinkDraft((current) => (current ? { ...current, label } : current))
        }
        showLabel
        title={linkDraft?.editing ? "Edit link" : "Add link"}
        description="Set the URL and the visible label for this link."
        submitLabel={linkDraft?.editing ? "Save" : "Add link"}
        onSubmit={() => {
          if (!editor || !linkDraft) {
            return;
          }
          applyLinkDraft(editor, linkDraft);
          closeLinkDialog(false);
        }}
        onRemove={
          linkDraft?.editing
            ? () => {
                if (!editor || !linkDraft) {
                  return;
                }
                removeLinkDraft(editor, linkDraft);
                closeLinkDialog(false);
              }
            : undefined
        }
      />
    </div>
  );
}
