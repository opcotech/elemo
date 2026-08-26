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
import { organizationRefPath, projectIdPath } from "@/lib/api/refs";
import { Action, can } from "@/lib/auth/permissions";
import { internalPath } from "@/lib/internal-url";
import {
  namespaceAdministrationPath,
  namespaceDocumentsPath,
  namespacePath,
  namespaceProjectsPath,
  namespaceWorkPath,
  organizationDocumentsPath,
  organizationPath,
  projectActivityPath,
  projectDocumentsPath,
  projectPath,
  projectWorkPath,
} from "@/lib/paths";
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
    ...v1ProjectGetOptions({ path: projectIdPath(context.projectId ?? "") }),
    enabled: Boolean(context.projectId),
  });
  const { data: organization } = useQuery({
    ...v1OrganizationGetOptions({
      path: organizationRefPath(context.organizationId ?? ""),
    }),
    enabled: context.type === "organization" && Boolean(context.organizationId),
  });

  const navigation = useMemo(() => {
    if (
      context.type === "project" &&
      context.organizationSlug &&
      context.namespaceSlug &&
      context.projectKey
    ) {
      const projectInput = {
        organizationSlug: context.organizationSlug,
        namespaceSlug: context.namespaceSlug,
        projectKey: context.projectKey,
      };

      return [
        {
          label: "Overview",
          href: projectPath(projectInput),
          icon: LayoutDashboardIcon,
        },
        {
          label: "Work",
          href: projectWorkPath(projectInput),
          icon: ListTodoIcon,
        },
        {
          label: "Documents",
          href: projectDocumentsPath(projectInput),
          icon: FileTextIcon,
        },
        {
          label: "Activity",
          href: projectActivityPath(projectInput),
          icon: ActivityIcon,
        },
      ];
    }

    if (
      context.type === "namespace" &&
      context.organizationSlug &&
      context.namespaceSlug
    ) {
      const namespaceInput = {
        organizationSlug: context.organizationSlug,
        namespaceSlug: context.namespaceSlug,
      };
      const items: {
        label: string;
        href: string;
        icon: typeof LayoutDashboardIcon;
        separated?: boolean;
      }[] = [
        {
          label: "Overview",
          href: namespacePath(namespaceInput),
          icon: LayoutDashboardIcon,
        },
        {
          label: "Projects",
          href: namespaceProjectsPath(namespaceInput),
          icon: FolderKanbanIcon,
        },
        {
          label: "Work",
          href: namespaceWorkPath(namespaceInput),
          icon: ListTodoIcon,
        },
      ];
      if (can(namespacePermissions, Action.DocumentRead)) {
        items.push({
          label: "Documents",
          href: namespaceDocumentsPath(namespaceInput),
          icon: FileTextIcon,
        });
      }
      if (can(namespacePermissions, Action.NamespaceRead)) {
        items.push({
          label: "Administration",
          href: namespaceAdministrationPath(namespaceInput),
          icon: SettingsIcon,
          separated: true,
        });
      }
      return items;
    }

    if (context.type === "organization" && context.organizationSlug) {
      const organizationInput = {
        organizationSlug: context.organizationSlug,
      };

      return [
        {
          label: "Overview",
          href: organizationPath(organizationInput),
          icon: LayoutDashboardIcon,
        },
        {
          label: "Documents",
          href: organizationDocumentsPath(organizationInput),
          icon: FileTextIcon,
        },
      ];
    }

    return browseNavigation;
  }, [
    context.namespaceSlug,
    context.organizationSlug,
    context.projectKey,
    context.type,
    namespacePermissions,
  ]);

  useEffect(() => {
    if (currentPath.startsWith("/settings")) {
      return;
    }

    if (
      context.type === "project" &&
      project &&
      context.organizationSlug &&
      context.namespaceSlug
    ) {
      const href = internalPath(
        projectPath({
          organizationSlug: context.organizationSlug,
          namespaceSlug: context.namespaceSlug,
          projectKey: project.key,
        })
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
      const href = internalPath(
        namespacePath({
          organizationSlug: namespace.organizationSlug,
          namespaceSlug: namespace.slug,
        })
      );
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
    context.namespaceSlug,
    context.organizationSlug,
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
