import { Link } from "@tanstack/react-router";
import { Home, Inbox } from "lucide-react";
import { useEffect, useState } from "react";
import type { ReactNode } from "react";

import { CommandPalette } from "@/components/command-palette/command-palette";
import { ContextualNavigationSection } from "@/components/sidebar/contextual-navigation-section";
import { GlobalContextSection } from "@/components/sidebar/global-context-section";
import { NavHeader } from "@/components/sidebar/nav-header";
import { NavUser, NavUserSkeleton } from "@/components/sidebar/nav-user";
import { TodoSheet } from "@/components/todo/todo-sheet";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
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
import { useNavigationContext } from "@/hooks/use-navigation-context";

interface AuthenticatedLayoutProps {
  children: ReactNode;
  sidebarContent?: ReactNode;
}

interface SidebarNavigationItem {
  label: string;
  href: string;
  icon: React.ElementType;
}

export function AuthenticatedLayout({
  children,
  sidebarContent,
}: AuthenticatedLayoutProps) {
  const { user } = useAuth();
  const navigationContext = useNavigationContext();
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);

  const navigation: SidebarNavigationItem[] = [
    { label: "Dashboard", href: "/dashboard", icon: Home },
    { label: "Inbox", href: "/inbox", icon: Inbox },
  ];

  // Add Command-K keyboard shortcut
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key === "k") {
        event.preventDefault();
        setCommandPaletteOpen((open) => !open);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  return (
    <AddTodoFormProvider>
      <EditTodoFormProvider>
        <TodoSheetProvider>
          <SidebarProvider>
            <div className="flex h-screen w-full overflow-hidden">
              <Sidebar variant="inset">
                <SidebarHeader>
                  <h2 className="px-2 py-2 text-lg font-semibold">Elemo</h2>
                </SidebarHeader>

                <SidebarContent>
                  {/* Global Navigation */}
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

                  {/* Custom sidebar content (for settings pages, etc.) */}
                  {sidebarContent}

                  {/* Global Context Section - Namespaces with projects */}
                  {navigationContext.type === "global" && (
                    <GlobalContextSection />
                  )}

                  {/* Contextual Navigation Section - Dynamic menu based on context */}
                  {navigationContext.type !== "global" && (
                    <ContextualNavigationSection />
                  )}
                </SidebarContent>

                <SidebarFooter>
                  <SidebarMenu>
                    {user ? <NavUser user={user} /> : <NavUserSkeleton />}
                  </SidebarMenu>
                </SidebarFooter>
              </Sidebar>

              <SidebarInset className="flex flex-col overflow-hidden shadow-sm">
                <NavHeader />
                <div className="flex-1 overflow-auto">{children}</div>
              </SidebarInset>
            </div>
          </SidebarProvider>

          <TodoSheet />

          {/* Global Command Palette */}
          <CommandPalette
            open={commandPaletteOpen}
            onOpenChange={setCommandPaletteOpen}
          />
        </TodoSheetProvider>
      </EditTodoFormProvider>
    </AddTodoFormProvider>
  );
}
