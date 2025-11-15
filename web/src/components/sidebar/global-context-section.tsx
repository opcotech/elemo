import { useQueries, useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { ChevronRight, FileText, Folder, Star } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { Badge } from "@/components/ui/badge";
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { Skeleton } from "@/components/ui/skeleton";
import { useNavigationContext } from "@/hooks/use-navigation-context";
import {
  v1OrganizationsGetOptions,
  v1OrganizationsNamespacesGetOptions,
} from "@/lib/api";
import type { Namespace } from "@/lib/api";
import { cn } from "@/lib/utils";

const FAVORITES_STORAGE_KEY = "elemo_sidebar_favorites";

interface NamespaceWithOrganization extends Namespace {
  organizationId: string;
  organizationName: string;
}

function getFavorites(): string[] {
  if (typeof window === "undefined") return [];
  try {
    const stored = localStorage.getItem(FAVORITES_STORAGE_KEY);
    return stored ? JSON.parse(stored) : [];
  } catch {
    return [];
  }
}

function setFavorites(favorites: string[]): void {
  if (typeof window === "undefined") return;
  try {
    localStorage.setItem(FAVORITES_STORAGE_KEY, JSON.stringify(favorites));
  } catch {
    // Ignore storage errors
  }
}

function toggleFavorite(namespaceId: string): void {
  const favorites = getFavorites();
  const index = favorites.indexOf(namespaceId);
  if (index > -1) {
    favorites.splice(index, 1);
  } else {
    favorites.push(namespaceId);
  }
  setFavorites(favorites);
}

function isFavorite(namespaceId: string): boolean {
  return getFavorites().includes(namespaceId);
}

export function GlobalContextSection() {
  const navigationContext = useNavigationContext();
  const [expandedNamespace, setExpandedNamespace] = useState<
    string | undefined
  >(navigationContext.namespaceId || undefined);
  const [favoritesVersion, setFavoritesVersion] = useState(0); // Force re-render when favorites change

  // Sync expanded namespace with navigation context
  useEffect(() => {
    setExpandedNamespace(navigationContext.namespaceId || undefined);
  }, [navigationContext.namespaceId]);

  const { data: organizations, isLoading: isLoadingOrgs } = useQuery(
    v1OrganizationsGetOptions()
  );

  // Fetch namespaces for all organizations using useQueries
  const namespaceQueries = useQueries({
    queries:
      organizations && organizations.length > 0
        ? organizations.map((org) =>
            v1OrganizationsNamespacesGetOptions({
              path: { id: org.id },
            })
          )
        : [],
  });

  // Combine all namespaces with organization info
  const allNamespaces = useMemo(() => {
    if (!organizations) return [];
    const results: NamespaceWithOrganization[] = [];
    namespaceQueries.forEach((query, index) => {
      const org = organizations[index];
      if (org && query.data) {
        for (const ns of query.data) {
          results.push({
            ...ns,
            organizationId: org.id,
            organizationName: org.name,
          });
        }
      }
    });
    return results;
  }, [organizations, namespaceQueries]);

  const isLoadingNamespaces = namespaceQueries.some((q) => q.isLoading);

  // Sort namespaces: favorites first, then by name
  const sortedNamespaces = useMemo(() => {
    return [...allNamespaces].sort((a, b) => {
      const aIsFavorite = isFavorite(a.id);
      const bIsFavorite = isFavorite(b.id);
      if (aIsFavorite !== bIsFavorite) {
        return aIsFavorite ? -1 : 1;
      }
      return a.name.localeCompare(b.name);
    });
  }, [allNamespaces, favoritesVersion]);

  // Limit visible namespaces (show favorites + some recent)
  const visibleNamespaces = useMemo(() => {
    const favoriteNamespaces = sortedNamespaces.filter((ns) =>
      isFavorite(ns.id)
    );
    const otherNamespaces = sortedNamespaces.filter((ns) => !isFavorite(ns.id));
    // Show all favorites + up to 10 others
    return [...favoriteNamespaces, ...otherNamespaces.slice(0, 10)];
  }, [sortedNamespaces]);

  const handleFavoriteToggle = (e: React.MouseEvent, namespaceId: string) => {
    e.stopPropagation(); // Prevent accordion toggle
    toggleFavorite(namespaceId);
    setFavoritesVersion((v) => v + 1); // Trigger re-render
  };

  if (isLoadingOrgs || isLoadingNamespaces) {
    return (
      <SidebarGroup>
        <SidebarGroupLabel>Namespaces</SidebarGroupLabel>
        <SidebarGroupContent>
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-7 w-full" />
            ))}
          </div>
        </SidebarGroupContent>
      </SidebarGroup>
    );
  }

  if (allNamespaces.length === 0) {
    return (
      <SidebarGroup>
        <SidebarGroupLabel>Namespaces</SidebarGroupLabel>
        <SidebarGroupContent>
          <div className="text-muted-foreground px-2 py-4 text-sm">
            No namespaces available
          </div>
        </SidebarGroupContent>
      </SidebarGroup>
    );
  }

  return (
    <SidebarGroup>
      <SidebarGroupLabel>Namespaces</SidebarGroupLabel>
      <SidebarGroupContent>
        <Accordion
          type="single"
          collapsible
          value={expandedNamespace}
          onValueChange={(value) => {
            setExpandedNamespace(value);
          }}
          className="w-full space-y-1"
        >
          {visibleNamespaces.map((namespace) => {
            const isActive =
              navigationContext.type === "namespace" &&
              navigationContext.namespaceId === namespace.id;
            const namespaceIsFavorite = isFavorite(namespace.id);
            const projectCount = namespace.projects?.length || 0;

            return (
              <AccordionItem
                key={namespace.id}
                value={namespace.id}
                className="border-none"
              >
                <SidebarMenuItem>
                  <AccordionTrigger
                    className={cn(
                      "group/item hover:bg-primary/5 hover:text-primary data-[state=open]:bg-primary/5 data-[state=open]:text-primary px-2 py-1.5",
                      isActive && "bg-primary/10 text-primary"
                    )}
                  >
                    <div className="flex w-full flex-1 items-center gap-2">
                      <Folder className="size-4 shrink-0" />
                      <span className="flex-1 truncate text-left font-normal">
                        {namespace.name}
                      </span>
                      {projectCount > 0 && (
                        <Badge variant="secondary" className="text-xs">
                          {projectCount}
                        </Badge>
                      )}
                      <button
                        type="button"
                        onClick={(e) => handleFavoriteToggle(e, namespace.id)}
                        className={cn(
                          "focus:ring-primary rounded p-0.5 transition-opacity hover:opacity-80 focus:ring-2 focus:ring-offset-2 focus:outline-none",
                          namespaceIsFavorite
                            ? "text-yellow-400"
                            : "text-muted-foreground opacity-0 transition-opacity group-hover/item:opacity-100 hover:text-yellow-400"
                        )}
                        aria-label={
                          namespaceIsFavorite
                            ? "Remove from favorites"
                            : "Add to favorites"
                        }
                      >
                        <Star
                          className={cn(
                            "size-3 shrink-0",
                            namespaceIsFavorite && "fill-yellow-400"
                          )}
                        />
                      </button>
                    </div>
                  </AccordionTrigger>
                  <AccordionContent className="pt-1 pb-0">
                    <div className="ml-4 space-y-2">
                      {/* Projects Section */}
                      <div>
                        <div className="text-muted-foreground mb-1 px-2 text-xs font-medium">
                          Projects
                        </div>
                        <SidebarMenu className="space-y-1">
                          {namespace.projects &&
                          namespace.projects.length > 0 ? (
                            namespace.projects.map((project) => {
                              const isProjectActive =
                                navigationContext.type === "project" &&
                                navigationContext.projectId === project.id;
                              return (
                                <SidebarMenuItem key={project.id}>
                                  <SidebarMenuButton
                                    asChild
                                    isActive={isProjectActive}
                                    className="h-8"
                                  >
                                    <Link
                                      to="/settings/organizations/$organizationId/namespaces/$namespaceId"
                                      params={{
                                        organizationId:
                                          namespace.organizationId,
                                        namespaceId: namespace.id,
                                      }}
                                    >
                                      <ChevronRight className="size-3 shrink-0" />
                                      <span className="truncate">
                                        {project.name}
                                      </span>
                                    </Link>
                                  </SidebarMenuButton>
                                </SidebarMenuItem>
                              );
                            })
                          ) : (
                            <SidebarMenuItem>
                              <div className="text-muted-foreground px-2 py-1 text-xs">
                                No projects
                              </div>
                            </SidebarMenuItem>
                          )}
                        </SidebarMenu>
                      </div>

                      {/* Documents Section */}
                      <div>
                        <div className="text-muted-foreground mb-1 px-2 text-xs font-medium">
                          Documents
                        </div>
                        <SidebarMenu className="space-y-1">
                          {namespace.documents &&
                          namespace.documents.length > 0 ? (
                            namespace.documents.map((document) => {
                              return (
                                <SidebarMenuItem key={document.id}>
                                  <SidebarMenuButton asChild className="h-8">
                                    <Link
                                      to="/settings/organizations/$organizationId/namespaces/$namespaceId"
                                      params={{
                                        organizationId:
                                          namespace.organizationId,
                                        namespaceId: namespace.id,
                                      }}
                                    >
                                      <FileText className="size-3 shrink-0" />
                                      <span className="truncate">
                                        {document.name}
                                      </span>
                                    </Link>
                                  </SidebarMenuButton>
                                </SidebarMenuItem>
                              );
                            })
                          ) : (
                            <SidebarMenuItem>
                              <div className="text-muted-foreground px-2 py-1 text-xs">
                                No documents
                              </div>
                            </SidebarMenuItem>
                          )}
                        </SidebarMenu>
                      </div>
                    </div>
                  </AccordionContent>
                </SidebarMenuItem>
              </AccordionItem>
            );
          })}
        </Accordion>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
