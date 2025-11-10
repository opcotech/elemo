import type { ReactNode } from "react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";

interface FormCardProps {
  title?: string;
  description?: string;
  onSubmit: (e: React.FormEvent<HTMLFormElement>) => void;
  onCancel?: () => void;
  isPending?: boolean;
  error?: Error | string | null;
  children: ReactNode;
  submitButtonText?: string;
  cancelButtonText?: string;
  className?: string;
}

export function FormCard({
  title,
  description,
  onSubmit,
  onCancel,
  isPending = false,
  error,
  children,
  submitButtonText = "Save Changes",
  cancelButtonText = "Cancel",
  className,
}: FormCardProps) {
  const errorMessage =
    typeof error === "string" ? error : error?.message || null;

  return (
    <Card className={className}>
      {(title || description) && (
        <CardHeader>
          {title && <CardTitle>{title}</CardTitle>}
          {description && <CardDescription>{description}</CardDescription>}
        </CardHeader>
      )}
      <CardContent>
        <form onSubmit={onSubmit} className="flex flex-col gap-y-6">
          {errorMessage && (
            <Alert variant="destructive">
              <AlertTitle>Failed to save</AlertTitle>
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
          )}

          {children}

          <div className="flex justify-end gap-2">
            {onCancel && (
              <Button
                type="button"
                variant="outline"
                onClick={onCancel}
                disabled={isPending}
              >
                {cancelButtonText}
              </Button>
            )}
            <Button type="submit" disabled={isPending}>
              {isPending ? (
                <>
                  <Spinner size="xs" className="mr-0.5 text-white" />
                  <span>Saving...</span>
                </>
              ) : (
                submitButtonText
              )}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
