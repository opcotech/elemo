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
          "field-sizing-content mt-3 w-full min-w-0 resize-none overflow-hidden bg-transparent font-normal text-lg text-muted-foreground leading-8 outline-none placeholder:text-muted-foreground",
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
        <span className="mt-1 block font-normal text-destructive text-sm">
          {error}
        </span>
      ) : null}
    </span>
  );
}
