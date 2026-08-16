import { XIcon } from "lucide-react";
import type { ReactNode } from "react";
import { useEffect, useRef } from "react";

import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { cn } from "@/lib/utils";

interface ResponsiveInspectorShellProps {
  children: ReactNode;
  inspector?: ReactNode;
  inspectorTitle?: string;
  inspectorDescription?: string;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  className?: string;
}

export function ResponsiveInspectorShell({
  children,
  inspector,
  inspectorTitle = "Details",
  inspectorDescription,
  open: controlledOpen,
  onOpenChange,
  className,
}: ResponsiveInspectorShellProps) {
  const open = controlledOpen ?? Boolean(inspector);
  const setOpen = onOpenChange ?? (() => undefined);
  const closeButtonRef = useRef<HTMLButtonElement>(null);
  const showInspector = Boolean(inspector) && open;

  useEffect(() => {
    if (!showInspector) return;

    const frame = window.requestAnimationFrame(() =>
      closeButtonRef.current?.focus()
    );
    return () => window.cancelAnimationFrame(frame);
  }, [inspectorTitle, showInspector]);

  return (
    <div
      className={cn(
        "h-full min-h-0 min-w-0 flex-1 overflow-x-hidden overflow-y-auto",
        className
      )}
    >
      {children}
      <Sheet open={showInspector} onOpenChange={setOpen}>
        <SheetContent
          side="right"
          showCloseButton={false}
          className="bg-surface-raised w-full gap-0 p-0 data-[side=right]:w-full data-[side=right]:max-w-none sm:data-[side=right]:w-137.5 sm:data-[side=right]:max-w-187.5 sm:data-[side=right]:min-w-137.5"
        >
          <SheetHeader className="sr-only">
            <SheetTitle>{inspectorTitle}</SheetTitle>
            <SheetDescription>
              {inspectorDescription ?? "Contextual details"}
            </SheetDescription>
          </SheetHeader>
          <div
            className="relative flex h-full min-h-0 flex-col"
            role="complementary"
            aria-label={`${inspectorTitle} details`}
            data-section="work-inspector"
          >
            <Button
              ref={closeButtonRef}
              type="button"
              variant="ghost"
              size="icon-sm"
              className="absolute top-3 right-3 z-10 rounded-full"
              onClick={() => setOpen(false)}
              aria-label="Close inspector"
            >
              <XIcon />
            </Button>
            <div className="min-h-0 flex-1 overflow-y-auto">{inspector}</div>
          </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}

export type { ResponsiveInspectorShellProps };
