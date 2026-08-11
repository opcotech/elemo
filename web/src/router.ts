import type { QueryClient } from "@tanstack/react-query";
import { createRouter as createTanStackRouter } from "@tanstack/react-router";
import { setupRouterSsrQueryIntegration } from "@tanstack/react-router-ssr-query";

import {
  DefaultErrorComponent,
  DefaultPendingComponent,
} from "./components/router/default-boundaries";
import type { RouteBreadcrumb } from "./lib/breadcrumb";
import { createQueryClient } from "./lib/query-client";
import { routeTree } from "./routeTree.gen";

export interface RouterContext {
  queryClient: QueryClient;
}

export function getRouter() {
  const queryClient = createQueryClient();
  const router = createTanStackRouter({
    routeTree,
    context: {
      queryClient,
    },
    defaultPreload: "intent",
    // With React Query, always re-run loaders on preload/visit; freshness is
    // controlled by Query staleTime / fetchQuery, not the router.
    defaultPreloadStaleTime: 0,
    scrollRestoration: true,
    defaultPendingComponent: DefaultPendingComponent,
    defaultErrorComponent: DefaultErrorComponent,
  });

  setupRouterSsrQueryIntegration({
    router,
    queryClient,
  });

  return router;
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof getRouter>;
  }

  interface StaticDataRouteOption {
    breadcrumb?: RouteBreadcrumb | ((loaderData: unknown) => RouteBreadcrumb);
  }
}
