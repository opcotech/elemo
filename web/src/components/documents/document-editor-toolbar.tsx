import { useEditorState } from "@tiptap/react";
import type { Editor } from "@tiptap/react";
import {
  BetweenHorizontalEndIcon,
  BetweenVerticalEndIcon,
  BoldIcon,
  CodeIcon,
  Grid2X2XIcon,
  Heading1Icon,
  Heading2Icon,
  Heading3Icon,
  HighlighterIcon,
  IndentDecreaseIcon,
  IndentIncreaseIcon,
  ItalicIcon,
  LinkIcon,
  ListChecksIcon,
  ListIcon,
  ListOrderedIcon,
  MinusIcon,
  QuoteIcon,
  Redo2Icon,
  SquareCodeIcon,
  StrikethroughIcon,
  TableColumnsSplitIcon,
  TableIcon,
  TableRowsSplitIcon,
  UnderlineIcon,
  Undo2Icon,
} from "lucide-react";
import { useState } from "react";
import type { ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { LinkAddDialog } from "@/components/ui/link-add-dialog";
import { CODE_BLOCK_LANGUAGES } from "@/components/ui/rich-text-extensions";
import {
  applyLinkDraft,
  captureLinkDraft,
  removeLinkDraft,
} from "@/components/ui/rich-text-link";
import type { LinkDialogDraft } from "@/components/ui/rich-text-link";
import { cn } from "@/lib/utils";

const LIST_ITEM_TYPES = ["listItem", "taskItem"] as const;

function canSinkListItem(editor: Editor): boolean {
  return LIST_ITEM_TYPES.some((type) => editor.can().sinkListItem(type));
}

function canLiftListItem(editor: Editor): boolean {
  return LIST_ITEM_TYPES.some((type) => editor.can().liftListItem(type));
}

function sinkListItem(editor: Editor) {
  const type = LIST_ITEM_TYPES.find((itemType) =>
    editor.can().sinkListItem(itemType)
  );
  if (!type) {
    return;
  }
  editor.chain().focus().sinkListItem(type).run();
}

function liftListItem(editor: Editor) {
  const type = LIST_ITEM_TYPES.find((itemType) =>
    editor.can().liftListItem(itemType)
  );
  if (!type) {
    return;
  }
  editor.chain().focus().liftListItem(type).run();
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
      size="icon-sm"
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
      className="bg-border mx-0.5 hidden h-4 w-px shrink-0 sm:inline-block"
    />
  );
}

const idleToolbarState = {
  bold: false,
  italic: false,
  underline: false,
  strike: false,
  highlight: false,
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
  canDeleteTable: false,
  canIndent: false,
  canOutdent: false,
};

