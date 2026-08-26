import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { CommandIcon, SearchIcon } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";

import { ContentWidth } from "@/components/layout/content-width";
import { openQuickCreate } from "@/components/quick-create/open";
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
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useAccessibleNamespaces } from "@/lib/api/accessible-namespaces";
import {
  collectedListQuery,
  cursorPageQuery,
  flattenCursorPages,
  nextCursorPageToken,
} from "@/lib/api/cursor-pages";
import {
  v1NamespacesProjectsGetOptions,
  v1SearchGetOptions,
} from "@/lib/api/query-options";
import { namespaceRefPath } from "@/lib/api/refs";
import { v1NamespacesProjectsGet, v1SearchGet } from "@/lib/api/sdk";
import { recentEntityLinkType } from "@/lib/recent-entity";
import {
  SEARCH_DEBOUNCE_MS,
  hasActiveSearch,
  searchQueryFromRoute,
} from "@/lib/search/params";
import type { SearchRouteSearch, SearchRouteType } from "@/lib/search/params";
import {
  SEARCH_RESOURCE_TYPES,
  SEARCH_TYPE_LABELS,
  groupSearchResults,
  searchResultEntityType,
  searchResultHref,
} from "@/lib/search/result";
import { useUiSelector } from "@/lib/ui-store";

const ALL_FILTER = "all";

