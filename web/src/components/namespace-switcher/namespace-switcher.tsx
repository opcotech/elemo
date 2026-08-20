import { useNavigate } from "@tanstack/react-router";
import { ChevronsUpDownIcon, Layers3Icon } from "lucide-react";
import { useMemo, useState } from "react";

import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
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
import type { AccessibleNamespace } from "@/lib/api/accessible-namespaces";
import { useAccessibleNamespaces } from "@/lib/api/accessible-namespaces";
import { internalPath } from "@/lib/internal-url";
import { uiActions } from "@/lib/ui-store";
import { cn } from "@/lib/utils";

function namespaceCommandValue(
  namespace: AccessibleNamespace,
  organizationName: string
) {
  return [namespace.name, organizationName, namespace.id]
    .filter(Boolean)
    .join(" ");
}

export function NamespaceSwitcher() {
  const navigate = useNavigate();
  const { state: sidebarState, isMobile } = useSidebar();
  const context = useNavigationContext();
  const {
    data: accessibleWorkspace,
    isLoading,
    isError,
  } = useAccessibleNamespaces();
  const [open, setOpen] = useState(false);
  const organizations = accessibleWorkspace?.organizations ?? [];
  const namespaces = accessibleWorkspace?.namespaces ?? [];
  const activeNamespaceId = context.namespaceId;
  const activeNamespace = namespaces.find(
    (namespace) => namespace.id === activeNamespaceId
  );
  const isCollapsed = sidebarState === "collapsed" && !isMobile;
  const switcherLabel = activeNamespace?.name ?? "Choose namespace";
  const namespaceGroups = useMemo(
    () =>
      organizations
        .map((organization) => ({
          organization,
          namespaces: namespaces.filter(
            (namespace) => namespace.organizationId === organization.id
          ),
        }))
        .filter((group) => group.namespaces.length > 0),
    [namespaces, organizations]
  );

  const selectNamespace = (namespace: AccessibleNamespace) => {
    const href = internalPath(`/namespaces/${namespace.id}`);
    uiActions.rememberRecentEntity({
      id: namespace.id,
      type: "namespace",
      label: namespace.name,
      href,
      namespaceId: namespace.id,
    });
    setOpen(false);
    void navigate({
      to: "/namespaces/$namespaceId",
      params: { namespaceId: namespace.id },
    });
  };

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
        <Popover open={open} onOpenChange={setOpen}>
          <Tooltip>
            <TooltipTrigger
              render={
                <PopoverTrigger
                  render={
                    <SidebarMenuButton
                      size="lg"
                      role="combobox"
                      aria-expanded={open}
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
          <PopoverContent
            side={isCollapsed ? "right" : "bottom"}
            align="start"
            sideOffset={8}
            className="min-w-64 gap-0 p-0"
          >
            <Command>
              <CommandInput placeholder="Search namespaces..." />
              <CommandList>
                <CommandEmpty>
                  {namespaces.length === 0
                    ? "No namespaces available."
                    : "No namespace found."}
                </CommandEmpty>
                {namespaceGroups.map(({ organization, namespaces: items }) => (
                  <CommandGroup
                    key={organization.id}
                    heading={organization.name}
                  >
                    {items.map((namespace) => (
                      <CommandItem
                        key={namespace.id}
                        value={namespaceCommandValue(
                          namespace,
                          organization.name
                        )}
                        data-checked={
                          namespace.id === activeNamespaceId || undefined
                        }
                        onSelect={() => selectNamespace(namespace)}
                      >
                        <span className="bg-muted text-muted-foreground flex size-6 items-center justify-center rounded-md text-xs font-semibold">
                          {namespace.name.slice(0, 1).toUpperCase()}
                        </span>
                        <span className="min-w-0 flex-1 truncate">
                          {namespace.name}
                        </span>
                      </CommandItem>
                    ))}
                  </CommandGroup>
                ))}
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
