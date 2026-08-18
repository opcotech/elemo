import type { Editor } from "@tiptap/react";

export interface LinkDialogDraft {
  href: string;
  label: string;
  from: number;
  to: number;
  empty: boolean;
  selectedText: string;
  editing: boolean;
}

export function captureLinkDraft(editor: Editor): LinkDialogDraft {
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

export function applyLinkDraft(editor: Editor, draft: LinkDialogDraft) {
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

export function removeLinkDraft(editor: Editor, draft: LinkDialogDraft) {
  editor
    .chain()
    .focus()
    .setTextSelection({ from: draft.from, to: draft.to })
    .extendMarkRange("link")
    .unsetLink()
    .run();
}
