/// <reference types="vite/client" />
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import {
  HeadContent,
  Outlet,
  Scripts,
  createRootRoute,
} from "@tanstack/react-router";
import type { ReactNode } from "react";

import { AuthDebug } from "@/components/auth/auth-debug";
import { AuthProvider } from "@/components/auth/auth-provider";
import { BreadcrumbProvider } from "@/components/breadcrumb";
import { NotFound } from "@/components/not-found";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import { TopProgressBar } from "@/components/ui/top-progress-bar";
import { config } from "@/config";
import { queryClient } from "@/lib/query-client";
import appCss from "@/styles/app.css?url";

export const Route = createRootRoute({
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
  return (
    <RootDocument>
      <QueryClientProvider client={queryClient}>
        <ThemeProvider>
          <BreadcrumbProvider>
            <AuthProvider>
              {config.auth().debugBanner && <AuthDebug />}
              <TopProgressBar />
              <Outlet />
              <Toaster position="top-center" duration={3000} richColors />
            </AuthProvider>
          </BreadcrumbProvider>
        </ThemeProvider>
        <ReactQueryDevtools initialIsOpen={false} />
      </QueryClientProvider>
    </RootDocument>
  );
}

function RootDocument({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <html>
      <head>
        <HeadContent />
      </head>
      <body>
        {children}
        <Scripts />
      </body>
    </html>
  );
}
