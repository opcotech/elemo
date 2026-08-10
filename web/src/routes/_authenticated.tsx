import { Outlet, createFileRoute, redirect } from "@tanstack/react-router";

import { AuthProvider } from "@/components/auth/auth-provider";
import { AuthenticatedLayout } from "@/components/layout/authenticated-layout";
import { sanitizeRedirectTarget } from "@/lib/auth/redirect";
import { prefetchAuthenticatedChrome } from "@/lib/route-data";

export const Route = createFileRoute("/_authenticated")({
  beforeLoad: async ({ context, location }) => {
    const { currentSessionFn } = await import("@/lib/auth/functions");
    const session = await currentSessionFn();

    if (!session) {
      throw redirect({
        to: "/login",
        search: {
          redirect: sanitizeRedirectTarget(location.href),
        },
      });
    }

    prefetchAuthenticatedChrome(context.queryClient);
    return { session };
  },
  component: AuthenticatedRoute,
});

function AuthenticatedRoute() {
  const { session } = Route.useRouteContext();

  return (
    <AuthProvider initialSession={session}>
      <AuthenticatedLayout>
        <Outlet />
      </AuthenticatedLayout>
    </AuthProvider>
  );
}
