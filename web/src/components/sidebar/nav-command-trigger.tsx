import { useEffect } from "react";

import { CommandPalette, CommandTrigger } from "@/components/command-palette";
import { useCommandPalette } from "@/hooks/use-command-palette";
import {
  registerNamespaceCommands,
  registerOrganizationCommands,
  registerProjectCommands,
  registerThemeCommands,
  registerTodoCommands,
  unregisterNamespaceCommands,
  unregisterOrganizationCommands,
  unregisterProjectCommands,
  unregisterThemeCommands,
  unregisterTodoCommands,
} from "@/lib/commands";

export function NavCommandTrigger() {
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

  // Register commands on mount / when handlers or visibility change
  useEffect(() => {
    registerTodoCommands(handleAddTodo, handleShowTodos);
    registerThemeCommands(
      handleToggleTheme,
      handleSetLightTheme,
      handleSetDarkTheme,
      handleSetSystemTheme
    );
    registerOrganizationCommands(
      handleCreateOrganization,
      handleShowOrganizations,
      handleGoToOrganization,
      {
        canCreate: canCreateOrganization,
        hasOrganization,
      }
    );
    registerNamespaceCommands(
      handleCreateNamespace,
      handleShowNamespaces,
      handleGoToNamespace,
      {
        canCreate: canCreateNamespace,
        hasNamespace,
      }
    );
    registerProjectCommands(
      handleCreateProject,
      handleShowProjects,
      handleGoToProject,
      {
        canCreate: canCreateProject,
        hasProject,
        hasNamespace,
      }
    );

    return () => {
      unregisterTodoCommands();
      unregisterThemeCommands();
      unregisterOrganizationCommands();
      unregisterNamespaceCommands();
      unregisterProjectCommands();
    };
  }, [
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
  ]);

  return (
    <>
      <CommandTrigger onOpen={() => setOpen(true)} />

      <CommandPalette
        open={open}
        onOpenChange={setOpen}
        title="Quick Actions"
        placeholder="Type a command or select an action..."
      />
    </>
  );
}
