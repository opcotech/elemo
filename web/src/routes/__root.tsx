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
import { TopProgressBar } from "@/components/ui/top-progress-bar";
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
        <TopProgressBar />
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
  const themeScript = `(() => { const preference = ${JSON.stringify(
    theme
  )}; const resolved = preference === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : preference === "system" ? "light" : preference; document.documentElement.classList.remove("light", "dark"); document.documentElement.classList.add(resolved); })();`;

  return (
    <html
      className={theme === "system" ? undefined : theme}
      suppressHydrationWarning
    >
      <head>
        <HeadContent />
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
      </head>
      <body>
        {children}
        <Scripts />
      </body>
    </html>
  );
}
