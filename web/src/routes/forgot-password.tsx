import { createFileRoute } from "@tanstack/react-router";

import { PasswordResetRequestForm } from "@/components/auth/password-reset-request-form";
import { safeRedirectSearchSchema } from "@/lib/auth/redirect";
import { redirectIfAuthenticated } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/forgot-password")({
  beforeLoad: redirectIfAuthenticated,
  validateSearch: safeRedirectSearchSchema,
  component: ForgotPasswordPage,
});

function ForgotPasswordPage() {
  return <PasswordResetRequestForm />;
}
