import { CheckIcon, ChevronsUpDownIcon, Layers3Icon } from "lucide-react";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { InternalLink } from "@/components/ui/internal-link";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useNavigationContext } from "@/hooks/use-navigation-context";
import { useAccessibleNamespaces } from "@/lib/api/accessible-namespaces";
import { internalPath } from "@/lib/internal-url";
import { uiActions } from "@/lib/ui-store";
import { cn } from "@/lib/utils";

export function NamespaceSwitcher() {
  const { state: sidebarState, isMobile } = useSidebar();
  const context = useNavigationContext();
  const {
    data: accessibleWorkspace,
    isLoading,
    isError,
  } = useAccessibleNamespaces();
  const organizations = accessibleWorkspace?.organizations ?? [];
  const namespaces = accessibleWorkspace?.namespaces ?? [];
  const activeNamespaceId = context.namespaceId;
  const activeNamespace = namespaces.find(
    (namespace) => namespace.id === activeNamespaceId
  );
  const isCollapsed = sidebarState === "collapsed" && !isMobile;
  const switcherLabel = activeNamespace?.name ?? "Choose namespace";

  if (isLoading && namespaces.length === 0) {
    return (
      <Skeleton
        className={cn(
          "h-12 w-full rounded-md",
          "group-data-[collapsible=icon]:mx-auto group-data-[collapsible=icon]:size-9 group-data-[collapsible=icon]:rounded-md"
        )}
      />
    );
  }

  if (isError && namespaces.length === 0) {
    return (
      <div className="text-muted-foreground px-2 py-2 text-sm">
        Namespace list unavailable
      </div>
    );
  }

  return (
    <SidebarMenu className="group-data-[collapsible=icon]:items-center">
      <SidebarMenuItem className="group-data-[collapsible=icon]:flex group-data-[collapsible=icon]:w-full group-data-[collapsible=icon]:justify-center">
        <DropdownMenu>
          <Tooltip>
            <TooltipTrigger
              render={
                <DropdownMenuTrigger
                  render={
                    <SidebarMenuButton
                      size="lg"
                      aria-label={`Switch namespace, current: ${switcherLabel}`}
                      className={cn(
                        "border-sidebar-border bg-surface-raised hover:border-primary/25 h-12 border px-2.5 shadow-none",
                        "group-data-[collapsible=icon]:size-9! group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:gap-0 group-data-[collapsible=icon]:p-0!"
                      )}
                    />
                  }
                />
              }
            >
              <div className="bg-primary-subtle text-primary-on-subtle flex size-8 shrink-0 items-center justify-center rounded-md group-data-[collapsible=icon]:size-7">
                <Layers3Icon className="size-4" aria-hidden />
              </div>
              <div className="min-w-0 flex-1 text-left leading-tight group-data-[collapsible=icon]:hidden">
                <span className="block truncate text-sm font-medium">
                  {switcherLabel}
                </span>
                <span className="text-muted-foreground block truncate text-xs">
                  {activeNamespace?.organizationName ??
                    (namespaces.length === 0
                      ? "No namespaces available"
                      : "Your workspace")}
                </span>
              </div>
              <ChevronsUpDownIcon className="text-muted-foreground ml-auto size-4 group-data-[collapsible=icon]:hidden" />
            </TooltipTrigger>
            <TooltipContent side="right" align="center" hidden={!isCollapsed}>
              {switcherLabel}
            </TooltipContent>
          </Tooltip>
          <DropdownMenuContent
            side={isCollapsed ? "right" : "bottom"}
            align="start"
            sideOffset={8}
            className="min-w-64"
          >
            {organizations.map((organization, organizationIndex) => {
              const organizationNamespaces = namespaces.filter(
                (namespace) => namespace.organizationId === organization.id
              );

              if (organizationNamespaces.length === 0) {
                return null;
              }

              return (
                <DropdownMenuGroup key={organization.id}>
                  {organizationIndex > 0 && <DropdownMenuSeparator />}
                  <DropdownMenuLabel>{organization.name}</DropdownMenuLabel>
                  {organizationNamespaces.map((namespace) => {
                    const href = internalPath(`/namespaces/${namespace.id}`);
                    const isActive = namespace.id === activeNamespaceId;

                    return (
                      <DropdownMenuItem
                        key={namespace.id}
                        render={<InternalLink to={internalPath(href)} />}
                        onClick={() =>
                          uiActions.rememberRecentEntity({
                            id: namespace.id,
                            type: "namespace",
                            label: namespace.name,
                            href,
                            namespaceId: namespace.id,
                          })
                        }
                      >
                        <span className="bg-muted text-muted-foreground flex size-6 items-center justify-center rounded-md text-xs font-semibold">
                          {namespace.name.slice(0, 1).toUpperCase()}
                        </span>
                        <span className="min-w-0 flex-1 truncate">
                          {namespace.name}
                        </span>
                        {isActive && (
                          <CheckIcon className="text-primary size-4" />
                        )}
                      </DropdownMenuItem>
                    );
                  })}
                </DropdownMenuGroup>
              );
            })}
            {namespaces.length === 0 && (
              <p className="text-muted-foreground px-2 py-3 text-sm">
                No namespaces available.
              </p>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
