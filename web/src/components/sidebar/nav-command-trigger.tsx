import { useEffect } from "react";

import { CommandPalette, CommandTrigger } from "@/components/command-palette";
import { useCommandPalette } from "@/hooks/use-command-palette";
import {
  registerThemeCommands,
  registerTodoCommands,
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
  } = useCommandPalette();

  // Register commands on mount
  useEffect(() => {
    registerTodoCommands(handleAddTodo, handleShowTodos);
    registerThemeCommands(
      handleToggleTheme,
      handleSetLightTheme,
      handleSetDarkTheme,
      handleSetSystemTheme
    );

    // Cleanup function to unregister commands
    return () => {
      unregisterTodoCommands();
      unregisterThemeCommands();
    };
  }, [
    handleAddTodo,
    handleShowTodos,
    handleToggleTheme,
    handleSetLightTheme,
    handleSetDarkTheme,
    handleSetSystemTheme,
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
