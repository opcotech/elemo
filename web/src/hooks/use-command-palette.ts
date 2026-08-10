import { useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useState } from "react";

import { useTheme } from "@/components/theme-provider";
import { useAddTodoForm } from "@/contexts/add-todo-form-context";
import { useTodoSheet } from "@/contexts/todo-sheet-context";
import { useNavigationContext } from "@/hooks/use-navigation-context";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import { can } from "@/lib/auth/permissions";

interface CommandPaletteState {
  open: boolean;
}

interface CommandPaletteActions {
  setOpen: (open: boolean) => void;
  handleAddTodo: () => void;
  handleShowTodos: () => void;
  handleToggleTheme: () => void;
  handleSetLightTheme: () => void;
  handleSetDarkTheme: () => void;
  handleSetSystemTheme: () => void;
  handleCreateOrganization: () => void;
  handleShowOrganizations: () => void;
  handleGoToOrganization: () => void;
  handleCreateNamespace: () => void;
  handleShowNamespaces: () => void;
  handleGoToNamespace: () => void;
  handleCreateProject: () => void;
  handleShowProjects: () => void;
  handleGoToProject: () => void;
  canCreateOrganization: boolean;
  canCreateNamespace: boolean;
  canCreateProject: boolean;
  hasOrganization: boolean;
  hasNamespace: boolean;
  hasProject: boolean;
}

