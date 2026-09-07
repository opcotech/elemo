import type { ReactNode } from "react";

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { cn } from "@/lib/utils";

export function Section({
  title,
  description,
  action,
  children,
  className,
  "data-section": dataSection,
}: {
  title?: ReactNode;
  description?: ReactNode;
  action?: ReactNode;
  children: ReactNode;
  className?: string;
  "data-section"?: string;
}) {
  return (
    <section className={cn("min-w-0", className)} data-section={dataSection}>
      {(title || description || action) && (
        <div className="mb-3 flex min-h-8 items-center gap-3">
          {(title || description) && (
            <div className="min-w-0 flex-1">
              {title && (
                <h2 className="font-semibold text-sm uppercase tracking-wide">
                  {title}
                </h2>
              )}
              {description && (
                <p className="mt-0.5 text-muted-foreground text-xs">
                  {description}
                </p>
              )}
            </div>
          )}
          {action}
        </div>
      )}
      {children}
    </section>
  );
}

export function SectionAccordion({
  title,
  value,
  defaultOpen = false,
  action,
  children,
}: {
  title: string;
  value: string;
  defaultOpen?: boolean;
  action?: ReactNode;
  children: ReactNode;
}) {
  return (
    <Accordion defaultValue={defaultOpen ? [value] : []}>
      <AccordionItem value={value} className="border-0">
        <div className="flex min-h-8 items-center gap-3">
          <AccordionTrigger className="mb-0 min-h-8 flex-1 py-0 font-semibold text-foreground text-sm uppercase tracking-wide hover:no-underline">
            {title}
          </AccordionTrigger>
          {action}
        </div>
        <AccordionContent className="pt-3 pb-0">{children}</AccordionContent>
      </AccordionItem>
    </Accordion>
  );
}
