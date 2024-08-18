import { ListTodo, Plus } from "lucide-react";

import { commandRegistry } from "./registry";

export function registerTodoCommands(
  onAddTodo: () => void,
  onShowTodos: () => void
): void {
  const todoCommands = [
    {
      id: "add-todo",
      title: "Add Todo",
      description: "Create a new todo item",
      icon: <Plus className="mr-2 h-4 w-4" />,
      shortcut: ["shift", "t", "n"],
      keywords: ["create", "new", "add", "todo", "task"],
      category: "quick-actions",
      action: onAddTodo,
    },
    {
      id: "show-todos",
      title: "Show Todos",
      description: "Open todo list sheet",
      icon: <ListTodo className="mr-2 h-4 w-4" />,
      shortcut: ["shift", "t", "s"],
      keywords: ["todos", "tasks", "list", "view", "show"],
      category: "quick-actions",
      action: onShowTodos,
    },
  ];

  // Register all todo commands
  todoCommands.forEach((command) => {
    commandRegistry.register(command);
  });
}

export function unregisterTodoCommands(): void {
  commandRegistry.unregister("add-todo");
  commandRegistry.unregister("show-todos");
}
