import { Eye, FolderKanban, Plus } from "lucide-react";

import { commandRegistry } from "./registry";

interface ProjectCommandOptions {
  canCreate?: boolean;
  hasProject?: boolean;
  hasNamespace?: boolean;
}

export function registerProjectCommands(
  onCreateProject: () => void,
  onShowProjects: () => void,
  onGoToProject: () => void,
  options: ProjectCommandOptions = {}
): void {
  const {
    canCreate = false,
    hasProject = false,
    hasNamespace = false,
  } = options;

  const projectCommands = [
    {
      id: "create-project",
      title: "Create Project",
      description: "Create a new project in the current namespace",
      icon: <Plus className="size-4" />,
      shortcut: ["shift", "p", "n"],
      keywords: ["create", "new", "add", "project"],
      category: "projects",
      hidden: !canCreate,
      action: onCreateProject,
    },
    {
      id: "show-projects",
      title: "Show Projects",
      description: "View projects in the current namespace",
      icon: <FolderKanban className="size-4" />,
      shortcut: ["shift", "p", "s"],
      keywords: ["projects", "list", "view", "show"],
      category: "projects",
      hidden: !hasNamespace,
      action: onShowProjects,
    },
    {
      id: "go-to-project",
      title: "Go to Project",
      description: "Open the current project",
      icon: <Eye className="size-4" />,
      shortcut: ["shift", "p", "g"],
      keywords: ["go", "open", "project", "current", "detail"],
      category: "projects",
      hidden: !hasProject,
      action: onGoToProject,
    },
  ];

  projectCommands.forEach((command) => {
    commandRegistry.register(command);
  });
}

export function unregisterProjectCommands(): void {
  commandRegistry.unregister("create-project");
  commandRegistry.unregister("show-projects");
  commandRegistry.unregister("go-to-project");
}
