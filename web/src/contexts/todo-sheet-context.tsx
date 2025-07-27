"use client";

import type { ReactNode } from "react";
import { createContext, useContext, useState } from "react";

interface TodoSheetContextType {
  isOpen: boolean;
  open: () => void;
  close: () => void;
  toggle: () => void;
}

const TodoSheetContext = createContext<TodoSheetContextType | undefined>(
  undefined
);

interface TodoSheetProviderProps {
  children: ReactNode;
}

export function TodoSheetProvider({ children }: TodoSheetProviderProps) {
  const [isOpen, setIsOpen] = useState(false);

  const open = () => setIsOpen(true);
  const close = () => setIsOpen(false);
  const toggle = () => setIsOpen(!isOpen);

  return (
    <TodoSheetContext.Provider value={{ isOpen, open, close, toggle }}>
      {children}
    </TodoSheetContext.Provider>
  );
}

export function useTodoSheet() {
  const context = useContext(TodoSheetContext);
  if (context === undefined) {
    throw new Error("useTodoSheet must be used within a TodoSheetProvider");
  }
  return context;
}
