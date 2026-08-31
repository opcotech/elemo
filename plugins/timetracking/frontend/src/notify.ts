import { showErrorToast, showSuccessToast } from "@elemo/plugin-ui";

export function notifySuccess(title: string, description: string): void {
  showSuccessToast(title, description);
}

export function notifyError(title: string, cause: unknown): void {
  const description =
    cause instanceof Error
      ? cause.message
      : typeof cause === "string"
        ? cause
        : "Unknown error";
  console.error(title, cause);
  try {
    showErrorToast(title, description);
  } catch (error) {
    console.error("Failed to show error toast", error);
  }
}
