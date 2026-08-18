import { useNavigate } from "@tanstack/react-router";
import { SearchIcon } from "lucide-react";
import { Suspense, lazy, useEffect, useMemo, useState } from "react";

import { CommandTrigger } from "@/components/command-palette/command-trigger";
import { openQuickCreate } from "@/components/quick-create/open";
import { Button } from "@/components/ui/button";
import { useCommandPalette } from "@/hooks/use-command-palette";
import type { Command } from "@/lib/commands/registry";
import { internalPath } from "@/lib/internal-url";
import type { GlobalSearchEntry } from "@/lib/mock-data/types";
import { useUiSelector } from "@/lib/ui-store";
import { cn } from "@/lib/utils";

const CommandPalette = lazy(() =>
  import("@/components/command-palette/command-palette").then((module) => ({
    default: module.CommandPalette,
  }))
);

export function NavCommandTrigger({ className }: { className?: string }) {
  const navigate = useNavigate();
  const recentEntities = useUiSelector((state) => state.recentEntities);
  const [mockEntries, setMockEntries] = useState<
    readonly GlobalSearchEntry[] | null
  >(null);
  const {
    open,
    setOpen,
    handleAddTodo,
    handleShowTodos,
    handleToggleTheme,
    handleSetLightTheme,
    handleSetDarkTheme,
    handleSetSystemTheme,
    handleCreateOrganization,
    handleShowOrganizations,
    handleShowDocuments,
    handleGoToOrganization,
    handleCreateNamespace,
    handleShowNamespaces,
    handleGoToNamespace,
    handleCreateProject,
    handleShowProjects,
    handleGoToProject,
    canCreateOrganization,
    canCreateNamespace,
    canCreateProject,
    hasOrganization,
    hasNamespace,
    hasProject,
  } = useCommandPalette();

  useEffect(() => {
    if (!open || mockEntries !== null) return;

    let active = true;
    void import("@/lib/mock-data/command-fixtures").then(
      ({ mockCommandSearchEntries }) => {
        if (active) setMockEntries(mockCommandSearchEntries);
      },
      () => {
        if (active) setMockEntries([]);
      }
    );

    return () => {
      active = false;
    };
  }, [mockEntries, open]);

  const commands = useMemo<Command[]>(
    () => [
      ...recentEntities.map((entity) => ({
        id: `recent-${entity.type}-${entity.id}`,
        title: entity.label,
        description: `Recent ${entity.type}`,
        category: "recent",
        keywords: [entity.type],
        action: () => void navigate({ to: entity.href as never }),
      })),
      ...(mockEntries ?? []).map((entity) => ({
        id: `entity-${entity.kind}-${entity.id}`,
        title: entity.title,
        description: entity.subtitle,
        category: "entities",
        keywords: [...entity.keywords, entity.kind],
        action: () => {
          const href =
            entity.kind === "saved-view"
              ? `/my-work?view=${entity.id}`
              : entity.href;
          void navigate({ to: internalPath(href) as never });
        },
      })),
      {
        id: "open-home",
        title: "Open Home",
        category: "navigation",
        action: () => void navigate({ to: "/" }),
      },
      {
        id: "open-my-work",
        title: "Open My Work",
        category: "navigation",
        action: () =>
          void navigate({
            to: "/my-work",
            search: {
              display: "comfortable",
              group: "status",
              layout: "list",
              sort: "rank:asc",
            },
          }),
      },
      {
        id: "open-documents",
        title: "Open Documents",
        keywords: ["library", "folders"],
        category: "navigation",
        action: handleShowDocuments,
      },
      {
        id: "search-all",
        title: "Search all entities",
        keywords: ["find", "global"],
        category: "navigation",
        action: () =>
          void navigate({
            to: "/search",
            search: {
              page: 1,
              q: "",
              scope: "global",
              type: "all",
            },
          }),
      },
      {
        id: "quick-create",
        title: "Quick create",
        description: "Inherit the current namespace and project",
        keywords: ["work", "document", "todo"],
        category: "quick-actions",
        shortcut: ["C"],
        action: openQuickCreate,
      },
      {
        id: "add-todo",
        title: "Add Todo",
        keywords: ["create", "task"],
        category: "quick-actions",
        shortcut: ["shift", "t", "n"],
        action: handleAddTodo,
      },
      {
        id: "show-todos",
        title: "Show Todos",
        category: "quick-actions",
        shortcut: ["shift", "t", "s"],
        action: handleShowTodos,
      },
      {
        id: "toggle-theme",
        title: "Toggle Theme",
        category: "appearance",
        action: handleToggleTheme,
      },
      {
        id: "light-theme",
        title: "Light Theme",
        category: "appearance",
        action: handleSetLightTheme,
      },
      {
        id: "dark-theme",
        title: "Dark Theme",
        category: "appearance",
        action: handleSetDarkTheme,
      },
      {
        id: "system-theme",
        title: "System Theme",
        category: "appearance",
        action: handleSetSystemTheme,
      },
      {
        id: "create-organization",
        title: "Create Organization",
        category: "organizations",
        hidden: !canCreateOrganization,
        action: handleCreateOrganization,
      },
      {
        id: "show-organizations",
        title: "Show Organizations",
        category: "organizations",
        action: handleShowOrganizations,
      },
      {
        id: "go-to-organization",
        title: "Go to Organization",
        category: "organizations",
        hidden: !hasOrganization,
        action: handleGoToOrganization,
      },
      {
        id: "create-namespace",
        title: "Create Namespace",
        category: "namespaces",
        hidden: !canCreateNamespace,
        action: handleCreateNamespace,
      },
      {
        id: "show-namespaces",
        title: "Show Namespaces",
        category: "namespaces",
        action: handleShowNamespaces,
      },
      {
        id: "go-to-namespace",
        title: "Go to Namespace",
        category: "namespaces",
        hidden: !hasNamespace,
        action: handleGoToNamespace,
      },
      {
        id: "create-project",
        title: "Create Project",
        category: "projects",
        hidden: !canCreateProject,
        action: handleCreateProject,
      },
      {
        id: "show-projects",
        title: "Show Projects",
        category: "projects",
        hidden: !hasNamespace,
        action: handleShowProjects,
      },
      {
        id: "go-to-project",
        title: "Go to Project",
        category: "projects",
        hidden: !hasProject,
        action: handleGoToProject,
      },
    ],
    [
      canCreateNamespace,
      canCreateOrganization,
      canCreateProject,
      handleAddTodo,
      handleCreateNamespace,
      handleCreateOrganization,
      handleCreateProject,
      handleGoToNamespace,
      handleGoToOrganization,
      handleGoToProject,
      handleSetDarkTheme,
      handleSetLightTheme,
      handleSetSystemTheme,
      handleShowDocuments,
      handleShowNamespaces,
      handleShowOrganizations,
      handleShowProjects,
      handleShowTodos,
      handleToggleTheme,
      hasNamespace,
      hasOrganization,
      hasProject,
      navigate,
      mockEntries,
      recentEntities,
    ]
  );

  return (
    <>
      <CommandTrigger
        onOpen={() => setOpen(true)}
        className={cn("hidden sm:inline-flex", className)}
      />
      <Button
        type="button"
        variant="ghost"
        size="icon"
        className="sm:hidden"
        onClick={() => setOpen(true)}
        aria-label="Search or jump to"
        title="Search or jump to (⌘K)"
      >
        <SearchIcon />
      </Button>

      {open ? (
        <Suspense fallback={null}>
          <CommandPalette
            commands={commands}
            open={open}
            onOpenChange={setOpen}
            title="Search or run a command"
            placeholder="Search entities, navigation, or commands..."
          />
        </Suspense>
      ) : null}
    </>
  );
}