export function useCommandPalette(): CommandPaletteState &
  CommandPaletteActions {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();
  const navigationContext = useNavigationContext();
  const { open: openTodoSheet } = useTodoSheet();
  const { open: openAddTodoForm } = useAddTodoForm();
  const { theme, setTheme } = useTheme();

  const { organizationId, namespaceId, projectId } = navigationContext;
  const hasOrganization = !!organizationId;
  const hasNamespace = !!organizationId && !!namespaceId;
  const hasProject = hasNamespace && !!projectId;

  const { data: systemOrganizationPermissions } = usePermissions(
    withResourceType(ResourceType.Organization)
  );
  const { data: organizationPermissions } = usePermissions(
    withResourceType(ResourceType.Organization, organizationId),
    !organizationId
  );
  const { data: namespacePermissions } = usePermissions(
    withResourceType(ResourceType.Namespace, namespaceId),
    !namespaceId
  );

  const canCreateOrganization = can(systemOrganizationPermissions, "create");
  const canCreateNamespace = organizationId
    ? can(organizationPermissions, "write")
    : true;
  const canCreateProject = hasNamespace && can(namespacePermissions, "write");

  const handleAddTodo = useCallback(() => {
    openAddTodoForm();
    setOpen(false);
  }, [openAddTodoForm]);

  const handleShowTodos = useCallback(() => {
    setOpen(false);
    openTodoSheet();
  }, [openTodoSheet]);

  const handleToggleTheme = useCallback(() => {
    setTheme(theme === "light" ? "dark" : "light");
  }, [theme, setTheme]);

  const handleSetLightTheme = useCallback(() => {
    setTheme("light");
  }, [setTheme]);

  const handleSetDarkTheme = useCallback(() => {
    setTheme("dark");
  }, [setTheme]);

  const handleSetSystemTheme = useCallback(() => {
    setTheme("system");
  }, [setTheme]);

  const handleCreateOrganization = useCallback(() => {
    setOpen(false);
    navigate({ to: "/settings/organizations/new" });
  }, [navigate]);

  const handleShowOrganizations = useCallback(() => {
    setOpen(false);
    navigate({ to: "/settings/organizations" });
  }, [navigate]);

  const handleGoToOrganization = useCallback(() => {
    if (!organizationId) return;
    setOpen(false);
    navigate({
      to: "/settings/organizations/$organizationId",
      params: { organizationId },
    });
  }, [navigate, organizationId]);

  const handleCreateNamespace = useCallback(() => {
    setOpen(false);
    if (organizationId) {
      navigate({
        to: "/settings/organizations/$organizationId/namespaces/new",
        params: { organizationId },
      });
      return;
    }
    navigate({ to: "/settings/namespaces/new" });
  }, [navigate, organizationId]);

  const handleShowNamespaces = useCallback(() => {
    setOpen(false);
    navigate({ to: "/settings/namespaces" });
  }, [navigate]);

  const handleGoToNamespace = useCallback(() => {
    if (!organizationId || !namespaceId) return;
    setOpen(false);
    navigate({
      to: "/settings/organizations/$organizationId/namespaces/$namespaceId",
      params: { organizationId, namespaceId },
    });
  }, [navigate, organizationId, namespaceId]);

  const handleCreateProject = useCallback(() => {
    if (!organizationId || !namespaceId) return;
    setOpen(false);
    navigate({
      to: "/settings/organizations/$organizationId/namespaces/$namespaceId/projects/new",
      params: { organizationId, namespaceId },
    });
  }, [navigate, organizationId, namespaceId]);

  const handleShowProjects = useCallback(() => {
    if (!organizationId || !namespaceId) return;
    setOpen(false);
    navigate({
      to: "/settings/organizations/$organizationId/namespaces/$namespaceId",
      params: { organizationId, namespaceId },
    });
  }, [navigate, organizationId, namespaceId]);

  const handleGoToProject = useCallback(() => {
    if (!organizationId || !namespaceId || !projectId) return;
    setOpen(false);
    navigate({
      to: "/settings/organizations/$organizationId/namespaces/$namespaceId/projects/$projectId",
      params: { organizationId, namespaceId, projectId },
    });
  }, [navigate, organizationId, namespaceId, projectId]);

  // Consolidated keyboard event handling
  useEffect(() => {
    let keySequence: string[] = [];
    let sequenceTimeout: ReturnType<typeof setTimeout> | null = null;

    function clearSequence() {
      keySequence = [];
      if (sequenceTimeout) {
        clearTimeout(sequenceTimeout);
        sequenceTimeout = null;
      }
    }

    function handleKeyDown(e: KeyboardEvent) {
      // Handle Cmd+K / Ctrl+K to open command palette
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        const active = document.activeElement;
        if (
          !open &&
          active &&
          active.tagName !== "INPUT" &&
          active.tagName !== "TEXTAREA" &&
          active.getAttribute("contenteditable") !== "true"
        ) {
          e.preventDefault();
          setOpen(true);
          return;
        }
      }

      // Handle command palette shortcuts only when palette is open
      if (open) {
        // Handle key sequences (e.g., Shift+T+N for "Add todo")
        if (e.shiftKey && !e.metaKey && !e.ctrlKey && !e.altKey) {
          const key = e.key.toLowerCase();
          if (/^[a-z]$/.test(key)) {
            e.preventDefault();
            keySequence = [...keySequence, key].slice(-2); // Keep last 2 keys

            // Clear existing timeout
            if (sequenceTimeout) {
              clearTimeout(sequenceTimeout);
            }

            // Set timeout to clear sequence
            sequenceTimeout = setTimeout(() => {
              keySequence = [];
            }, 1000);

            const [first, second] = keySequence;
            if (!second) return;

            // Todo shortcuts
            if (first === "t" && second === "n") {
              handleAddTodo();
              clearSequence();
              return;
            }
            if (first === "t" && second === "s") {
              handleShowTodos();
              clearSequence();
              return;
            }

            // Theme shortcuts
            if (first === "t" && second === "t") {
              handleToggleTheme();
              clearSequence();
              return;
            }
            if (first === "t" && second === "l") {
              handleSetLightTheme();
              clearSequence();
              return;
            }
            if (first === "t" && second === "d") {
              handleSetDarkTheme();
              clearSequence();
              return;
            }

            // Organization shortcuts
            if (first === "o" && second === "n" && canCreateOrganization) {
              handleCreateOrganization();
              clearSequence();
              return;
            }
            if (first === "o" && second === "s") {
              handleShowOrganizations();
              clearSequence();
              return;
            }
            if (first === "o" && second === "g" && hasOrganization) {
              handleGoToOrganization();
              clearSequence();
              return;
            }

            // Namespace shortcuts
            if (first === "n" && second === "n" && canCreateNamespace) {
              handleCreateNamespace();
              clearSequence();
              return;
            }
            if (first === "n" && second === "s") {
              handleShowNamespaces();
              clearSequence();
              return;
            }
            if (first === "n" && second === "g" && hasNamespace) {
              handleGoToNamespace();
              clearSequence();
              return;
            }

            // Project shortcuts
            if (first === "p" && second === "n" && canCreateProject) {
              handleCreateProject();
              clearSequence();
              return;
            }
            if (first === "p" && second === "s" && hasNamespace) {
              handleShowProjects();
              clearSequence();
              return;
            }
            if (first === "p" && second === "g" && hasProject) {
              handleGoToProject();
              clearSequence();
              return;
            }
          }
        }
      }
    }

    window.addEventListener("keydown", handleKeyDown);

    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      if (sequenceTimeout) {
        clearTimeout(sequenceTimeout);
      }
    };
  }, [
    open,
    handleAddTodo,
    handleShowTodos,
    handleToggleTheme,
    handleSetLightTheme,
    handleSetDarkTheme,
    handleCreateOrganization,
    handleShowOrganizations,
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
  ]);

  return {
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
  };
}
