import type { ErrorComponentProps } from "@tanstack/react-router";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

export function DefaultPendingComponent() {
  return (
    <div
      className="space-y-6 px-4 py-6 sm:px-6 lg:px-8 lg:py-8"
      role="status"
      aria-busy="true"
    >
      <span className="sr-only">Loading page</span>
      <div className="space-y-2">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-80 max-w-full" />
      </div>
      <div className="space-y-3">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>
    </div>
  );
}

export function DefaultErrorComponent({ reset }: ErrorComponentProps) {
  return (
    <main className="container mx-auto p-6">
      <Alert variant="destructive">
        <AlertTitle>Page unavailable</AlertTitle>
        <AlertDescription>
          <p>
            This page could not be loaded. Try again, or return to a context you
            can access.
          </p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="mt-3"
            onClick={reset}
          >
            Try again
          </Button>
        </AlertDescription>
      </Alert>
    </main>
  );
}
