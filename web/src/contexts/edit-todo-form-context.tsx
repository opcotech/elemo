"use client";

import type { ReactNode } from "react";
import { createContext, useContext, useState } from "react";

import type { TodoPriority } from "@/lib/api";

interface TodoItem {
  id: string;
  title: string;
  description: string;
  priority: TodoPriority;
  completed: boolean;
  due_date: string | null;
  created_at: string;
}

interface EditTodoFormContextType {
  isOpen: boolean;
  todo: TodoItem | null;
  open: (todo: TodoItem) => void;
  close: () => void;
  toggle: () => void;
}

const EditTodoFormContext = createContext<EditTodoFormContextType | undefined>(
  undefined
);

interface EditTodoFormProviderProps {
  children: ReactNode;
}

export function EditTodoFormProvider({ children }: EditTodoFormProviderProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [todo, setTodo] = useState<TodoItem | null>(null);

  const open = (todoItem: TodoItem) => {
    setTodo(todoItem);
    setIsOpen(true);
  };

  const close = () => {
    setIsOpen(false);
    setTodo(null);
  };

  const toggle = () => {
    if (isOpen) {
      close();
    }
  };

  return (
    <EditTodoFormContext.Provider value={{ isOpen, todo, open, close, toggle }}>
      {children}
    </EditTodoFormContext.Provider>
  );
}

export function useEditTodoForm() {
  const context = useContext(EditTodoFormContext);
  if (context === undefined) {
    throw new Error(
      "useEditTodoForm must be used within an EditTodoFormProvider"
    );
  }
  return context;
}
