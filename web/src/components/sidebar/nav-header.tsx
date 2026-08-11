"use client";

import { PlusIcon, SidebarIcon } from "lucide-react";

import { BreadcrumbNav } from "@/components/breadcrumb";
import { openQuickCreate } from "@/components/quick-create/open";
import { NavCommandTrigger } from "@/components/sidebar/nav-command-trigger";
import { TodoSheetTrigger } from "@/components/todo/todo-sheet-trigger";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { useSidebar } from "@/components/ui/sidebar";

export function NavHeader() {
  const { toggleSidebar } = useSidebar();

  return (
    <header className="bg-surface-raised/95 sticky top-0 z-30 flex h-14 w-full shrink-0 items-center border-b backdrop-blur-sm">
      <div className="flex min-w-0 flex-1 items-center gap-2 px-3 sm:px-4">
        <Button
          className="size-8"
          variant="ghost"
          size="icon"
          onClick={toggleSidebar}
        >
          <SidebarIcon />
          <span className="sr-only">Toggle sidebar</span>
        </Button>
        <Separator orientation="vertical" className="h-5" />
        <BreadcrumbNav className="min-w-0 flex-1" />
        <div className="ml-auto flex items-center">
          <NavCommandTrigger className="w-44 lg:w-56" />
        </div>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => openQuickCreate()}
          aria-label="Quick create"
          title="Quick create (C)"
        >
          <PlusIcon />
        </Button>
        <TodoSheetTrigger />
      </div>
    </header>
  );
}
