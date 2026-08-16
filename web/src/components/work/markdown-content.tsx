import type { ReactNode } from "react";

import { cn } from "@/lib/utils";
import { markdownToSafeHtml } from "@/lib/work/markdown-html";

export function MarkdownContent({
  markdown,
  className,
  empty,
  size = "sm",
}: {
  markdown: string | null | undefined;
  className?: string;
  empty?: ReactNode;
  /**
   * `sm` / `xs`: muted secondary copy (inspector, cards).
   * `default`: full body description (details page).
   */
  size?: "xs" | "sm" | "default";
}) {
  const trimmed = markdown?.trim() ?? "";
  if (!trimmed) {
    return empty ? <>{empty}</> : null;
  }

  return (
    <div
      data-size={size === "default" ? undefined : size}
      className={cn("rich-text-content", className)}
      dangerouslySetInnerHTML={{ __html: markdownToSafeHtml(trimmed) }}
    />
  );
}
