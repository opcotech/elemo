import { createFileRoute } from "@tanstack/react-router";

import { PasswordResetForm } from "@/components/auth/password-reset-form";

export const Route = createFileRoute("/reset-password")({
  validateSearch: (search: Record<string, unknown>) => ({
    token: (search.token as string) || undefined,
  }),
  component: ResetPasswordPage,
});

function ResetPasswordPage() {
  return <PasswordResetForm />;
}
