import type { Node } from "@tiptap/pm/model";

export type DocumentHeadingLevel = 1 | 2 | 3;

export interface DocumentHeading {
  pos: number;
  level: DocumentHeadingLevel;
  text: string;
}

function isHeadingLevel(value: unknown): value is DocumentHeadingLevel {
  return value === 1 || value === 2 || value === 3;
}

/** Collect H1–H3 headings in document order. Empty headings are skipped. */
export function collectDocumentHeadings(doc: Node): DocumentHeading[] {
  const headings: DocumentHeading[] = [];
  doc.descendants((node, pos) => {
    if (node.type.name !== "heading") {
      return;
    }
    const level = node.attrs.level;
    if (!isHeadingLevel(level)) {
      return;
    }
    const text = node.textContent.replace(/\s+/g, " ").trim();
    if (!text) {
      return;
    }
    headings.push({ pos, level, text });
  });
  return headings;
}

/**
 * Heading that owns the cursor: the last heading at or before `cursor`.
 * That covers both standing in the heading and in the body beneath it.
 */
export function activeDocumentHeadingPos(
  headings: readonly DocumentHeading[],
  cursor: number
): number | null {
  let active: number | null = null;
  for (const heading of headings) {
    if (heading.pos <= cursor) {
      active = heading.pos;
    } else {
      break;
    }
  }
  return active;
}
