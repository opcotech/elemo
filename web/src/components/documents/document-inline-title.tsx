import { cn } from "@/lib/utils";

interface DocumentInlineTitleProps {
  value: string;
  disabled?: boolean;
  error?: string | null;
  onChange: (value: string) => void;
  onReset?: () => void;
}

export function DocumentInlineTitle({
  value,
  disabled = false,
  error,
  onChange,
  onReset,
}: DocumentInlineTitleProps) {
  return (
    <span className="block w-full">
      <input
        value={value}
        disabled={disabled}
        aria-invalid={error ? true : undefined}
        aria-label="Document title"
        placeholder="Untitled"
        className={cn(
          "placeholder:text-muted-foreground w-full min-w-0 bg-transparent text-4xl leading-tight font-bold tracking-tight outline-none sm:text-5xl",
          error && "text-destructive"
        )}
        onChange={(event) => {
          onChange(event.target.value);
        }}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            event.preventDefault();
            event.currentTarget.blur();
          }
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
