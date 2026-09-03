import { Link, useRouterState } from "@tanstack/react-router";
import {
  ArrowLeft,
  Building2,
  Folder,
  Puzzle,
  Shield,
  User,
} from "lucide-react";
import React from "react";
import type { ReactNode } from "react";

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
} from "@/components/ui/sidebar";
import { cn } from "@/lib/utils";

interface SettingsLayoutProps {
  children: ReactNode;
}

interface SettingsNavigationItem {
  label: string;
  href:
    | "/settings"
    | "/settings/organizations"
    | "/settings/namespaces"
    | "/settings/security"
    | "/settings/plugins";
  icon: React.ElementType;
  description?: string;
}

interface SettingsNavigationGroup {
  group: string;
  items: SettingsNavigationItem[];
}

export const settingsNavigation: SettingsNavigationGroup[] = [
  {
    group: "General",
    items: [
      {
        label: "Profile & Account",
        href: "/settings",
        icon: User,
        description: "Manage your personal information",
      },
      {
        label: "Organizations",
        href: "/settings/organizations",
        icon: Building2,
        description: "View and manage organizations",
      },
      {
        label: "Namespaces",
        href: "/settings/namespaces",
        icon: Folder,
        description: "View and manage namespaces",
      },
      {
        label: "Plugins",
        href: "/settings/plugins",
        icon: Puzzle,
        description: "Install and manage instance plugins",
      },
    ],
  },
  {
    group: "Security",
    items: [
      {
        label: "Password & Authentication",
        href: "/settings/security",
        icon: Shield,
        description: "Manage your password and authentication settings",
      },
    ],
  },
];

export function SettingsSidebar() {
  const currentPath = useRouterState({
    select: (state) => state.location.pathname,
  });

  return (
    <>
      {settingsNavigation.map((group) => (
        <SidebarGroup key={group.group}>
          <SidebarGroupLabel>{group.group}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {group.items.map((item) => {
                const isActive = currentPath === item.href;
                return (
                  <SidebarMenuItem key={item.href}>
                    <SidebarMenuButton
                      render={
                        <Link
                          to={item.href}
                          className="flex items-start gap-2"
                        />
                      }
                      isActive={isActive}
                      className="h-auto bg-transparent"
                    >
                      <item.icon className="mt-0.5 size-4" />
                      <div className="flex flex-col items-start">
                        <span className="font-medium">{item.label}</span>
                        {item.description && (
                          <span className="text-muted-foreground text-xs">
                            {item.description}
                          </span>
                        )}
                      </div>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      ))}
    </>
  );
}

export function SettingsLayout({ children }: SettingsLayoutProps) {
  const currentPath = useRouterState({
    select: (state) => state.location.pathname,
  });

  return (
    <SidebarProvider>
      <div className="flex h-full w-full flex-col">
        <header className="flex h-14 shrink-0 items-center border-b px-4 sm:px-6">
          <Link
            to="/"
            className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 text-sm font-medium"
          >
            <ArrowLeft className="size-4" />
            Back to Home
          </Link>
          <span className="ml-auto text-sm font-semibold">Settings</span>
        </header>
        <nav
          aria-label="Settings"
          className="flex gap-1 overflow-x-auto border-b p-2 md:hidden"
        >
          {settingsNavigation.flatMap((group) =>
            group.items.map((item) => {
              const isActive = currentPath === item.href;
              return (
                <Link
                  key={item.href}
                  to={item.href}
                  aria-current={isActive ? "page" : undefined}
                  className={cn(
                    "shrink-0 rounded-lg px-3 py-2 text-sm",
                    isActive
                      ? "bg-muted text-foreground font-medium"
                      : "text-muted-foreground hover:bg-muted"
                  )}
                >
                  {item.label}
                </Link>
              );
            })
          )}
        </nav>
        <div className="flex min-h-0 flex-1 justify-center overflow-hidden">
          <div className="flex min-h-0 w-full max-w-6xl">
            <Sidebar
              variant="sidebar"
              collapsible="none"
              className="w-72! shrink-0 border-0 bg-transparent max-md:hidden"
            >
              <SidebarContent className="pt-6">
                <SettingsSidebar />
              </SidebarContent>
            </Sidebar>

            <SidebarInset className="overflow-auto">
              <div className="w-full px-4 py-6 sm:px-6 md:px-8 md:py-8">
                {children}
              </div>
            </SidebarInset>
          </div>
        </div>
      </div>
    </SidebarProvider>
  );
}
