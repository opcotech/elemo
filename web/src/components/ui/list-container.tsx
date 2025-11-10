import type { ReactNode } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

interface ListContainerProps {
  title: string;
  description: string;
  isLoading?: boolean;
  error?: unknown;
  emptyState?: {
    icon?: ReactNode;
    title: string;
    description: string;
    action?: ReactNode;
  };
  actionButton?: ReactNode;
  searchInput?: ReactNode;
  children: ReactNode;
  className?: string;
  "data-section"?: string;
}

export function ListContainer({
  title,
  description,
  isLoading = false,
  error,
  emptyState,
  actionButton,
  searchInput,
  children,
  className,
  "data-section": dataSection,
}: ListContainerProps) {
  const showEmpty = !isLoading && !error && emptyState;
  const showSearchWithEmpty = showEmpty && searchInput;

  return (
    <Card data-section={dataSection} className={className}>
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <CardTitle>{title}</CardTitle>
            <CardDescription>{description}</CardDescription>
          </div>
          {actionButton}
        </div>
      </CardHeader>
      <CardContent
        className={showEmpty && !showSearchWithEmpty ? undefined : "space-y-4"}
      >
        {error ? (
          <Alert variant="destructive">
            <AlertDescription>
              {error instanceof Error
                ? error.message
                : "Failed to load data. Please try again later."}
            </AlertDescription>
          </Alert>
        ) : showSearchWithEmpty ? (
          <>
            {searchInput && <div>{searchInput}</div>}
            <Empty>
              <EmptyHeader>
                {emptyState.icon && (
                  <EmptyMedia variant="icon">{emptyState.icon}</EmptyMedia>
                )}
                <EmptyTitle>{emptyState.title}</EmptyTitle>
                <EmptyDescription>{emptyState.description}</EmptyDescription>
              </EmptyHeader>
              {emptyState.action && (
                <EmptyContent>{emptyState.action}</EmptyContent>
              )}
            </Empty>
          </>
        ) : showEmpty ? (
          <Empty>
            <EmptyHeader>
              {emptyState.icon && (
                <EmptyMedia variant="icon">{emptyState.icon}</EmptyMedia>
              )}
              <EmptyTitle>{emptyState.title}</EmptyTitle>
              <EmptyDescription>{emptyState.description}</EmptyDescription>
            </EmptyHeader>
            {emptyState.action && (
              <EmptyContent>{emptyState.action}</EmptyContent>
            )}
          </Empty>
        ) : (
          <>
            {searchInput && <div>{searchInput}</div>}
            {children}
          </>
        )}
      </CardContent>
    </Card>
  );
}
