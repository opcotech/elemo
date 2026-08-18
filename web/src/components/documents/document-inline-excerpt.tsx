import { cn } from "@/lib/utils";

interface DocumentInlineExcerptProps {
  value: string;
  disabled?: boolean;
  error?: string | null;
  onChange: (value: string) => void;
  onReset?: () => void;
}

export function DocumentInlineExcerpt({
  value,
  disabled = false,
  error,
  onChange,
  onReset,
}: DocumentInlineExcerptProps) {
  return (
    <span className="block w-full">
      <textarea
        value={value}
        disabled={disabled}
        rows={1}
        aria-invalid={error ? true : undefined}
        aria-label="Document excerpt"
        placeholder="Add excerpt (optional)…"
        className={cn(
          "placeholder:text-muted-foreground text-muted-foreground mt-3 field-sizing-content w-full min-w-0 resize-none overflow-hidden bg-transparent text-lg leading-8 font-normal outline-none",
          error && "text-destructive"
        )}
        onChange={(event) => {
          onChange(event.target.value);
        }}
        onKeyDown={(event) => {
          if (event.key === "Escape") {
            event.preventDefault();
            onReset?.();
            event.currentTarget.blur();
          }
        }}
      />
      {error ? (
        <span className="text-destructive mt-1 block text-sm font-normal">
          {error}
        </span>
      ) : null}
    </span>
  );
}
