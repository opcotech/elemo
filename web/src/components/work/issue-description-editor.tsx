import { useEffect, useMemo, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { RichTextEditor } from "@/components/ui/rich-text-editor";
import type {
  RichTextEditorValue,
  RichTextMentionItem,
} from "@/components/ui/rich-text-editor";
import { cn, getDefaultValue } from "@/lib/utils";
import { parseIssueDescription } from "@/lib/work/issue-edit";
import { markdownToSafeHtml } from "@/lib/work/markdown-html";
import { useOrganizationMembersForNamespace } from "@/lib/work/use-organization-members-for-namespace";

const EDIT_CLICK_DELAY_MS = 300;

interface IssueDescriptionEditorProps {
  description: string | null | undefined;
  namespaceId?: string;
  disabled?: boolean;
  onCommit: (description: string | null) => Promise<void>;
}

function draftFromDescription(
  description: string | null | undefined
): RichTextEditorValue {
  const markdown = getDefaultValue(description);
  return {
    markdown,
    plainText: markdown,
  };
}

function hasSelectionInElement(element: HTMLElement | null): boolean {
  if (!element) {
    return false;
  }

  const selection = window.getSelection();
  if (!selection || selection.isCollapsed || selection.rangeCount === 0) {
    return false;
  }

  const range = selection.getRangeAt(0);
  return element.contains(range.commonAncestorContainer);
}

export function IssueDescriptionEditor({
  description,
  namespaceId,
  disabled = false,
  onCommit,
}: IssueDescriptionEditorProps) {
  const [editing, setEditing] = useState(false);
  const [editorKey, setEditorKey] = useState(0);
  const [seed, setSeed] = useState(() => getDefaultValue(description));
  const [draft, setDraft] = useState<RichTextEditorValue>(() =>
    draftFromDescription(description)
  );
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const surfaceRef = useRef<HTMLDivElement>(null);
  const openTimerRef = useRef<number | null>(null);
  const { members } = useOrganizationMembersForNamespace(namespaceId, {
    enabled: editing,
  });

  const mentionItems = useMemo<RichTextMentionItem[]>(
    () =>
      members.map((member) => {
        const label =
          `${member.first_name} ${member.last_name}`.trim() || member.email;
        return {
          id: member.id,
          label,
          detail: member.email,
        };
      }),
    [members]
  );

  const clearOpenTimer = () => {
    if (openTimerRef.current != null) {
      window.clearTimeout(openTimerRef.current);
      openTimerRef.current = null;
    }
  };

  useEffect(() => {
    if (!editing) {
      setSeed(getDefaultValue(description));
      setDraft(draftFromDescription(description));
      setError(null);
    }
  }, [description, editing]);

  useEffect(() => clearOpenTimer, []);

  const isEmpty = !description?.trim();
  const safeHtml = isEmpty ? "" : markdownToSafeHtml(description ?? "");

  const beginEditing = () => {
    if (disabled) {
      return;
    }
    clearOpenTimer();
    const nextSeed = getDefaultValue(description);
    setSeed(nextSeed);
    setDraft(draftFromDescription(description));
    setEditorKey((key) => key + 1);
    setError(null);
    setEditing(true);
  };

  const scheduleBeginEditing = () => {
    if (disabled) {
      return;
    }

    clearOpenTimer();
    openTimerRef.current = window.setTimeout(() => {
      openTimerRef.current = null;
      if (hasSelectionInElement(surfaceRef.current)) {
        return;
      }
      beginEditing();
    }, EDIT_CLICK_DELAY_MS);
  };

  const cancel = () => {
    setDraft(draftFromDescription(description));
    setError(null);
    setEditing(false);
  };

  const save = async () => {
    if (saving || disabled) {
      return;
    }

    const parsed = parseIssueDescription(draft.plainText, draft.markdown);
    if (!parsed.ok) {
      setError(parsed.error);
      return;
    }

    const current = description?.trim() || null;
    if (parsed.description === current) {
      setEditing(false);
      setError(null);
      return;
    }

    setSaving(true);
    try {
      await onCommit(parsed.description);
      setEditing(false);
      setError(null);
    } catch {
      setError("Could not save description");
    } finally {
      setSaving(false);
    }
  };

  if (!editing) {
    return (
      <div
        ref={surfaceRef}
        role="button"
        tabIndex={disabled ? -1 : 0}
        aria-label="Edit description"
        aria-disabled={disabled || undefined}
        className={cn(
          "bg-card hover:bg-primary/5 hover:border-primary/20 min-h-40 w-full cursor-text rounded-xl border p-5 text-left transition-colors select-text",
          "focus-visible:ring-ring/50 focus-visible:ring-2 focus-visible:outline-none",
          disabled && "pointer-events-none opacity-60"
        )}
        onPointerDown={(event) => {
          if (event.button !== 0 || disabled) {
            return;
          }
          clearOpenTimer();
        }}
        onPointerUp={(event) => {
          if (event.button !== 0 || disabled) {
            return;
          }
          scheduleBeginEditing();
        }}
        onKeyDown={(event) => {
          if (disabled) {
            return;
          }
          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            beginEditing();
          }
        }}
      >
        {isEmpty ? (
          <p className="text-muted-foreground leading-7">Add a description…</p>
        ) : (
          <div
            className="rich-text-content leading-7"
            dangerouslySetInnerHTML={{ __html: safeHtml }}
          />
        )}
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <RichTextEditor
        key={editorKey}
        content={seed}
        disabled={saving || disabled}
        autoFocus
        aria-label="Issue description"
        placeholder="Describe the work…"
        mentionItems={mentionItems}
        fallback={
          isEmpty ? (
            <p className="text-muted-foreground leading-7">
              Add a description…
            </p>
          ) : (
            <div
              className="rich-text-content leading-7"
              dangerouslySetInnerHTML={{ __html: safeHtml }}
            />
          )
        }
        onChange={(value) => {
          setDraft(value);
          if (error) {
            setError(null);
          }
        }}
      />
      {error && <p className="text-destructive text-sm">{error}</p>}
      <div className="flex items-center justify-end gap-2">
        <Button
          type="button"
          variant="outline"
          disabled={saving}
          onClick={cancel}
        >
          Cancel
        </Button>
        <Button
          type="button"
          disabled={saving || disabled}
          onClick={() => {
            void save();
          }}
        >
          {saving ? "Saving…" : "Save"}
        </Button>
      </div>
    </div>
  );
}
