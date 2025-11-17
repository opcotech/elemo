import { Link, useRouter } from "@tanstack/react-router";
import { ChevronRight, FileText, Folder } from "lucide-react";
import { useState } from "react";

import type { NavigationContext } from "./types";

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from "@/components/ui/sidebar";
import type { Namespace } from "@/lib/api";
import { cn } from "@/lib/utils";

interface NamespaceNavigationProps {
  namespace: Namespace;
  context: NavigationContext;
  navigationItems: { href: string; label: string; icon: React.ElementType }[];
}

export function NamespaceNavigation({
  namespace,
  context,
  navigationItems,
}: NamespaceNavigationProps) {
  const router = useRouter();
  const currentPath = router.state.location.pathname;
  const [expandedProjects, setExpandedProjects] = useState(true);
  const [expandedDocuments, setExpandedDocuments] = useState(true);

  const projects = namespace.projects || [];
  const documents = namespace.documents || [];

  return (
    <SidebarGroup>
      <SidebarGroupLabel>{namespace.name || "Namespace"}</SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu>
          {/* Projects Section */}
          <Collapsible
            open={expandedProjects}
            onOpenChange={setExpandedProjects}
          >
            <SidebarMenuItem>
              <CollapsibleTrigger asChild>
                <SidebarMenuButton>
                  <Folder className="size-4" />
                  <span>Projects</span>
                  <ChevronRight
                    className={cn(
                      "ml-auto size-4 transition-transform",
                      expandedProjects && "rotate-90"
                    )}
                  />
                </SidebarMenuButton>
              </CollapsibleTrigger>
              <CollapsibleContent>
                <SidebarMenuSub>
                  {projects.length > 0 ? (
                    projects.map((project) => (
                      <SidebarMenuSubItem key={project.id}>
                        <SidebarMenuSubButton
                          asChild
                          isActive={
                            currentPath ===
                            `/settings/organizations/${context.organizationId}/namespaces/${context.namespaceId}`
                          }
                        >
                          <Link
                            to="/settings/organizations/$organizationId/namespaces/$namespaceId"
                            params={{
                              organizationId: context.organizationId!,
                              namespaceId: context.namespaceId!,
                            }}
                          >
                            <span>{project.name}</span>
                          </Link>
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                    ))
                  ) : (
                    <SidebarMenuSubItem>
                      <div className="text-muted-foreground px-2 py-1 text-xs">
                        No projects
                      </div>
                    </SidebarMenuSubItem>
                  )}
                </SidebarMenuSub>
              </CollapsibleContent>
            </SidebarMenuItem>
          </Collapsible>

          {/* Documents Section */}
          <Collapsible
            open={expandedDocuments}
            onOpenChange={setExpandedDocuments}
          >
            <SidebarMenuItem>
              <CollapsibleTrigger asChild>
                <SidebarMenuButton>
                  <FileText className="size-4" />
                  <span>Documents</span>
                  <ChevronRight
                    className={cn(
                      "ml-auto size-4 transition-transform",
                      expandedDocuments && "rotate-90"
                    )}
                  />
                </SidebarMenuButton>
              </CollapsibleTrigger>
              <CollapsibleContent>
                <SidebarMenuSub>
                  {documents.length > 0 ? (
                    documents.map((document) => (
                      <SidebarMenuSubItem key={document.id}>
                        <SidebarMenuSubButton
                          asChild
                          isActive={
                            currentPath ===
                            `/settings/organizations/${context.organizationId}/namespaces/${context.namespaceId}`
                          }
                        >
                          <Link
                            to="/settings/organizations/$organizationId/namespaces/$namespaceId"
                            params={{
                              organizationId: context.organizationId!,
                              namespaceId: context.namespaceId!,
                            }}
                          >
                            <span>{document.name}</span>
                          </Link>
                        </SidebarMenuSubButton>
                      </SidebarMenuSubItem>
                    ))
                  ) : (
                    <SidebarMenuSubItem>
                      <div className="text-muted-foreground px-2 py-1 text-xs">
                        No documents
                      </div>
                    </SidebarMenuSubItem>
                  )}
                </SidebarMenuSub>
              </CollapsibleContent>
            </SidebarMenuItem>
          </Collapsible>

          {/* Settings */}
          {navigationItems.map((item) => {
            const isActive = currentPath === item.href;
            return (
              <SidebarMenuItem key={item.href}>
                <SidebarMenuButton asChild isActive={isActive}>
                  <Link to={item.href}>
                    <item.icon className="size-4" />
                    <span>{item.label}</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            );
          })}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
