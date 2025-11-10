import { useEffect, useRef } from "react";
import type { ReactNode } from "react";
import type { FieldValues, UseFormReturn } from "react-hook-form";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Form } from "@/components/ui/form";
import { Spinner } from "@/components/ui/spinner";

interface DialogFormProps<TFormValues extends FieldValues> {
  form: UseFormReturn<TFormValues>;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  onSubmit: (e: React.FormEvent<HTMLFormElement>) => void;
  isPending?: boolean;
  error?: Error | string | null;
  children: ReactNode;
  submitButtonText?: string;
  cancelButtonText?: string;
  onReset?: () => void;
  className?: string;
}

export function DialogForm<TFormValues extends FieldValues>({
  form,
  open,
  onOpenChange,
  title,
  onSubmit,
  isPending = false,
  error,
  children,
  submitButtonText = "Save",
  cancelButtonText = "Cancel",
  onReset,
  className,
}: DialogFormProps<TFormValues>) {
  const prevOpenRef = useRef(open);

  useEffect(() => {
    if (prevOpenRef.current && !open && onReset) {
      onReset();
    }
    prevOpenRef.current = open;
  }, [open]);

  const errorMessage =
    typeof error === "string" ? error : error?.message || null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className={className}>
        <Form {...form}>
          <form onSubmit={onSubmit} className="flex flex-col gap-y-6">
            <DialogHeader>
              <DialogTitle>{title}</DialogTitle>
            </DialogHeader>

            {errorMessage && (
              <Alert variant="destructive">
                <AlertTitle>Failed to save</AlertTitle>
                <AlertDescription>{errorMessage}</AlertDescription>
              </Alert>
            )}

            {children}

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={isPending}
              >
                {cancelButtonText}
              </Button>
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
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
