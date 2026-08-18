import { useQueries } from "@tanstack/react-query";
import { CommandIcon, SearchIcon, SlidersHorizontalIcon } from "lucide-react";
import { useMemo } from "react";

import { ContentWidth } from "@/components/layout/content-width";
import { ResponsiveInspectorShell } from "@/components/layout/responsive-inspector-shell";
import { openQuickCreate } from "@/components/quick-create/open";
import { MockDataAlert } from "@/components/shared/app-feedback";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { Input } from "@/components/ui/input";
import { InternalLink } from "@/components/ui/internal-link";
import { ListSkeleton } from "@/components/ui/list-skeleton";
import { PageHeader } from "@/components/ui/page-header";
import { Section } from "@/components/ui/section";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { WorkInspector } from "@/components/work/work-inspector";
import { useAccessibleNamespaces } from "@/lib/api/accessible-namespaces";
import { collectedListQuery, cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1NamespacesProjectsGetOptions } from "@/lib/api/query-options";
import { v1NamespacesProjectsGet } from "@/lib/api/sdk";
import { getWorkItem, searchGlobalFixtures } from "@/lib/mock-data";
import type { Scope, SearchResultKind } from "@/lib/mock-data";
import { recentEntityLinkType } from "@/lib/recent-entity";
import { useUiSelector } from "@/lib/ui-store";
import type { SearchRouteSearch } from "@/lib/work-route-search";

