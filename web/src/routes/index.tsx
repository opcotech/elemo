import { createFileRoute, redirect } from "@tanstack/react-router";

import { redirectIfAuthenticated } from "@/lib/auth/require-auth";

export const Route = createFileRoute("/")({
  beforeLoad: () => {
    // First check if user is authenticated and redirect to dashboard
    redirectIfAuthenticated();

    // If not authenticated, redirect to login
    throw redirect({
      to: "/login",
      search: {
        redirect: undefined,
      },
    });
  },
  component: () => null, // This will never render due to redirects
});