export function SearchPage({
  search,
  onSearchChange,
}: {
  search: SearchRouteSearch;
  onSearchChange: (patch: Partial<SearchRouteSearch>) => void;
}) {
  const { data: accessibleWorkspace, isLoading: isWorkspaceLoading } =
    useAccessibleNamespaces();
  const organizations = accessibleWorkspace?.organizations ?? [];
  const namespaces = accessibleWorkspace?.namespaces ?? [];
  const recent = useUiSelector((state) => state.recentEntities);
  const [queryInput, setQueryInput] = useState(search.q);
  const [previousSearchQ, setPreviousSearchQ] = useState(search.q);
  if (search.q !== previousSearchQ) {
    setPreviousSearchQ(search.q);
    setQueryInput(search.q);
  }
  const debouncedQ = useDebouncedValue(queryInput, SEARCH_DEBOUNCE_MS);
  const qForSearch = queryInput === search.q ? search.q : debouncedQ;
  const searchForQuery = { ...search, q: qForSearch, page_token: undefined };
  const filterKey = [
    qForSearch,
    search.type,
    search.organization_id ?? "",
    search.namespace_id ?? "",
    search.project_id ?? "",
  ].join("|");
  const scrollRootRef = useRef<HTMLDivElement | null>(null);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (debouncedQ === search.q || debouncedQ !== queryInput) {
      return;
    }
    onSearchChange({ q: debouncedQ, page_token: undefined });
  }, [debouncedQ, onSearchChange, queryInput, search.q]);

  const namespacesForOrg = useMemo(
    () =>
      search.organization_id
        ? namespaces.filter(
            (namespace) => namespace.organizationId === search.organization_id
          )
        : namespaces,
    [namespaces, search.organization_id]
  );
  const selectedNamespace = namespaces.find(
    (namespace) => namespace.id === search.namespace_id
  );
  const projectListOptions = v1NamespacesProjectsGetOptions({
    path: namespaceRefPath(
      selectedNamespace?.organizationId ?? "",
      search.namespace_id ?? ""
    ),
  });
  const { data: projectsPage, isLoading: isProjectsLoading } = useQuery({
    ...collectedListQuery(projectListOptions, async (pageToken, signal) => {
      const { data } = await v1NamespacesProjectsGet({
        path: namespaceRefPath(
          selectedNamespace?.organizationId ?? "",
          search.namespace_id ?? ""
        ),
        query: cursorPageQuery(pageToken),
        signal,
        throwOnError: true,
      });
      return data;
    }),
    enabled: Boolean(search.namespace_id),
  });
  const projects = projectsPage?.items ?? [];
  const active = hasActiveSearch(searchForQuery);
  const searchQuery = searchQueryFromRoute(searchForQuery);
  const searchListOptions = v1SearchGetOptions({
    query: searchQuery,
  });
  const {
    data: searchPages,
    error: searchError,
    isPending: isSearchPending,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery({
    staleTime: searchListOptions.staleTime,
    gcTime: searchListOptions.gcTime,
    queryKey: [...searchListOptions.queryKey, "infinite"],
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ pageParam, signal }) => {
      const { data } = await v1SearchGet({
        query: {
          ...searchQuery,
          ...(pageParam ? { page_token: pageParam } : {}),
        },
        signal,
        throwOnError: true,
      });
      return data;
    },
    getNextPageParam: (lastPage) => nextCursorPageToken(lastPage),
    enabled: active,
  });
  const items = active ? flattenCursorPages(searchPages?.pages ?? []) : [];
  const groupedResults = useMemo(() => groupSearchResults(items), [items]);
  const isError = Boolean(searchError);

  useEffect(() => {
    if (!active || !hasNextPage || isFetchingNextPage || isSearchPending) {
      return;
    }
    const root = scrollRootRef.current;
    const sentinel = sentinelRef.current;
    if (!sentinel) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        if (!entries.some((entry) => entry.isIntersecting)) {
          return;
        }
        void fetchNextPage();
      },
      { root, rootMargin: "0px 0px 80px 0px" }
    );
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [
    active,
    fetchNextPage,
    filterKey,
    hasNextPage,
    isFetchingNextPage,
    isSearchPending,
    items.length,
  ]);

  function applyFilters(patch: Partial<SearchRouteSearch>) {
    onSearchChange({ ...patch, page_token: undefined });
  }

  function handleOrganizationChange(value: string | null) {
    const organizationId = value && value !== ALL_FILTER ? value : undefined;
    const namespaceStillValid =
      selectedNamespace &&
      (!organizationId || selectedNamespace.organizationId === organizationId);
    applyFilters({
      organization_id: organizationId,
      namespace_id: namespaceStillValid ? search.namespace_id : undefined,
      project_id: namespaceStillValid ? search.project_id : undefined,
    });
  }

  function handleNamespaceChange(value: string | null) {
    const namespaceId = value && value !== ALL_FILTER ? value : undefined;
    const namespace = namespaces.find((item) => item.id === namespaceId);
    applyFilters({
      namespace_id: namespaceId,
      organization_id: namespace?.organizationId ?? search.organization_id,
      project_id: undefined,
    });
  }

  function handleProjectChange(value: string | null) {
    applyFilters({
      project_id: value && value !== ALL_FILTER ? value : undefined,
    });
  }

  const typeItems = {
    [ALL_FILTER]: SEARCH_TYPE_LABELS.all,
    ...Object.fromEntries(
      SEARCH_RESOURCE_TYPES.map((type) => [type, SEARCH_TYPE_LABELS[type]])
    ),
  };
  const organizationItems = {
    [ALL_FILTER]: "All organizations",
    ...Object.fromEntries(
      organizations.map((organization) => [organization.id, organization.name])
    ),
  };
  const namespaceItems = {
    [ALL_FILTER]: "All namespaces",
    ...Object.fromEntries(
      namespacesForOrg.map((namespace) => [namespace.id, namespace.name])
    ),
  };
  const projectItems = {
    [ALL_FILTER]: "All projects",
    ...Object.fromEntries(
      projects.map((project) => [project.id, project.name])
    ),
  };

  return (
    <div
      ref={scrollRootRef}
      className="h-[calc(100svh-var(--app-header-height))] min-h-0 overflow-y-auto"
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
              type="search"
              value={queryInput}
              onChange={(event) => setQueryInput(event.target.value)}
              placeholder="Search organizations, work, documents..."
              className="h-11 pl-10 text-base"
              aria-label="Search"
            />
          </div>
          <div className="flex flex-wrap gap-2">
            <Select
              value={search.type}
              onValueChange={(value) =>
                applyFilters({
                  type: (value ?? ALL_FILTER) as SearchRouteType,
                })
              }
              items={typeItems}
            >
              <SelectTrigger size="sm" aria-label="Filter by type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ALL_FILTER}>
                  {SEARCH_TYPE_LABELS.all}
                </SelectItem>
                {SEARCH_RESOURCE_TYPES.map((type) => (
                  <SelectItem key={type} value={type}>
                    {SEARCH_TYPE_LABELS[type]}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={search.organization_id ?? ALL_FILTER}
              onValueChange={handleOrganizationChange}
              items={organizationItems}
            >
              <SelectTrigger size="sm" aria-label="Filter by organization">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ALL_FILTER}>All organizations</SelectItem>
                {organizations.map((organization) => (
                  <SelectItem key={organization.id} value={organization.id}>
                    {organization.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={search.namespace_id ?? ALL_FILTER}
              onValueChange={handleNamespaceChange}
              items={namespaceItems}
            >
              <SelectTrigger size="sm" aria-label="Filter by namespace">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ALL_FILTER}>All namespaces</SelectItem>
                {namespacesForOrg.map((namespace) => (
                  <SelectItem key={namespace.id} value={namespace.id}>
                    {namespace.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={search.project_id ?? ALL_FILTER}
              onValueChange={handleProjectChange}
              items={projectItems}
              disabled={!search.namespace_id || isProjectsLoading}
            >
              <SelectTrigger size="sm" aria-label="Filter by project">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ALL_FILTER}>All projects</SelectItem>
                {projects.map((project) => (
                  <SelectItem key={project.id} value={project.id}>
                    {project.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        {!active ? (
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
                  <kbd className="text-muted-foreground ml-auto text-xs">C</kbd>
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
        ) : (isSearchPending || isWorkspaceLoading) && items.length === 0 ? (
          <ListSkeleton />
        ) : isError ? (
          <EmptyState
            icon={<SearchIcon />}
            title="Search failed"
            description="Try again in a moment, or adjust your filters."
            action={
              <Button
                variant="outline"
                onClick={() =>
                  applyFilters({
                    q: "",
                    type: "all",
                    organization_id: undefined,
                    namespace_id: undefined,
                    project_id: undefined,
                  })
                }
              >
                Clear search
              </Button>
            }
          />
        ) : items.length === 0 ? (
          <EmptyState
            icon={<SearchIcon />}
            title="No results"
            description="Try different terms, a broader scope, or another entity type."
            action={
              <Button
                variant="outline"
                onClick={() =>
                  applyFilters({
                    q: "",
                    type: "all",
                    organization_id: undefined,
                    namespace_id: undefined,
                    project_id: undefined,
                  })
                }
              >
                Clear search
              </Button>
            }
          />
        ) : (
          <div className="space-y-8">
            {groupedResults.map((group) => (
              <Section key={group.type} title={SEARCH_TYPE_LABELS[group.type]}>
                <AppList>
                  {group.items.map((item) => {
                    const href = searchResultHref(item);
                    if (!href) {
                      return (
                        <div
                          key={`${item.type}:${item.id}`}
                          className="text-muted-foreground px-3 py-2.5 text-sm"
                        >
                          {item.title}
                        </div>
                      );
                    }
                    return (
                      <EntityLink
                        key={`${item.type}:${item.id}`}
                        type={searchResultEntityType(item)}
                        href={href}
                        title={item.title}
                        subtitle={item.subtitle}
                      />
                    );
                  })}
                </AppList>
              </Section>
            ))}

            {hasNextPage ? (
              <div
                ref={sentinelRef}
                className="h-px w-full shrink-0"
                aria-hidden
              />
            ) : null}
          </div>
        )}
      </ContentWidth>
    </div>
  );
}
