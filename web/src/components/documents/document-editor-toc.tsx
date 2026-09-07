import type { Editor } from "@tiptap/react";
import { useEditorState } from "@tiptap/react";
import { PanelLeftCloseIcon } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  activeDocumentHeadingPos,
  collectDocumentHeadings,
} from "@/lib/documents/document-headings";
import { cn } from "@/lib/utils";

const TOC_STORAGE_KEY = "elemo_document_toc_open";

export function useDocumentTocOpen() {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (window.localStorage.getItem(TOC_STORAGE_KEY) === "true") {
      setOpen(true);
    }
  }, []);

  const setTocOpen = (next: boolean | ((current: boolean) => boolean)) => {
    setOpen((current) => {
      const value = typeof next === "function" ? next(current) : next;
      window.localStorage.setItem(TOC_STORAGE_KEY, String(value));
      return value;
    });
  };

  return [open, setTocOpen] as const;
}

const idleTocState = {
  headings: [] as ReturnType<typeof collectDocumentHeadings>,
  activePos: null as number | null,
};

const levelPad: Record<1 | 2 | 3, string> = {
  1: "pl-2",
  2: "pl-4",
  3: "pl-6",
};

export function DocumentEditorToc({
  editor,
  onClose,
}: {
  editor: Editor | null;
  onClose: () => void;
}) {
  const tocState = useEditorState({
    editor,
    selector: ({ editor: current }) => {
      if (!current) {
        return idleTocState;
      }
      const headings = collectDocumentHeadings(current.state.doc);
      return {
        headings,
        activePos: activeDocumentHeadingPos(
          headings,
          current.state.selection.from
        ),
      };
    },
  });

  const headings = tocState?.headings ?? idleTocState.headings;
  const activePos = tocState?.activePos ?? null;

  return (
    <nav
      aria-label="Table of contents"
      className="sticky top-16 hidden max-h-[calc(100svh-8rem)] w-56 shrink-0 self-start overflow-y-auto lg:block"
    >
      <div className="mb-1 flex items-center justify-between gap-1 px-1">
        <p className="font-medium text-muted-foreground text-xs">Contents</p>
        <Button
          type="button"
          variant="ghost"
          size="icon-xs"
          aria-label="Hide table of contents"
          title="Hide table of contents"
          onClick={onClose}
        >
          <PanelLeftCloseIcon />
        </Button>
      </div>
      {headings.length === 0 ? (
        <p className="px-2 py-1 text-muted-foreground text-sm">
          Headings will appear here
        </p>
      ) : (
        <ul className="flex flex-col gap-0.5">
          {headings.map((heading) => {
            const active = heading.pos === activePos;
            return (
              <li key={heading.pos}>
                <button
                  type="button"
                  aria-current={active ? "true" : undefined}
                  className={cn(
                    "w-full truncate rounded-md py-1 text-left text-sm hover:bg-primary/10 hover:text-primary-on-subtle",
                    levelPad[heading.level],
                    active
                      ? "bg-primary/10 font-medium text-primary-on-subtle"
                      : "text-muted-foreground"
                  )}
                  onClick={() => {
                    editor
                      ?.chain()
                      .focus()
                      .setTextSelection(heading.pos + 1)
                      .scrollIntoView()
                      .run();
                  }}
                >
                  {heading.text}
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </nav>
  );
}
