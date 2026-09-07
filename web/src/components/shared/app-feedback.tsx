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
      className={cn("min-w-0 max-w-full overflow-hidden py-2.5", className)}
    >
      <FlaskConicalIcon />
      <AlertTitle className="min-w-0">{title}</AlertTitle>
      <AlertDescription className="wrap-break-word min-w-0">
        {children}
      </AlertDescription>
    </Alert>
  );
}
