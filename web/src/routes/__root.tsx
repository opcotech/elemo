/// <reference types="vite/client" />
import {
  HeadContent,
  Outlet,
  Scripts,
  createRootRouteWithContext,
} from "@tanstack/react-router";
import { lazy } from "react";
import type { ReactNode } from "react";

import { NotFound } from "@/components/shared/not-found";
import { ThemeProvider } from "@/components/shared/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import { NavigationProgress } from "@/components/ui/top-progress-bar";
import { PLUGIN_IMPORT_MAP } from "@/lib/plugins/runtime";
import type { Theme } from "@/lib/theme";
import type { RouterContext } from "@/router";
import appCss from "@/styles/app.css?url";

const ReactQueryDevtools = lazy(async () => {
  const module = await import("@tanstack/react-query-devtools");
  return { default: module.ReactQueryDevtools };
});

export const Route = createRootRouteWithContext<RouterContext>()({
  beforeLoad: async () => {
    const { getThemeFn } = await import("@/lib/theme");
    return {
      theme: await getThemeFn(),
    };
  },
  head: () => ({
    meta: [
      {
        charSet: "utf-8",
      },
      {
        name: "viewport",
        content: "width=device-width, initial-scale=1",
      },
      {
        title: "Elemo - The next generation project management platform",
      },
    ],
    links: [
      {
        rel: "stylesheet",
        href: appCss,
      },
    ],
  }),
  component: RootComponent,
  notFoundComponent: NotFound,
});

function RootComponent() {
  const { theme } = Route.useRouteContext();

  return (
    <RootDocument theme={theme}>
      <ThemeProvider initialTheme={theme}>
        <NavigationProgress />
        <Outlet />
        <Toaster position="top-center" duration={3000} richColors />
      </ThemeProvider>
      {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
    </RootDocument>
  );
}

function RootDocument({
  children,
  theme,
}: Readonly<{ children: ReactNode; theme: Theme }>) {
  // Keep the bootstrap script static and read preference from the DOM so
  // theme is not interpolated into executable JavaScript.
  const themeBootstrapScript = `(() => {
  const preference = document.documentElement.dataset.themePreference;
  const resolved =
    preference === "system" &&
    window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : preference === "dark"
        ? "dark"
        : "light";
  document.documentElement.classList.remove("light", "dark");
  document.documentElement.classList.add(resolved);
})();`;

  return (
    <html
      className={theme === "system" ? undefined : theme}
      data-theme-preference={theme}
      suppressHydrationWarning
    >
      <head>
        <script
          type="importmap"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify(PLUGIN_IMPORT_MAP),
          }}
        />
        <HeadContent />
        <script dangerouslySetInnerHTML={{ __html: themeBootstrapScript }} />
      </head>
      <body>
        {children}
        <Scripts />
      </body>
    </html>
  );
}
