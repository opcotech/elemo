import type { ErrorComponentProps } from "@tanstack/react-router";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

export function DefaultPendingComponent() {
  return (
    <div className="flex min-h-48 items-center justify-center" aria-busy="true">
      <Spinner />
      <span className="sr-only">Loading page</span>
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
