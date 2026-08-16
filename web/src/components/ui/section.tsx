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
                <h2 className="text-sm font-semibold tracking-wide uppercase">
                  {title}
                </h2>
              )}
              {description && (
                <p className="text-muted-foreground mt-0.5 text-xs">
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
  children,
}: {
  title: string;
  value: string;
  defaultOpen?: boolean;
  children: ReactNode;
}) {
  return (
    <Accordion defaultValue={defaultOpen ? [value] : []}>
      <AccordionItem value={value} className="border-0">
        <AccordionTrigger className="text-foreground mb-0 min-h-8 py-0 text-sm font-semibold tracking-wide uppercase hover:no-underline">
          {title}
        </AccordionTrigger>
        <AccordionContent className="pt-3 pb-0">{children}</AccordionContent>
      </AccordionItem>
    </Accordion>
  );
}
