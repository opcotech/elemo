import { showErrorToast, showSuccessToast } from "@elemo/plugin-ui";

export function notifySuccess(title: string, description: string): void {
  showSuccessToast(title, description);
}

export function errorMessage(cause: unknown): string {
  if (cause instanceof Error && cause.message) {
    return cause.message;
  }
  if (typeof cause === "string" && cause) {
    return cause;
  }
  if (cause && typeof cause === "object" && "message" in cause) {
    const message = (cause as { message: unknown }).message;
    if (typeof message === "string" && message) {
      return message;
    }
  }
  return "Unknown error";
}

export function notifyError(title: string, cause: unknown): void {
  const description = errorMessage(cause);
  console.error(title, cause);
  try {
    showErrorToast(title, description);
  } catch (error) {
    console.error("Failed to show error toast", error);
  }
}