export function SearchPage({
  search,
  onSearchChange,
}: {
  search: SearchRouteSearch;
  onSearchChange: (patch: Partial<SearchRouteSearch>) => void;
}) {
  const { data: accessibleWorkspace, isLoading: isWorkspaceLoading } =
    useAccessibleNamespaces();
  const namespaces = accessibleWorkspace?.namespaces ?? [];
  const recent = useUiSelector((state) => state.recentEntities);
  const scope: Scope =
    search.scope === "global"
      ? { type: "global" }
      : { type: "namespace", namespaceId: search.scope };
  const fixtureKinds =
    search.type === "all"
      ? undefined
      : ([search.type] as readonly SearchResultKind[]);
  const fixtureResults = searchGlobalFixtures({
    text: search.q,
    scope,
    kinds: fixtureKinds,
    limit: 100,
  });
  const normalizedQuery = search.q.trim().toLowerCase();
  const projectQueries = useQueries({
    queries: namespaces.map((namespace) => {
      const listOptions = v1NamespacesProjectsGetOptions({
        path: { id: namespace.id },
      });
      return {
        ...collectedListQuery(listOptions, async (pageToken, signal) => {
          const { data } = await v1NamespacesProjectsGet({
            path: { id: namespace.id },
            query: cursorPageQuery(pageToken),
            signal,
            throwOnError: true,
          });
          return data;
        }),
        enabled: search.type === "all" && Boolean(normalizedQuery),
      };
    }),
  });
  const realNamespaceResults =
    search.type === "all" && normalizedQuery
      ? namespaces.filter((namespace) =>
          [namespace.name, namespace.description, namespace.organizationName]
            .filter(Boolean)
            .join(" ")
            .toLowerCase()
            .includes(normalizedQuery)
        )
      : [];
  const realProjectResults =
    search.type === "all" && normalizedQuery
      ? namespaces.flatMap((namespace, index) =>
          (projectQueries[index]?.data?.items ?? [])
            .filter((project) =>
              [project.name, project.key, project.description]
                .filter(Boolean)
                .join(" ")
                .toLowerCase()
                .includes(normalizedQuery)
            )
            .map((project) => ({ ...project, namespace }))
        )
      : [];
  const pageSize = 20;
  const pageStart = (search.page - 1) * pageSize;
  const pagedFixtures = fixtureResults.slice(pageStart, pageStart + pageSize);
  const selectedItem = search.selected?.startsWith("work:")
    ? getWorkItem(search.selected.slice(5))
    : undefined;
  const total =
    fixtureResults.length +
    realNamespaceResults.length +
    realProjectResults.length;
  const isSearchLoading =
    Boolean(search.q) &&
    search.type === "all" &&
    (isWorkspaceLoading || projectQueries.some((query) => query.isLoading));

  const groupedFixtures = useMemo(
    () =>
      Map.groupBy(pagedFixtures, (result) =>
        result.kind === "work-item"
          ? "Work"
          : result.kind === "document"
            ? "Documents"
            : result.kind === "person"
              ? "People"
              : "Saved views"
      ),
    [pagedFixtures]
  );

  return (
    <div className="h-[calc(100svh-var(--app-header-height))] min-h-0 overflow-hidden">
      <ResponsiveInspectorShell
        className="h-full min-h-0 min-w-0"
        inspector={
          selectedItem ? <WorkInspector item={selectedItem} /> : undefined
        }
        inspectorTitle={selectedItem?.key}
        inspectorDescription={selectedItem?.title}
        open={Boolean(selectedItem)}
        onOpenChange={(open) => {
          if (!open) onSearchChange({ selected: undefined });
        }}
      >
        <ContentWidth width="overview" className="max-w-6xl space-y-6">
          <PageHeader
            title="Search"
            description="Find entities and actions across contexts you can access."
          />

          <div className="bg-background sticky top-0 z-20 space-y-3 border-y py-3">
            <div className="relative">
              <SearchIcon className="text-muted-foreground absolute top-3 left-3 size-5" />
              <Input
                autoFocus
                value={search.q}
                onChange={(event) =>
                  onSearchChange({ q: event.target.value, page: 1 })
                }
                placeholder="Search work, projects, documents, people..."
                className="h-11 pl-10 text-base"
              />
            </div>
            <div className="flex flex-wrap gap-2">
              <Select
                value={search.scope}
                onValueChange={(value) =>
                  onSearchChange({ scope: value ?? "global", page: 1 })
                }
                items={{
                  global: "Everywhere",
                  ...Object.fromEntries(
                    namespaces.map((namespace) => [
                      namespace.id,
                      namespace.name,
                    ])
                  ),
                }}
              >
                <SelectTrigger size="sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="global">Everywhere</SelectItem>
                  {namespaces.map((namespace) => (
                    <SelectItem key={namespace.id} value={namespace.id}>
                      {namespace.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select
                value={search.type}
                onValueChange={(value) =>
                  onSearchChange({
                    type: (value ?? "all") as SearchRouteSearch["type"],
                    page: 1,
                  })
                }
                items={{
                  all: "All types",
                  "work-item": "Work",
                  document: "Documents",
                  "saved-view": "Saved views",
                  person: "People",
                }}
              >
                <SelectTrigger size="sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All types</SelectItem>
                  <SelectItem value="work-item">Work</SelectItem>
                  <SelectItem value="document">Documents</SelectItem>
                  <SelectItem value="saved-view">Saved views</SelectItem>
                  <SelectItem value="person">People</SelectItem>
                </SelectContent>
              </Select>
              <Button variant="outline" size="sm" disabled>
                <SlidersHorizontalIcon />
                More filters
              </Button>
              {search.q && (
                <span className="text-muted-foreground ml-auto self-center text-xs">
                  {total} {total === 1 ? "result" : "results"}
                </span>
              )}
            </div>
          </div>

          {!search.q ? (
            <div className="grid gap-8 md:grid-cols-2">
              <Section title="Recent entities">
                {recent.length > 0 ? (
                  <AppList>
                    {recent.slice(0, 8).map((entity) => (
                      <EntityLink
                        key={`${entity.type}:${entity.id}`}
                        type={recentEntityLinkType(entity.type)}
                        href={entity.href}
                        title={entity.label}
                      />
                    ))}
                  </AppList>
                ) : (
                  <EmptyState
                    compact
                    icon={<SearchIcon />}
                    title="No recent entities"
                    description="Recently opened entities will appear here."
                  />
                )}
              </Section>
              <Section title="Useful commands">
                <AppList>
                  <button
                    type="button"
                    onClick={() => openQuickCreate()}
                    className="hover:bg-muted/50 flex w-full items-center gap-3 px-3 py-3 text-left text-sm"
                  >
                    <CommandIcon className="text-muted-foreground size-4" />
                    Quick create
                    <kbd className="text-muted-foreground ml-auto text-xs">
                      C
                    </kbd>
                  </button>
                  <InternalLink
                    to="/my-work"
                    className="hover:bg-muted/50 flex items-center gap-3 px-3 py-3 text-sm"
                  >
                    <CommandIcon className="text-muted-foreground size-4" />
                    Open My Work
                  </InternalLink>
                  <InternalLink
                    to="/namespaces"
                    className="hover:bg-muted/50 flex items-center gap-3 px-3 py-3 text-sm"
                  >
                    <CommandIcon className="text-muted-foreground size-4" />
                    Browse namespaces
                  </InternalLink>
                </AppList>
              </Section>
            </div>
          ) : isSearchLoading && total === 0 ? (
            <ListSkeleton />
          ) : total === 0 ? (
            <EmptyState
              icon={<SearchIcon />}
              title="No results"
              description="Try different terms, a broader scope, or another entity type."
              action={
                <Button
                  variant="outline"
                  onClick={() =>
                    onSearchChange({
                      q: "",
                      scope: "global",
                      type: "all",
                      page: 1,
                    })
                  }
                >
                  Clear search
                </Button>
              }
            />
          ) : (
            <div className="space-y-8">
              {(realNamespaceResults.length > 0 ||
                realProjectResults.length > 0) && (
                <Section title="Namespaces & projects">
                  <AppList>
                    {realNamespaceResults.map((namespace) => (
                      <EntityLink
                        key={namespace.id}
                        type="namespace"
                        href={`/namespaces/${namespace.id}`}
                        title={namespace.name}
                        subtitle={namespace.organizationName}
                      />
                    ))}
                    {realProjectResults.map(({ namespace, ...project }) => (
                      <EntityLink
                        key={project.id}
                        type="project"
                        href={`/namespaces/${namespace.id}/projects/${project.id}`}
                        title={project.name}
                        subtitle={`${namespace.name} · ${project.status}`}
                      />
                    ))}
                  </AppList>
                </Section>
              )}

              {pagedFixtures.length > 0 && (
                <>
                  <MockDataAlert title="Illustrative entity search results">
                    Some results are illustrative examples. Namespace and
                    project matches above come from your workspace.
                  </MockDataAlert>
                  {[...groupedFixtures.entries()].map(([group, results]) => (
                    <Section key={group} title={group}>
                      <AppList>
                        {results.map((result) => (
                          <div
                            key={`${result.kind}:${result.id}`}
                            onClick={() => {
                              if (result.kind === "work-item") {
                                onSearchChange({
                                  selected: `work:${result.id}`,
                                });
                              }
                            }}
                          >
                            <EntityLink
                              href={
                                result.kind === "saved-view"
                                  ? `/my-work?view=${result.id}`
                                  : result.href
                              }
                              type={result.kind}
                              title={result.title}
                              subtitle={result.subtitle}
                            />
                          </div>
                        ))}
                      </AppList>
                    </Section>
                  ))}
                </>
              )}

              {fixtureResults.length > pageSize && (
                <div className="flex items-center justify-center gap-2">
                  <Button
                    variant="outline"
                    disabled={search.page <= 1}
                    onClick={() => onSearchChange({ page: search.page - 1 })}
                  >
                    Previous
                  </Button>
                  <span className="text-muted-foreground text-sm">
                    Page {search.page}
                  </span>
                  <Button
                    variant="outline"
                    disabled={pageStart + pageSize >= fixtureResults.length}
                    onClick={() => onSearchChange({ page: search.page + 1 })}
                  >
                    Next
                  </Button>
                </div>
              )}
            </div>
          )}
        </ContentWidth>
      </ResponsiveInspectorShell>
    </div>
  );
}
