import { Link, useRouter } from "@tanstack/react-router";

import type { NavigationConfig } from "./types";

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";

interface ProjectNavigationProps {
  navigationConfig: NavigationConfig;
}

export function ProjectNavigation({
  navigationConfig,
}: ProjectNavigationProps) {
  const router = useRouter();
  const currentPath = router.state.location.pathname;

  return (
    <SidebarGroup>
      <SidebarGroupLabel>{navigationConfig.label}</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {navigationConfig.items.map((item) => {
            const isActive = currentPath === item.href;

            return (
              <SidebarMenuItem key={item.href}>
                <SidebarMenuButton asChild isActive={isActive}>
                  <Link to={item.href}>
                    <item.icon className="size-4" />
                    <span>{item.label}</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            );
          })}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
