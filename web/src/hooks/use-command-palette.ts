import { useCallback, useEffect, useState } from "react";

import { useTheme } from "@/components/theme-provider";
import { useAddTodoForm } from "@/contexts/add-todo-form-context";
import { useTodoSheet } from "@/contexts/todo-sheet-context";

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
}

export function useCommandPalette(): CommandPaletteState &
  CommandPaletteActions {
  const [open, setOpen] = useState(false);
  const { open: openTodoSheet } = useTodoSheet();
  const { open: openAddTodoForm } = useAddTodoForm();
  const { theme, setTheme } = useTheme();

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

  // Consolidated keyboard event handling
  useEffect(() => {
    let keySequence: string[] = [];
    let sequenceTimeout: NodeJS.Timeout | null = null;

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

            // Check for matching shortcuts
            if (keySequence[0] === "t" && keySequence[1] === "n") {
              handleAddTodo();
              keySequence = [];
              if (sequenceTimeout) {
                clearTimeout(sequenceTimeout);
                sequenceTimeout = null;
              }
              return;
            }

            if (keySequence[0] === "t" && keySequence[1] === "s") {
              handleShowTodos();
              keySequence = [];
              if (sequenceTimeout) {
                clearTimeout(sequenceTimeout);
                sequenceTimeout = null;
              }
              return;
            }

            // Theme shortcuts
            if (keySequence[0] === "t" && keySequence[1] === "t") {
              handleToggleTheme();
              keySequence = [];
              if (sequenceTimeout) {
                clearTimeout(sequenceTimeout);
                sequenceTimeout = null;
              }
              return;
            }

            if (keySequence[0] === "t" && keySequence[1] === "l") {
              handleSetLightTheme();
              keySequence = [];
              if (sequenceTimeout) {
                clearTimeout(sequenceTimeout);
                sequenceTimeout = null;
              }
              return;
            }

            if (keySequence[0] === "t" && keySequence[1] === "d") {
              handleSetDarkTheme();
              keySequence = [];
              if (sequenceTimeout) {
                clearTimeout(sequenceTimeout);
                sequenceTimeout = null;
              }
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
  }, [open, handleAddTodo, handleShowTodos]);

  return {
    open,
    setOpen,
    handleAddTodo,
    handleShowTodos,
    handleToggleTheme,
    handleSetLightTheme,
    handleSetDarkTheme,
    handleSetSystemTheme,
  };
}
