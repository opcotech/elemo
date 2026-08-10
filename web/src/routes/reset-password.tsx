import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";

import { PasswordResetForm } from "@/components/auth/password-reset-form";

export const Route = createFileRoute("/reset-password")({
  validateSearch: z.object({
    token: z.string().min(1).max(4096),
  }),
  component: ResetPasswordPage,
});

function ResetPasswordPage() {
  return <PasswordResetForm />;
}
