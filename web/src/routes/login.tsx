import { createFileRoute } from "@tanstack/react-router";

import { AuthProvider } from "@/components/auth/auth-provider";
import { LoginForm } from "@/components/auth/login-form";
import { safeRedirectSearchSchema } from "@/lib/auth/redirect";
import { redirectIfAuthenticated } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/login")({
  beforeLoad: redirectIfAuthenticated,
  validateSearch: safeRedirectSearchSchema,
  component: LoginPage,
});

function LoginPage() {
  const { redirect: target } = Route.useSearch();
  return (
    <AuthProvider>
      <LoginForm redirectTo={target} />
    </AuthProvider>
  );
}
