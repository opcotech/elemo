import { createFileRoute } from "@tanstack/react-router";

import { PasswordResetRequestForm } from "@/components/auth/password-reset-request-form";
import { redirectIfAuthenticated } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/forgot-password")({
  beforeLoad: redirectIfAuthenticated,
  validateSearch: (search: Record<string, unknown>) => ({
    redirect: (search.redirect as string) || undefined,
  }),
  component: ForgotPasswordPage,
});

function ForgotPasswordPage() {
  return <PasswordResetRequestForm />;
}
