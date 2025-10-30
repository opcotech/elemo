import { Link, useRouter } from "@tanstack/react-router";
import { Building2, Shield, User } from "lucide-react";
import React, { useEffect, useState } from "react";
import type { ReactNode } from "react";

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
} from "@/components/ui/sidebar";

interface SettingsLayoutProps {
  children: ReactNode;
}

interface SettingsNavigationItem {
  label: string;
  href: string;
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
  const router = useRouter();
  const [currentPath, setCurrentPath] = useState(
    router.state.location.pathname
  );

  useEffect(() => {
    const unsub = router.subscribe("onResolved", () => {
      setCurrentPath(router.state.location.pathname);
    });
    return unsub;
  }, [router]);

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
                      asChild
                      isActive={isActive}
                      className="h-auto bg-transparent"
                    >
                      <Link to={item.href} className="flex items-start gap-2">
                        <item.icon className="mt-0.5 size-4" />
                        <div className="flex flex-col items-start">
                          <span className="font-medium">{item.label}</span>
                          {item.description && (
                            <span className="text-xs text-gray-500">
                              {item.description}
                            </span>
                          )}
                        </div>
                      </Link>
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
  return (
    <SidebarProvider>
      <Sidebar variant="sidebar" collapsible="none" className="bg-transparent">
        <SidebarContent className="pt-6">
          <SettingsSidebar />
        </SidebarContent>
      </Sidebar>

      <main className="flex-1 overflow-auto px-12 py-8">{children}</main>
    </SidebarProvider>
  );
}
