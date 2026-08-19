import { useQuery } from "@tanstack/react-query";
import { useRouterState } from "@tanstack/react-router";
import {
  ActivityIcon,
  Building2Icon,
  FileTextIcon,
  FolderKanbanIcon,
  Layers3Icon,
  LayoutDashboardIcon,
  ListTodoIcon,
  SettingsIcon,
} from "lucide-react";
import { useEffect, useMemo } from "react";

import { InternalLink } from "@/components/ui/internal-link";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { useNavigationContext } from "@/hooks/use-navigation-context";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import { useAccessibleNamespaces } from "@/lib/api/accessible-namespaces";
import {
  v1OrganizationGetOptions,
  v1ProjectGetOptions,
} from "@/lib/api/query-options";
import { Action, can } from "@/lib/auth/permissions";
import { internalPath } from "@/lib/internal-url";
import { uiActions } from "@/lib/ui-store";

const browseNavigation = [
  {
    label: "Organizations",
    href: "/organizations",
    icon: Building2Icon,
  },
  {
    label: "Namespaces",
    href: "/namespaces",
    icon: Layers3Icon,
  },
] as const;

export function ContextualNavigationSection() {
  const context = useNavigationContext();
  const currentPath = useRouterState({
    select: (state) => state.location.pathname,
  });
  const { data: accessibleWorkspace } = useAccessibleNamespaces();
  const namespace = accessibleWorkspace?.namespaces.find(
    (item) => item.id === context.namespaceId
  );
  const { data: namespacePermissions } = usePermissions(
    withResourceType(ResourceType.Namespace, context.namespaceId ?? ""),
    context.type !== "namespace" || !context.namespaceId
  );
  const { data: project } = useQuery({
    ...v1ProjectGetOptions({ path: { id: context.projectId ?? "" } }),
    enabled: Boolean(context.projectId),
  });
  const { data: organization } = useQuery({
    ...v1OrganizationGetOptions({
      path: { id: context.organizationId ?? "" },
    }),
    enabled: context.type === "organization" && Boolean(context.organizationId),
  });

  const navigation = useMemo(() => {
    if (
      context.type === "project" &&
      context.projectId &&
      context.namespaceId
    ) {
      const projectBase = `/namespaces/${context.namespaceId}/projects/${context.projectId}`;

      return [
        {
          label: "Overview",
          href: projectBase,
          icon: LayoutDashboardIcon,
        },
        { label: "Work", href: `${projectBase}/work`, icon: ListTodoIcon },
        {
          label: "Documents",
          href: `${projectBase}/documents`,
          icon: FileTextIcon,
        },
        {
          label: "Activity",
          href: `${projectBase}/activity`,
          icon: ActivityIcon,
        },
      ];
    }

    if (context.type === "namespace" && context.namespaceId) {
      const namespaceBase = `/namespaces/${context.namespaceId}`;
      const items: {
        label: string;
        href: string;
        icon: typeof LayoutDashboardIcon;
        separated?: boolean;
      }[] = [
        {
          label: "Overview",
          href: namespaceBase,
          icon: LayoutDashboardIcon,
        },
        {
          label: "Projects",
          href: `${namespaceBase}/projects`,
          icon: FolderKanbanIcon,
        },
        {
          label: "Work",
          href: `${namespaceBase}/work`,
          icon: ListTodoIcon,
        },
      ];
      if (can(namespacePermissions, Action.DocumentRead)) {
        items.push({
          label: "Documents",
          href: `${namespaceBase}/documents`,
          icon: FileTextIcon,
        });
      }
      if (can(namespacePermissions, Action.NamespaceRead)) {
        items.push({
          label: "Administration",
          href: `${namespaceBase}/administration`,
          icon: SettingsIcon,
          separated: true,
        });
      }
      return items;
    }

    if (context.type === "organization" && context.organizationId) {
      const organizationBase = `/organizations/${context.organizationId}`;

      return [
        {
          label: "Overview",
          href: organizationBase,
          icon: LayoutDashboardIcon,
        },
        {
          label: "Documents",
          href: `${organizationBase}/documents`,
          icon: FileTextIcon,
        },
      ];
    }

    return browseNavigation;
  }, [
    context.namespaceId,
    context.organizationId,
    context.projectId,
    context.type,
    namespacePermissions,
  ]);

  useEffect(() => {
    if (currentPath.startsWith("/settings")) {
      return;
    }

    if (context.type === "project" && project && context.namespaceId) {
      const href = internalPath(
        `/namespaces/${context.namespaceId}/projects/${project.id}`
      );
      uiActions.rememberRecentEntity({
        id: project.id,
        type: "project",
        label: project.name,
        href,
        namespaceId: context.namespaceId,
      });
      return;
    }

    if (context.type === "namespace" && namespace) {
      const href = internalPath(`/namespaces/${namespace.id}`);
      uiActions.rememberRecentEntity({
        id: namespace.id,
        type: "namespace",
        label: namespace.name,
        href,
        namespaceId: namespace.id,
      });
    }
  }, [
    context.namespaceId,
    context.projectId,
    context.type,
    currentPath,
    namespace,
    project,
  ]);

  const isBrowse = context.type === "global";
  const groupLabel = isBrowse
    ? "Browse"
    : (project?.name ??
      namespace?.name ??
      organization?.name ??
      "Current context");

  return (
    <SidebarGroup>
      <SidebarGroupLabel className="h-auto min-h-8 py-1">
        <span className="truncate">{groupLabel}</span>
      </SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {navigation.map((item) => (
            <SidebarMenuItem
              key={item.href}
              className={
                "separated" in item && item.separated
                  ? "mt-2 border-t pt-2"
                  : undefined
              }
            >
              <SidebarMenuButton
                render={<InternalLink to={internalPath(item.href)} />}
                tooltip={item.label}
                isActive={
                  currentPath === item.href ||
                  (item.label !== "Overview" &&
                    currentPath.startsWith(`${item.href}/`))
                }
              >
                <item.icon />
                <span>{item.label}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
