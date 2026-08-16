import { useEffect, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import { parseIssueTitle } from "@/lib/work/issue-edit";

interface IssueInlineTitleProps {
  title: string;
  disabled?: boolean;
  onCommit: (title: string) => Promise<void>;
}

export function IssueInlineTitle({
  title,
  disabled = false,
  onCommit,
}: IssueInlineTitleProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(title);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const skipBlurCommit = useRef(false);

  useEffect(() => {
    if (!editing) {
      setDraft(title);
      setError(null);
    }
  }, [title, editing]);

  useEffect(() => {
    if (editing) {
      inputRef.current?.focus();
      inputRef.current?.select();
    }
  }, [editing]);

  const cancel = () => {
    skipBlurCommit.current = true;
    setDraft(title);
    setError(null);
    setEditing(false);
  };

  const commit = async () => {
    if (saving || disabled) {
      return;
    }

    const parsed = parseIssueTitle(draft);
    if (!parsed.ok) {
      setError(parsed.error);
      inputRef.current?.focus();
      return;
    }

    if (parsed.title === title) {
      setEditing(false);
      setError(null);
      return;
    }

    setSaving(true);
    try {
      await onCommit(parsed.title);
      setEditing(false);
      setError(null);
    } catch {
      setError("Could not save title");
      inputRef.current?.focus();
    } finally {
      setSaving(false);
    }
  };

  if (!editing) {
    return (
      <button
        type="button"
        className={cn(
          "hover:bg-primary/5 hover:border-primary/20 hover:text-foreground dark:hover:bg-primary/10 dark:hover:border-primary/30 -mx-1 w-full cursor-text rounded-md border border-transparent px-1 text-left transition-colors",
          disabled && "pointer-events-none"
        )}
        onClick={() => {
          if (!disabled) {
            setEditing(true);
          }
        }}
        disabled={disabled}
        aria-label="Edit title"
      >
        {title}
      </button>
    );
  }

  return (
    <span className="block w-full">
      <input
        ref={inputRef}
        value={draft}
        disabled={saving || disabled}
        aria-invalid={error ? true : undefined}
        aria-label="Issue title"
        className={cn(
          "placeholder:text-muted-foreground w-full min-w-0 bg-transparent p-px outline-none",
          error && "text-destructive"
        )}
        onChange={(event) => {
          setDraft(event.target.value);
          if (error) {
            setError(null);
          }
        }}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            event.preventDefault();
            void commit();
          }
          if (event.key === "Escape") {
            event.preventDefault();
            cancel();
          }
        }}
        onBlur={() => {
          if (skipBlurCommit.current) {
            skipBlurCommit.current = false;
            return;
          }
          void commit();
        }}
      />
      {error && (
        <span className="text-destructive mt-1 block text-sm font-normal">
          {error}
        </span>
      )}
    </span>
  );
}
