import { Trash2 } from "lucide-react";
import type { ReactNode } from "react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

interface DeleteConfirmationDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string | ReactNode;
  consequences?: string[];
  deleteButtonIcon?: React.ElementType;
  deleteButtonText?: string;
  onConfirm: () => void;
  isPending?: boolean;
  children?: ReactNode;
}

export function DeleteConfirmationDialog({
  open,
  onOpenChange,
  title,
  description,
  consequences,
  deleteButtonIcon = Trash2,
  deleteButtonText = "Delete",
  onConfirm,
  isPending = false,
  children,
}: DeleteConfirmationDialogProps) {
  const DeleteButtonIcon = deleteButtonIcon;

  const handleConfirm = () => {
    onConfirm();
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription className="space-y-2">
            {typeof description === "string" ? (
              <p>{description}</p>
            ) : (
              description
            )}
            {children}
            {consequences && consequences.length > 0 && (
              <>
                <p className="font-medium">What will happen:</p>
                <ul className="list-inside list-disc space-y-1 text-sm">
                  {consequences.map((consequence, index) => (
                    <li key={index}>{consequence}</li>
                  ))}
                </ul>
              </>
            )}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            onClick={handleConfirm}
            disabled={isPending}
          >
            {isPending ? (
              <>
                <span>Deleting...</span>
              </>
            ) : (
              <>
                <DeleteButtonIcon className="size-4" />
                {deleteButtonText}
              </>
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
