import CodeBlockLowlight from "@tiptap/extension-code-block-lowlight";
import Emoji, { emojis } from "@tiptap/extension-emoji";
import Link from "@tiptap/extension-link";
import Mention from "@tiptap/extension-mention";
import Placeholder from "@tiptap/extension-placeholder";
import { Table } from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import TableHeader from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Underline from "@tiptap/extension-underline";
import { Markdown } from "@tiptap/markdown";
import { EditorContent, useEditor, useEditorState } from "@tiptap/react";
import type { Editor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import { common, createLowlight } from "lowlight";
import {
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
  TableIcon,
  UnderlineIcon,
  Undo2Icon,
} from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { LinkAddDialog } from "@/components/ui/link-add-dialog";
import { createSuggestionListRenderer } from "@/components/ui/rich-text-suggestion";
import type { SuggestionListItem } from "@/components/ui/rich-text-suggestion";
import { cn } from "@/lib/utils";

const lowlight = createLowlight(common);

export const CODE_BLOCK_LANGUAGES = [
  "plaintext",
  "bash",
  "c",
  "cpp",
  "css",
  "go",
  "java",
  "javascript",
  "json",
  "kotlin",
  "markdown",
  "php",
  "python",
  "ruby",
  "rust",
  "shell",
  "sql",
  "typescript",
  "xml",
  "yaml",
] as const;

export interface RichTextMentionItem {
  id: string;
  label: string;
  detail?: string;
}

export interface RichTextEditorValue {
  markdown: string;
  plainText: string;
}

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

interface LinkDialogDraft {
  href: string;
  label: string;
  from: number;
  to: number;
  empty: boolean;
  selectedText: string;
  editing: boolean;
}

function captureLinkDraft(editor: Editor): LinkDialogDraft {
  const editing = editor.isActive("link");
  if (editing) {
    editor.chain().focus().extendMarkRange("link").run();
  }

  const { from, to, empty } = editor.state.selection;
  const selectedText = empty ? "" : editor.state.doc.textBetween(from, to, " ");
  const previousHref = editor.getAttributes("link").href as string | undefined;

  return {
    href: previousHref ?? "https://",
    label: selectedText || previousHref || "",
    from,
    to,
    empty,
    selectedText,
    editing,
  };
}

function applyLinkDraft(editor: Editor, draft: LinkDialogDraft) {
  const href = draft.href.trim();
  const label = draft.label.trim();
  if (!href || !label) {
    return;
  }

  const chain = editor
    .chain()
    .focus()
    .setTextSelection({ from: draft.from, to: draft.to });

  if (draft.empty || label !== draft.selectedText) {
    chain
      .insertContent({
        type: "text",
        text: label,
        marks: [{ type: "link", attrs: { href } }],
      })
      .run();
    return;
  }

  chain.setLink({ href }).run();
}

function removeLinkDraft(editor: Editor, draft: LinkDialogDraft) {
  editor
    .chain()
    .focus()
    .setTextSelection({ from: draft.from, to: draft.to })
    .extendMarkRange("link")
    .unsetLink()
    .run();
}

function filterMentionItems(
  items: readonly RichTextMentionItem[],
  query: string
): SuggestionListItem[] {
  const normalized = query.trim().toLowerCase();
  return items
    .filter((item) => {
      if (!normalized) {
        return true;
      }
      return (
        item.label.toLowerCase().includes(normalized) ||
        (item.detail?.toLowerCase().includes(normalized) ?? false) ||
        item.id.toLowerCase().includes(normalized)
      );
    })
    .slice(0, 8)
    .map((item) => ({
      id: item.id,
      label: item.label,
      detail: item.detail,
    }));
}

function filterEmojiItems(query: string): SuggestionListItem[] {
  const normalized = query.trim().toLowerCase();
  return emojis
    .filter((item) => {
      if (!item.emoji) {
        return false;
      }
      if (!normalized) {
        return true;
      }
      return (
        item.name.includes(normalized) ||
        item.shortcodes.some((code) => code.includes(normalized)) ||
        item.tags.some((tag) => tag.includes(normalized))
      );
    })
    .slice(0, 8)
    .map((item) => ({
      id: item.name,
      label: `${item.emoji} :${item.name}:`,
    }));
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
    () => [
      StarterKit.configure({
        heading: { levels: [1, 2, 3] },
        codeBlock: false,
      }),
      CodeBlockLowlight.configure({
        lowlight,
        defaultLanguage: "plaintext",
      }),
      Link.configure({
        openOnClick: false,
        enableClickSelection: true,
        autolink: true,
        defaultProtocol: "https",
        HTMLAttributes: {
          rel: "noopener noreferrer",
          target: "_blank",
        },
      }),
      Underline,
      Emoji.configure({
        enableEmoticons: false,
        suggestion: {
          items: ({ query }) => filterEmojiItems(query),
          render: createSuggestionListRenderer((item) => ({
            name: item.id,
          })),
        },
      }),
      Mention.configure({
        HTMLAttributes: {
          class: "mention",
        },
        suggestion: {
          items: ({ query }) =>
            filterMentionItems(mentionItemsRef.current, query),
          render: createSuggestionListRenderer((item) => ({
            id: item.id,
            label: item.label,
          })),
        },
      }),
      Table.configure({
        resizable: false,
      }),
      TableRow,
      TableHeader,
      TableCell,
      TaskList,
      TaskItem.configure({
        nested: true,
      }),
      Markdown.configure({
        markedOptions: {
          gfm: true,
        },
      }),
      Placeholder.configure({
        placeholder,
      }),
    ],
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
        // TipTap openOnClick:false skips window.open but does not cancel the
        // browser default for <a target="_blank">, so block navigation while editing.
        click: (view, event) => {
          if (!view.editable) {
            return false;
          }
          const target = event.target;
          if (!(target instanceof Element)) {
            return false;
          }
          const link = target.closest("a");
          if (link && view.dom.contains(link)) {
            event.preventDefault();
          }
          return false;
        },
        auxclick: (view, event) => {
          if (!view.editable || event.button !== 1) {
            return false;
          }
          const target = event.target;
          if (!(target instanceof Element)) {
            return false;
          }
          const link = target.closest("a");
          if (link && view.dom.contains(link)) {
            event.preventDefault();
          }
          return false;
        },
      },
    },
    onCreate: ({ editor: current }) => {
      onChange?.({
        markdown: current.getMarkdown(),
        plainText: current.getText(),
      });
      onReady?.();
    },
    onUpdate: ({ editor: current }) => {
      onChange?.({
        markdown: current.getMarkdown(),
        plainText: current.getText(),
      });
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
