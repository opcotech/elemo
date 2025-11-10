import { createRouter as createTanStackRouter } from "@tanstack/react-router";

import { routeTree } from "./routeTree.gen";

export function createRouter(options?: { context?: Record<string, unknown> }) {
  const router = createTanStackRouter({
    routeTree,
    scrollRestoration: false,
    context: options?.context as any,
  });

  router.subscribe("onResolved", () => {
    requestAnimationFrame(() => {
      window.scrollTo({ top: 0, behavior: "smooth" });

      const scrollableSelectors = [
        "[data-radix-sidebar-inset][class*='overflow-auto']",
        "[data-radix-scroll-area-viewport]",
        "div[class*='flex-1'][class*='overflow-auto']",
        "div[class*='flex-1'][class*='overflow-y-auto']",
      ];

      scrollableSelectors.forEach((selector) => {
        try {
          const elements = document.querySelectorAll(selector);
          elements.forEach((element) => {
            if (element instanceof HTMLElement) {
              element.scrollTo({ top: 0, behavior: "smooth" });
            }
          });
        } catch {
          // Ignore invalid selectors
        }
      });
    });
  });

  return router;
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof createRouter>;
  }
}
