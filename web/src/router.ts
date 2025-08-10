import { createRouter as createTanStackRouter } from "@tanstack/react-router";

import { routeTree } from "./routeTree.gen";

export function createRouter(options?: { context?: Record<string, unknown> }) {
  const router = createTanStackRouter({
    routeTree,
    scrollRestoration: true,
    context: options?.context as any,
  });

  return router;
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof createRouter>;
  }
}
