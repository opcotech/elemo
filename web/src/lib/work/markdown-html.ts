import createDOMPurify from "dompurify";
import hljs from "highlight.js/lib/common";
import { Marked } from "marked";

import { getPurifyWindow } from "./purify-window";

const MENTIONS_SHORTCODE_RE = /\[@\s+([^\]]*)\]/g;
const ATTR_RE = /(\w+)=(?:"([^"]*)"|'([^']*)')/g;

const DOMPurify = createDOMPurify(getPurifyWindow());

function escapeHtml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function parseShortcodeAttributes(attrString: string): Record<string, string> {
  const attrs: Record<string, string> = {};
  ATTR_RE.lastIndex = 0;
  let match = ATTR_RE.exec(attrString);
  while (match !== null) {
    const key = match[1];
    const value = match[2] ?? match[3] ?? "";
    if (key) {
      attrs[key] = value;
    }
    match = ATTR_RE.exec(attrString);
  }
  return attrs;
}

/** Replace TipTap mention shortcodes with safe HTML chips before Markdown parse. */
export function expandMentionShortcodes(markdown: string): string {
  return markdown.replace(
    MENTIONS_SHORTCODE_RE,
    (_full, attrString: string) => {
      const attrs = parseShortcodeAttributes(attrString);
      const id = attrs.id?.trim() ?? "";
      const label = (attrs.label?.trim() || id || "mention").replaceAll(
        "@",
        ""
      );
      const idAttr = id ? ` data-mention-id="${escapeHtml(id)}"` : "";
      return `<span class="mention"${idAttr}>@${escapeHtml(label)}</span>`;
    }
  );
}

function highlightCode(code: string, language: string | undefined): string {
  const normalized = language?.trim().toLowerCase();
  if (normalized && hljs.getLanguage(normalized)) {
    return hljs.highlight(code, { language: normalized }).value;
  }
  return hljs.highlightAuto(code).value;
}

const marked = new Marked({
  gfm: true,
  breaks: false,
  renderer: {
    code({ text, lang }: { text: string; lang?: string }) {
      const language = lang?.trim() || undefined;
      const highlighted = highlightCode(text, language);
      const className = language
        ? `hljs language-${escapeHtml(language)}`
        : "hljs";
      return `<pre><code class="${className}">${highlighted}</code></pre>\n`;
    },
  },
});

/** Convert Markdown to sanitized HTML for safe view-mode rendering. */
export function markdownToSafeHtml(markdown: string): string {
  const trimmed = markdown.trim();
  if (trimmed.length === 0) {
    return "";
  }

  const withMentions = expandMentionShortcodes(trimmed);
  const html = marked.parse(withMentions, { async: false });
  return DOMPurify.sanitize(typeof html === "string" ? html : String(html), {
    ADD_ATTR: ["data-mention-id"],
  });
}
