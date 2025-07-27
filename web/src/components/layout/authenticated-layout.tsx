import { Link } from "@tanstack/react-router";
import { Home, Inbox } from "lucide-react";
import type { ReactNode } from "react";

import { NavHeader } from "@/components/sidebar/nav-header";
import { NavUser, NavUserSkeleton } from "@/components/sidebar/nav-user";
import { TodoSheet } from "@/components/todo/todo-sheet";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
} from "@/components/ui/sidebar";
import { AddTodoFormProvider } from "@/contexts/add-todo-form-context";
import { EditTodoFormProvider } from "@/contexts/edit-todo-form-context";
import { TodoSheetProvider } from "@/contexts/todo-sheet-context";
import { useAuth } from "@/hooks/use-auth";

interface AuthenticatedLayoutProps {
  children: ReactNode;
}

interface SidebarNavigationItem {
  label: string;
  href: string;
  icon: React.ElementType;
}

export function AuthenticatedLayout({ children }: AuthenticatedLayoutProps) {
  const { user } = useAuth();

  const navigation: SidebarNavigationItem[] = [
    { label: "Dashboard", href: "/dashboard", icon: Home },
    { label: "Inbox", href: "/inbox", icon: Inbox },
  ];

  const workspaceNavigation: SidebarNavigationItem[] = [];

  const projectsNavigation: SidebarNavigationItem[] = [];

  return (
    <AddTodoFormProvider>
      <EditTodoFormProvider>
        <TodoSheetProvider>
          <SidebarProvider>
            <div className="flex h-screen w-screen">
              <Sidebar variant="inset">
                <SidebarHeader>
                  <h2 className="px-2 py-2 text-lg font-semibold">Elemo</h2>
                </SidebarHeader>
                <SidebarContent>
                  <SidebarGroup>
                    <SidebarGroupContent>
                      <SidebarMenu>
                        {navigation.map((item) => (
                          <SidebarMenuItem key={item.href}>
                            <SidebarMenuButton asChild>
                              <Link to={item.href}>
                                <item.icon className="h-4 w-4" />
                                <span>{item.label}</span>
                              </Link>
                            </SidebarMenuButton>
                          </SidebarMenuItem>
                        ))}
                      </SidebarMenu>
                    </SidebarGroupContent>
                  </SidebarGroup>
                  <SidebarGroup>
                    <SidebarGroupLabel>Workspace</SidebarGroupLabel>
                    <SidebarGroupContent>
                      <SidebarMenu>
                        {workspaceNavigation.map((item) => (
                          <SidebarMenuItem key={item.href}>
                            <SidebarMenuButton asChild>
                              <Link to={item.href}>
                                <item.icon className="h-4 w-4" />
                                <span>{item.label}</span>
                              </Link>
                            </SidebarMenuButton>
                          </SidebarMenuItem>
                        ))}
                      </SidebarMenu>
                    </SidebarGroupContent>
                  </SidebarGroup>
                  <SidebarGroup>
                    <SidebarGroupLabel>Projects</SidebarGroupLabel>
                    <SidebarGroupContent>
                      {projectsNavigation.map((item) => (
                        <SidebarMenuItem key={item.href}>
                          <SidebarMenuButton asChild>
                            <Link to={item.href}>
                              <item.icon className="h-4 w-4" />
                              <span>{item.label}</span>
                            </Link>
                          </SidebarMenuButton>
                        </SidebarMenuItem>
                      ))}
                    </SidebarGroupContent>
                  </SidebarGroup>
                </SidebarContent>
                <SidebarFooter>
                  <SidebarMenu>
                    {user ? <NavUser user={user} /> : <NavUserSkeleton />}
                  </SidebarMenu>
                </SidebarFooter>
              </Sidebar>
              <SidebarInset className="flex flex-col border">
                <NavHeader />
                <main className="flex-1 overflow-hidden">{children}</main>
              </SidebarInset>
            </div>
          </SidebarProvider>

          <TodoSheet />
        </TodoSheetProvider>
      </EditTodoFormProvider>
    </AddTodoFormProvider>
  );
}
