import {
  FileTextIcon,
  FolderKanbanIcon,
  ListTodoIcon,
  XIcon,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { InternalLink } from "@/components/ui/internal-link";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import type { RecentEntityType } from "@/lib/ui-store";
import { uiActions, useUiSelector } from "@/lib/ui-store";
import { cn } from "@/lib/utils";

function RecentEntitiesSection({
  type,
  label,
  icon: Icon,
}: {
  type: RecentEntityType;
  label: string;
  icon: LucideIcon;
}) {
  const entities = useUiSelector((state) =>
    state.recentEntities.filter((entity) => entity.type === type)
  );

  if (entities.length === 0) {
    return null;
  }

  return (
    <SidebarGroup className="shrink-0">
      <SidebarGroupLabel>{label}</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {entities.map((entity) => (
            <SidebarMenuItem
              key={`${entity.type}-${entity.id}`}
              className={cn(
                "has-[[data-sidebar=menu-action]:hover]:[&_[data-sidebar=menu-button]]:bg-destructive/10",
                "has-[[data-sidebar=menu-action]:hover]:[&_[data-sidebar=menu-button]]:text-destructive",
                "has-[[data-sidebar=menu-action]:hover]:[&_[data-sidebar=menu-button]]:ring-1",
                "has-[[data-sidebar=menu-action]:hover]:[&_[data-sidebar=menu-button]]:ring-inset",
                "has-[[data-sidebar=menu-action]:hover]:[&_[data-sidebar=menu-button]]:ring-destructive/20",
                "has-[[data-sidebar=menu-action]:hover]:[&_[data-sidebar=menu-button]]:[&_svg]:text-destructive"
              )}
            >
              <SidebarMenuButton
                render={<InternalLink to={entity.href} />}
                tooltip={entity.label}
                size="sm"
                className={cn(
                  "group-hover/menu-item:bg-primary/10",
                  "group-hover/menu-item:text-primary-on-subtle",
                  "group-hover/menu-item:ring-1",
                  "group-hover/menu-item:ring-inset",
                  "group-hover/menu-item:ring-primary/20"
                )}
              >
                <Icon />
                <span>{entity.label}</span>
              </SidebarMenuButton>
              <SidebarMenuAction
                showOnHover
                aria-label={`Remove ${entity.label} from recents`}
                title="Remove from recents"
                className="hover:text-destructive hover:bg-transparent hover:ring-0"
                onClick={(event) => {
                  event.preventDefault();
                  event.stopPropagation();
                  uiActions.forgetRecentEntity({
                    id: entity.id,
                    type: entity.type,
                  });
                }}
              >
                <XIcon />
              </SidebarMenuAction>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}

export function RecentWorkItems() {
  return (
    <RecentEntitiesSection
      type="work-item"
      label="Recent Work Items"
      icon={ListTodoIcon}
    />
  );
}

export function RecentDocuments() {
  return (
    <RecentEntitiesSection
      type="document"
      label="Recent Documents"
      icon={FileTextIcon}
    />
  );
}

export function RecentProjects() {
  return (
    <RecentEntitiesSection
      type="project"
      label="Recent Projects"
      icon={FolderKanbanIcon}
    />
  );
}
