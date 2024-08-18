import { createFileRoute } from "@tanstack/react-router";

import { LoginForm } from "@/components/auth/login-form";
import { redirectIfAuthenticated } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/login")({
  beforeLoad: redirectIfAuthenticated,
  validateSearch: (search: Record<string, unknown>) => ({
    redirect: (search.redirect as string) || undefined,
  }),
  component: LoginPage,
});

function LoginPage() {
  const { redirect: target } = Route.useSearch();
  return <LoginForm redirectTo={target} />;
}
