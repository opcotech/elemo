import { toast } from "sonner";

export function showToast(
  title: string,
  description: string,
  action: { label: string; onClick: () => void }
) {
  toast.info(title, { description, action });
}

export function showSuccessToast(
  title: string,
  description: string,
  action?: { label: string; onClick: () => void }
) {
  toast.success(title, { description, action });
}

export function showErrorToast(title: string, error: Error | string) {
  toast.error(title, {
    description: typeof error === "string" ? error : error.message,
  });
}
