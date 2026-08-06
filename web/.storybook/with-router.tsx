import type { Decorator } from "@storybook/react-vite";
import {
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";

/**
 * Provides a minimal TanStack Router so components using `Link` can render
 * outside the app (e.g. Storybook).
 */
export const withRouter: Decorator = (Story) => {
  const rootRoute = createRootRoute({
    component: () => <Story />,
  });

  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "/",
  });

  // Absorb arbitrary breadcrumb / story hrefs without route-matching errors.
  const splatRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: "$",
  });

  const router = createRouter({
    routeTree: rootRoute.addChildren([indexRoute, splatRoute]),
    history: createMemoryHistory({ initialEntries: ["/"] }),
  });

  return <RouterProvider router={router} />;
};
