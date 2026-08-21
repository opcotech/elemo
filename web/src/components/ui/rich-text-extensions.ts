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
import type { AnyExtension, Editor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import { common, createLowlight } from "lowlight";

import { createSuggestionListRenderer } from "@/components/ui/rich-text-suggestion";
import type { SuggestionListItem } from "@/components/ui/rich-text-suggestion";

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

export type RichTextEditorFeatures = {
  headings: boolean;
  history: boolean;
  tables: boolean;
  emoji: boolean;
  horizontalRule: boolean;
};

export const DEFAULT_RICH_TEXT_FEATURES: RichTextEditorFeatures = {
  headings: true,
  history: true,
  tables: true,
  emoji: true,
  horizontalRule: true,
};

export type RichTextPlaceholder =
  string | ((props: { node: { type: { name: string } } }) => string);

export function resolveRichTextFeatures(
  features?: Partial<RichTextEditorFeatures>
): RichTextEditorFeatures {
  return { ...DEFAULT_RICH_TEXT_FEATURES, ...features };
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

export function editorValueFrom(editor: Editor): RichTextEditorValue {
  return {
    markdown: editor.getMarkdown(),
    plainText: editor.getText(),
  };
}

export function preventLinkNavigation(
  view: { editable: boolean; dom: Node },
  event: Event
): boolean {
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
}

export function createRichTextExtensions(options: {
  placeholder: RichTextPlaceholder;
  getMentionItems: () => readonly RichTextMentionItem[];
  features?: Partial<RichTextEditorFeatures>;
}): AnyExtension[] {
  const features = resolveRichTextFeatures(options.features);

  const extensions: AnyExtension[] = [
    StarterKit.configure({
      heading: features.headings ? { levels: [1, 2, 3] } : false,
      undoRedo: features.history ? {} : false,
      horizontalRule: features.horizontalRule ? {} : false,
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
  ];

  if (features.emoji) {
    extensions.push(
      Emoji.configure({
        enableEmoticons: false,
        suggestion: {
          items: ({ query }) => filterEmojiItems(query),
          render: createSuggestionListRenderer((item) => ({
            name: item.id,
          })),
        },
      })
    );
  }

  extensions.push(
    Mention.configure({
      HTMLAttributes: {
        class: "mention",
      },
      suggestion: {
        items: ({ query }) =>
          filterMentionItems(options.getMentionItems(), query),
        render: createSuggestionListRenderer((item) => ({
          id: item.id,
          label: item.label,
        })),
      },
    })
  );

  if (features.tables) {
    extensions.push(
      Table.configure({
        resizable: false,
      }),
      TableRow,
      TableHeader,
      TableCell
    );
  }

  extensions.push(
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
      placeholder: options.placeholder,
    })
  );

  return extensions;
}
