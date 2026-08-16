import { FlaskConicalIcon } from "lucide-react";
import type { ReactNode } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { cn } from "@/lib/utils";

export function MockDataAlert({
  title = "Illustrative data",
  children,
  className,
}: {
  title?: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <Alert
      variant="warning"
      className={cn("max-w-full min-w-0 overflow-hidden py-2.5", className)}
    >
      <FlaskConicalIcon />
      <AlertTitle className="min-w-0">{title}</AlertTitle>
      <AlertDescription className="min-w-0 wrap-break-word">
        {children}
      </AlertDescription>
    </Alert>
  );
}
