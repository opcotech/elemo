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
          "w-full min-w-0 bg-transparent font-bold text-4xl leading-tight tracking-tight outline-none placeholder:text-muted-foreground sm:text-5xl",
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
        <span className="mt-1 block font-normal text-destructive text-sm">
          {error}
        </span>
      ) : null}
    </span>
  );
}
