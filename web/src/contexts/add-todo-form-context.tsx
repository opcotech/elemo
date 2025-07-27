"use client";

import type { ReactNode } from "react";
import { createContext, useContext, useState } from "react";

interface AddTodoFormContextType {
  isOpen: boolean;
  open: () => void;
  close: () => void;
  toggle: () => void;
}

const AddTodoFormContext = createContext<AddTodoFormContextType | undefined>(
  undefined
);

interface AddTodoFormProviderProps {
  children: ReactNode;
}

export function AddTodoFormProvider({ children }: AddTodoFormProviderProps) {
  const [isOpen, setIsOpen] = useState(false);

  const open = () => setIsOpen(true);
  const close = () => setIsOpen(false);
  const toggle = () => setIsOpen(!isOpen);

  return (
    <AddTodoFormContext.Provider value={{ isOpen, open, close, toggle }}>
      {children}
    </AddTodoFormContext.Provider>
  );
}

export function useAddTodoForm() {
  const context = useContext(AddTodoFormContext);
  if (context === undefined) {
    throw new Error(
      "useAddTodoForm must be used within an AddTodoFormProvider"
    );
  }
  return context;
}
