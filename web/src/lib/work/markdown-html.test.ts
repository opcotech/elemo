import { describe, expect, it } from "vitest";

import { markdownToSafeHtml } from "@/lib/work/markdown-html";

describe("markdownToSafeHtml", () => {
  it("returns empty string for blank markdown", () => {
    expect(markdownToSafeHtml("  ")).toBe("");
  });

  it("renders markdown and strips unsafe HTML", () => {
    const html = markdownToSafeHtml(
      "Hello **world**\n\n<script>alert(1)</script>"
    );
    expect(html).toContain("<strong>world</strong>");
    expect(html).not.toContain("<script>");
  });

  it("expands mention shortcodes and highlights fenced code", () => {
    const html = markdownToSafeHtml(
      'Hi [@ id="user-1" label="Ada Lovelace"]\n\n```javascript\nconst x = 1\n```'
    );
    expect(html).toContain('data-mention-id="user-1"');
    expect(html).toContain("@Ada Lovelace");
    expect(html).toContain("hljs");
    expect(html).toContain("language-javascript");
  });
});
