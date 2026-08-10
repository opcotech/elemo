import { BoxIcon, FlaskConicalIcon } from "lucide-react";
import type { ReactNode } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
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

export function AppEmptyState({
  icon = <BoxIcon />,
  title,
  description,
  action,
  compact = false,
}: {
  icon?: ReactNode;
  title: string;
  description: string;
  action?: ReactNode;
  compact?: boolean;
}) {
  return (
    <Empty className={cn("border", compact ? "min-h-36" : "min-h-56")}>
      <EmptyHeader>
        <EmptyMedia variant="icon">{icon}</EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        <EmptyDescription>{description}</EmptyDescription>
      </EmptyHeader>
      {action && <EmptyContent>{action}</EmptyContent>}
    </Empty>
  );
}
