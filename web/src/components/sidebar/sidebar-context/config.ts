import { FileText, LayoutDashboard, Settings, Target } from "lucide-react";

import type { NavigationConfigBuilder, NavigationItemConfig } from "./types";

export const namespaceNavigationConfig: NavigationConfigBuilder = (ctx, ns) => {
  if (!ctx.namespaceId || !ctx.organizationId || !ns) return null;

  const basePath = `/settings/organizations/${ctx.organizationId}/namespaces/${ctx.namespaceId}`;
  const items: NavigationItemConfig[] = [
    { label: "Settings", pathSuffix: "/edit", icon: Settings },
  ];

  return {
    label: ns.name || "Namespace",
    items: items.map((item) => ({
      ...item,
      href: `${basePath}${item.pathSuffix || ""}`,
    })),
  };
};

export const projectNavigationConfig: NavigationConfigBuilder = (ctx) => {
  if (!ctx.projectId) return null;

  const basePath = `/projects/${ctx.projectId}`;
  const items: NavigationItemConfig[] = [
    { label: "Overview", icon: LayoutDashboard },
    { label: "Issues", pathSuffix: "/issues", icon: Target },
    { label: "Documents", pathSuffix: "/documents", icon: FileText },
    { label: "Roadmap", pathSuffix: "/roadmap", icon: Target },
    { label: "Settings", pathSuffix: "/settings", icon: Settings },
  ];

  return {
    label: "Project",
    items: items.map((item) => ({
      ...item,
      href: `${basePath}${item.pathSuffix || ""}`,
    })),
  };
};

// Map context types to their navigation config builders
export const configBuilders: Record<
  "namespace" | "project",
  NavigationConfigBuilder
> = {
  namespace: namespaceNavigationConfig,
  project: projectNavigationConfig,
};
