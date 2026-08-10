import { Link, useRouterState } from "@tanstack/react-router";
import {
  HomeIcon,
  InboxIcon,
  ListChecksIcon,
  SparklesIcon,
} from "lucide-react";
import type { ReactNode } from "react";
import { Suspense, lazy, useEffect, useLayoutEffect, useState } from "react";

import { NamespaceSwitcher } from "@/components/namespace-switcher";
import {
  QUICK_CREATE_EVENT,
  isQuickCreateType,
  isTypingTarget,
} from "@/components/quick-create/types";
import type { QuickCreateType } from "@/components/quick-create/types";
import { ContextualNavigationSection } from "@/components/sidebar/contextual-navigation-section";
import { NavHeader } from "@/components/sidebar/nav-header";
import { NavUser, NavUserSkeleton } from "@/components/sidebar/nav-user";
import {
  RecentDocuments,
  RecentProjects,
  RecentWorkItems,
} from "@/components/sidebar/recent-entities";
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
  SidebarSeparator,
} from "@/components/ui/sidebar";
import { useAuth } from "@/hooks/use-auth";
import { uiActions, useUiSelector } from "@/lib/ui-store";

const QuickCreate = lazy(() =>
  import("@/components/quick-create").then((module) => ({
    default: module.QuickCreate,
  }))
);

const TodoSheet = lazy(() =>
  import("@/components/todo/todo-sheet").then((module) => ({
    default: module.TodoSheet,
  }))
);

interface AuthenticatedLayoutProps {
  children: ReactNode;
}

interface QuickCreateMountState {
  ready: boolean;
  initialOpen: boolean;
  initialType: QuickCreateType;
}

export function AuthenticatedLayout({ children }: AuthenticatedLayoutProps) {
  const { user } = useAuth();
  const currentPath = useRouterState({
    select: (state) => state.location.pathname,
  });
  const todoSheetOpen = useUiSelector((state) => state.todoSheetOpen);
  const addTodoOpen = useUiSelector((state) => state.addTodoOpen);
  const [todoSheetReady, setTodoSheetReady] = useState(false);
  const [quickCreate, setQuickCreate] = useState<QuickCreateMountState>({
    ready: false,
    initialOpen: false,
    initialType: "todo",
  });

  const navigation = [
    { label: "Home", href: "/", icon: HomeIcon },
    { label: "My Work", href: "/my-work", icon: ListChecksIcon },
    { label: "Inbox", href: "/inbox", icon: InboxIcon },
  ] as const;

  useLayoutEffect(() => {
    uiActions.hydrateNavigation();
  }, []);

  useEffect(() => {
    if (todoSheetOpen || addTodoOpen) {
      setTodoSheetReady(true);
    }
  }, [todoSheetOpen, addTodoOpen]);

  useEffect(() => {
    const mountQuickCreate = (type: QuickCreateType = "todo") => {
      setQuickCreate((current) =>
        current.ready
          ? current
          : { ready: true, initialOpen: true, initialType: type }
      );
    };

    const openDialog = (event: Event) => {
      if (event instanceof CustomEvent && isQuickCreateType(event.detail)) {
        mountQuickCreate(event.detail);
      } else {
        mountQuickCreate("todo");
      }
    };

    const handleKeyDown = (event: KeyboardEvent) => {
      if (
        event.key.toLowerCase() === "c" &&
        !event.metaKey &&
        !event.ctrlKey &&
        !event.altKey &&
        !isTypingTarget(event.target)
      ) {
        mountQuickCreate("todo");
      }
    };

    window.addEventListener(QUICK_CREATE_EVENT, openDialog);
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener(QUICK_CREATE_EVENT, openDialog);
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  return (
    <SidebarProvider>
      <div className="flex h-svh w-full overflow-hidden">
        <Sidebar variant="sidebar" collapsible="icon">
          <SidebarHeader className="gap-3 p-3 group-data-[collapsible=icon]:items-center group-data-[collapsible=icon]:px-1.5">
            <Link
              to="/"
              className="focus-visible:ring-ring flex h-8 items-center gap-2 rounded-lg px-1.5 outline-none group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:px-0 focus-visible:ring-2"
            >
              <span className="bg-primary text-primary-foreground flex size-7 shrink-0 items-center justify-center rounded-lg">
                <SparklesIcon className="size-4" />
              </span>
              <span className="text-base font-semibold tracking-tight group-data-[collapsible=icon]:hidden">
                Elemo
              </span>
            </Link>
            <NamespaceSwitcher />
          </SidebarHeader>

          <SidebarContent>
            <SidebarGroup>
              <SidebarGroupContent>
                <SidebarMenu>
                  {navigation.map((item) => (
                    <SidebarMenuItem key={item.href}>
                      <SidebarMenuButton
                        render={<Link to={item.href} />}
                        tooltip={item.label}
                        isActive={currentPath === item.href}
                      >
                        <item.icon />
                        <span>{item.label}</span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  ))}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>

            <ContextualNavigationSection />
            <RecentWorkItems />
            <RecentDocuments />
            <RecentProjects />
          </SidebarContent>

          <SidebarSeparator />
          <SidebarFooter className="gap-1">
            <SidebarMenu>
              {user ? <NavUser user={user} /> : <NavUserSkeleton />}
            </SidebarMenu>
          </SidebarFooter>
        </Sidebar>

        <SidebarInset className="min-w-0 overflow-hidden">
          <NavHeader />
          <div className="min-h-0 flex-1 overflow-auto">{children}</div>
        </SidebarInset>
      </div>
      {todoSheetReady ? (
        <Suspense fallback={null}>
          <TodoSheet />
        </Suspense>
      ) : null}
      {quickCreate.ready ? (
        <Suspense fallback={null}>
          <QuickCreate
            initialOpen={quickCreate.initialOpen}
            initialType={quickCreate.initialType}
          />
        </Suspense>
      ) : null}
    </SidebarProvider>
  );
}
