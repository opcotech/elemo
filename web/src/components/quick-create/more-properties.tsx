import type { ReactNode } from "react";

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";

interface MorePropertiesProps {
  children: ReactNode;
}

export function MoreProperties({ children }: MorePropertiesProps) {
  return (
    <Accordion>
      <AccordionItem value="more" className="border-0">
        <AccordionTrigger className="text-muted-foreground font-normal hover:no-underline">
          More properties
        </AccordionTrigger>
        <AccordionContent>
          <div className="flex flex-col gap-4">{children}</div>
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  );
}
