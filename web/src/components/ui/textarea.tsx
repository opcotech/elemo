import * as React from "react";

import { cn } from "@/lib/utils";

const Textarea = React.forwardRef<
  HTMLTextAreaElement,
  React.ComponentProps<"textarea">
>(({ className, ...props }, ref) => {
  return (
    <textarea
      data-slot="textarea"
      className={cn(
        "border-border bg-background placeholder:text-muted-foreground focus-visible:ring-muted focus-visible:border-primary hover:border-primary/50 flex min-h-[80px] w-full resize-none rounded-md border px-3 py-2 text-sm transition-colors focus-visible:ring focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      ref={ref}
      {...props}
    />
  );
});
Textarea.displayName = "Textarea";

// Auto-resizing textarea
const AutoResizeTextarea = React.forwardRef<
  HTMLTextAreaElement,
  React.ComponentProps<"textarea">
>(({ className, ...props }, ref) => {
  const textareaRef = React.useRef<HTMLTextAreaElement>(null);

  React.useImperativeHandle(ref, () => textareaRef.current!);

  const adjustHeight = React.useCallback(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = "auto";
      textarea.style.height = `${textarea.scrollHeight}px`;
    }
  }, []);

  React.useEffect(() => {
    adjustHeight();
  }, [adjustHeight, props.value]);

  return (
    <textarea
      data-slot="textarea"
      className={cn(
        "border-border bg-background placeholder:text-muted-foreground focus-visible:ring-muted focus-visible:border-primary hover:border-primary/50 flex w-full resize-none overflow-hidden rounded-md border-2 px-3 py-2 text-sm transition-colors focus-visible:ring focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      ref={textareaRef}
      onInput={adjustHeight}
      {...props}
    />
  );
});
AutoResizeTextarea.displayName = "AutoResizeTextarea";

export { Textarea, AutoResizeTextarea };
