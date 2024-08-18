"use client";

import { SidebarIcon } from "lucide-react";

import { BreadcrumbNav } from "@/components/breadcrumb";
import { NavCommandTrigger } from "@/components/sidebar/nav-command-trigger";
import { TodoSheetTrigger } from "@/components/todo/todo-sheet";
import { Button } from "@/components/ui/button";
import { useSidebar } from "@/components/ui/sidebar";

export function NavHeader() {
  const { toggleSidebar } = useSidebar();

  return (
    <header className="sticky top-0 z-50 flex w-full items-center border-b">
      <div className="bg-card flex w-full items-center gap-2 rounded-t-xl px-2 py-1.5">
        <Button
          className="size-8"
          variant="ghost"
          size="icon"
          onClick={toggleSidebar}
        >
          <SidebarIcon />
        </Button>
        <BreadcrumbNav className="hidden sm:block" />
        <div className="flex w-full items-center gap-2 sm:ml-auto sm:w-auto">
          <NavCommandTrigger />
        </div>
        <TodoSheetTrigger />
      </div>
    </header>
  );
}
