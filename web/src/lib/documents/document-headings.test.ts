import { Schema } from "@tiptap/pm/model";
import { describe, expect, it } from "vitest";

import {
  activeDocumentHeadingPos,
  collectDocumentHeadings,
} from "@/lib/documents/document-headings";

const schema = new Schema({
  nodes: {
    doc: { content: "block+" },
    paragraph: { content: "inline*", group: "block" },
    heading: {
      attrs: { level: { default: 1 } },
      content: "inline*",
      group: "block",
    },
    text: { group: "inline" },
  },
});

function heading(level: 1 | 2 | 3, text: string) {
  return schema.node("heading", { level }, text ? [schema.text(text)] : []);
}

function paragraph(text: string) {
  return schema.node("paragraph", null, text ? [schema.text(text)] : []);
}

describe("collectDocumentHeadings", () => {
  it("returns headings in document order with pos, level, and text", () => {
    const doc = schema.node("doc", null, [
      heading(1, "Intro"),
      paragraph("body"),
      heading(2, "Details"),
      heading(3, "Notes"),
    ]);

    const headings = collectDocumentHeadings(doc);
    expect(headings.map(({ level, text }) => ({ level, text }))).toEqual([
      { level: 1, text: "Intro" },
      { level: 2, text: "Details" },
      { level: 3, text: "Notes" },
    ]);
    expect(headings[0]?.pos).toBe(0);
    expect(headings[1]?.pos).toBeGreaterThan(headings[0]?.pos ?? 0);
    expect(headings[2]?.pos).toBeGreaterThan(headings[1]?.pos ?? 0);
  });

  it("skips empty headings and non-heading blocks", () => {
    const doc = schema.node("doc", null, [
      paragraph("lead"),
      heading(1, "   "),
      heading(2, "Keep"),
    ]);

    expect(collectDocumentHeadings(doc)).toEqual([
      {
        pos: expect.any(Number),
        level: 2,
        text: "Keep",
      },
    ]);
  });

  it("collapses inner whitespace in heading text", () => {
    const doc = schema.node("doc", null, [heading(1, "Project   Plan")]);
    expect(collectDocumentHeadings(doc)[0]?.text).toBe("Project Plan");
  });
});

describe("activeDocumentHeadingPos", () => {
  const headings = [
    { pos: 0, level: 1 as const, text: "Intro" },
    { pos: 20, level: 2 as const, text: "Details" },
    { pos: 40, level: 3 as const, text: "Notes" },
  ];

  it("returns null when the cursor is before the first heading", () => {
    expect(activeDocumentHeadingPos(headings, -1)).toBeNull();
  });

  it("selects the heading that contains or precedes the cursor", () => {
    expect(activeDocumentHeadingPos(headings, 0)).toBe(0);
    expect(activeDocumentHeadingPos(headings, 19)).toBe(0);
    expect(activeDocumentHeadingPos(headings, 20)).toBe(20);
    expect(activeDocumentHeadingPos(headings, 50)).toBe(40);
  });
});