export function DocumentEditorToolbar({
  editor,
  disabled = false,
  className,
}: {
  editor: Editor | null;
  disabled?: boolean;
  className?: string;
}) {
  const [linkDialogOpen, setLinkDialogOpen] = useState(false);
  const [linkDraft, setLinkDraft] = useState<LinkDialogDraft | null>(null);
  const toolbarState = useEditorState({
    editor,
    selector: ({ editor: current }) => {
      if (!current) {
        return idleToolbarState;
      }
      return {
        bold: current.isActive("bold"),
        italic: current.isActive("italic"),
        underline: current.isActive("underline"),
        strike: current.isActive("strike"),
        highlight: current.isActive("highlight"),
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
        canDeleteTable: current.can().deleteTable(),
        canIndent: canSinkListItem(current),
        canOutdent: canLiftListItem(current),
      };
    },
  });

  const locked = !editor || disabled;

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
    <>
      <div
        className={cn(
          "document-editor__toolbar flex min-w-0 [scrollbar-width:none] items-center gap-0.5 overflow-x-auto [&::-webkit-scrollbar]:hidden",
          className
        )}
      >
        <ToolbarButton
          label="Undo"
          disabled={locked || !toolbarState?.canUndo}
          onClick={() => editor?.chain().focus().undo().run()}
        >
          <Undo2Icon />
        </ToolbarButton>
        <ToolbarButton
          label="Redo"
          disabled={locked || !toolbarState?.canRedo}
          onClick={() => editor?.chain().focus().redo().run()}
        >
          <Redo2Icon />
        </ToolbarButton>
        <ToolbarDivider />
        <ToolbarButton
          label="Heading 1"
          active={toolbarState?.heading1}
          disabled={locked}
          onClick={() =>
            editor?.chain().focus().toggleHeading({ level: 1 }).run()
          }
        >
          <Heading1Icon />
        </ToolbarButton>
        <ToolbarButton
          label="Heading 2"
          active={toolbarState?.heading2}
          disabled={locked}
          onClick={() =>
            editor?.chain().focus().toggleHeading({ level: 2 }).run()
          }
        >
          <Heading2Icon />
        </ToolbarButton>
        <ToolbarButton
          label="Heading 3"
          active={toolbarState?.heading3}
          disabled={locked}
          onClick={() =>
            editor?.chain().focus().toggleHeading({ level: 3 }).run()
          }
        >
          <Heading3Icon />
        </ToolbarButton>
        <ToolbarDivider />
        <ToolbarButton
          label="Bold"
          active={toolbarState?.bold}
          disabled={locked}
          onClick={() => editor?.chain().focus().toggleBold().run()}
        >
          <BoldIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Italic"
          active={toolbarState?.italic}
          disabled={locked}
          onClick={() => editor?.chain().focus().toggleItalic().run()}
        >
          <ItalicIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Underline"
          active={toolbarState?.underline}
          disabled={locked}
          onClick={() => editor?.chain().focus().toggleUnderline().run()}
        >
          <UnderlineIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Strikethrough"
          active={toolbarState?.strike}
          disabled={locked}
          onClick={() => editor?.chain().focus().toggleStrike().run()}
        >
          <StrikethroughIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Highlight"
          active={toolbarState?.highlight}
          disabled={locked}
          onClick={() => editor?.chain().focus().toggleHighlight().run()}
        >
          <HighlighterIcon />
        </ToolbarButton>
        <ToolbarDivider />
        <ToolbarButton
          label="Bullet list"
          active={toolbarState?.bulletList}
          disabled={locked}
          onClick={() => editor?.chain().focus().toggleBulletList().run()}
        >
          <ListIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Ordered list"
          active={toolbarState?.orderedList}
          disabled={locked}
          onClick={() => editor?.chain().focus().toggleOrderedList().run()}
        >
          <ListOrderedIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Task list"
          active={toolbarState?.taskList}
          disabled={locked}
          onClick={() => editor?.chain().focus().toggleTaskList().run()}
        >
          <ListChecksIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Decrease indent"
          disabled={locked || !toolbarState?.canOutdent}
          onClick={() => {
            if (editor) {
              liftListItem(editor);
            }
          }}
        >
          <IndentDecreaseIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Increase indent"
          disabled={locked || !toolbarState?.canIndent}
          onClick={() => {
            if (editor) {
              sinkListItem(editor);
            }
          }}
        >
          <IndentIncreaseIcon />
        </ToolbarButton>
        <ToolbarDivider />
        <ToolbarButton
          label="Quote"
          active={toolbarState?.blockquote}
          disabled={locked}
          onClick={() => editor?.chain().focus().toggleBlockquote().run()}
        >
          <QuoteIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Inline code"
          active={toolbarState?.code}
          disabled={locked}
          onClick={() => editor?.chain().focus().toggleCode().run()}
        >
          <CodeIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Code block"
          active={toolbarState?.codeBlock}
          disabled={locked}
          onClick={() => editor?.chain().focus().toggleCodeBlock().run()}
        >
          <SquareCodeIcon />
        </ToolbarButton>
        {toolbarState?.codeBlock ? (
          <label className="text-muted-foreground ml-1 flex items-center gap-1 text-xs">
            <span className="sr-only">Code language</span>
            <select
              className="border-border bg-background h-8 rounded-md border px-1.5 text-xs"
              value={
                CODE_BLOCK_LANGUAGES.includes(
                  toolbarState.codeLanguage as (typeof CODE_BLOCK_LANGUAGES)[number]
                )
                  ? toolbarState.codeLanguage
                  : "plaintext"
              }
              disabled={locked}
              onChange={(event) => {
                editor
                  ?.chain()
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
          disabled={locked}
          onClick={openLinkDialog}
        >
          <LinkIcon />
        </ToolbarButton>
        <ToolbarDivider />
        <ToolbarButton
          label="Horizontal rule"
          disabled={locked}
          onClick={() => editor?.chain().focus().setHorizontalRule().run()}
        >
          <MinusIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Insert table"
          disabled={locked}
          onClick={() =>
            editor
              ?.chain()
              .focus()
              .insertTable({ rows: 3, cols: 3, withHeaderRow: true })
              .run()
          }
        >
          <TableIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Delete table"
          disabled={locked || !toolbarState?.canDeleteTable}
          onClick={() => editor?.chain().focus().deleteTable().run()}
        >
          <Grid2X2XIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Add row"
          disabled={locked || !toolbarState?.canAddRow}
          onClick={() => editor?.chain().focus().addRowAfter().run()}
        >
          <BetweenHorizontalEndIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Delete row"
          disabled={locked || !toolbarState?.canDeleteRow}
          onClick={() => editor?.chain().focus().deleteRow().run()}
        >
          <TableRowsSplitIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Add column"
          disabled={locked || !toolbarState?.canAddColumn}
          onClick={() => editor?.chain().focus().addColumnAfter().run()}
        >
          <BetweenVerticalEndIcon />
        </ToolbarButton>
        <ToolbarButton
          label="Delete column"
          disabled={locked || !toolbarState?.canDeleteColumn}
          onClick={() => editor?.chain().focus().deleteColumn().run()}
        >
          <TableColumnsSplitIcon />
        </ToolbarButton>
      </div>
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
    </>
  );
}
