import { Building2, Eye, Plus } from "lucide-react";

import { commandRegistry } from "./registry";

interface OrganizationCommandOptions {
  canCreate?: boolean;
  hasOrganization?: boolean;
}

export function registerOrganizationCommands(
  onCreateOrganization: () => void,
  onShowOrganizations: () => void,
  onGoToOrganization: () => void,
  options: OrganizationCommandOptions = {}
): void {
  const { canCreate = true, hasOrganization = false } = options;

  const organizationCommands = [
    {
      id: "create-organization",
      title: "Create Organization",
      description: "Create a new organization",
      icon: <Plus className="size-4" />,
      shortcut: ["shift", "o", "n"],
      keywords: ["create", "new", "add", "organization", "org"],
      category: "organizations",
      hidden: !canCreate,
      action: onCreateOrganization,
    },
    {
      id: "show-organizations",
      title: "Show Organizations",
      description: "View all organizations",
      icon: <Building2 className="size-4" />,
      shortcut: ["shift", "o", "s"],
      keywords: ["organizations", "orgs", "list", "view", "show", "settings"],
      category: "organizations",
      action: onShowOrganizations,
    },
    {
      id: "go-to-organization",
      title: "Go to Organization",
      description: "Open the current organization",
      icon: <Eye className="size-4" />,
      shortcut: ["shift", "o", "g"],
      keywords: ["go", "open", "organization", "org", "current", "detail"],
      category: "organizations",
      hidden: !hasOrganization,
      action: onGoToOrganization,
    },
  ];

  organizationCommands.forEach((command) => {
    commandRegistry.register(command);
  });
}

export function unregisterOrganizationCommands(): void {
  commandRegistry.unregister("create-organization");
  commandRegistry.unregister("show-organizations");
  commandRegistry.unregister("go-to-organization");
}
