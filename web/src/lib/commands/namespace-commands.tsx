import { Eye, Folder, Plus } from "lucide-react";

import { commandRegistry } from "./registry";

interface NamespaceCommandOptions {
  canCreate?: boolean;
  hasNamespace?: boolean;
}

export function registerNamespaceCommands(
  onCreateNamespace: () => void,
  onShowNamespaces: () => void,
  onGoToNamespace: () => void,
  options: NamespaceCommandOptions = {}
): void {
  const { canCreate = true, hasNamespace = false } = options;

  const namespaceCommands = [
    {
      id: "create-namespace",
      title: "Create Namespace",
      description: "Create a new namespace",
      icon: <Plus className="size-4" />,
      shortcut: ["shift", "n", "n"],
      keywords: ["create", "new", "add", "namespace"],
      category: "namespaces",
      hidden: !canCreate,
      action: onCreateNamespace,
    },
    {
      id: "show-namespaces",
      title: "Show Namespaces",
      description: "View all namespaces",
      icon: <Folder className="size-4" />,
      shortcut: ["shift", "n", "s"],
      keywords: ["namespaces", "list", "view", "show", "settings"],
      category: "namespaces",
      action: onShowNamespaces,
    },
    {
      id: "go-to-namespace",
      title: "Go to Namespace",
      description: "Open the current namespace",
      icon: <Eye className="size-4" />,
      shortcut: ["shift", "n", "g"],
      keywords: ["go", "open", "namespace", "current", "detail"],
      category: "namespaces",
      hidden: !hasNamespace,
      action: onGoToNamespace,
    },
  ];

  namespaceCommands.forEach((command) => {
    commandRegistry.register(command);
  });
}

export function unregisterNamespaceCommands(): void {
  commandRegistry.unregister("create-namespace");
  commandRegistry.unregister("show-namespaces");
  commandRegistry.unregister("go-to-namespace");
}
