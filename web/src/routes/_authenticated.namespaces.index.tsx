import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { ChevronsUpDownIcon, Layers3Icon } from "lucide-react";
import { useMemo, useState } from "react";
import { z } from "zod";

import { ContentWidth } from "@/components/layout/content-width";
import { NamespaceEntitySubtitle } from "@/components/namespaces/namespace-entity-subtitle";
import {
  CursorPaginator,
  cursorPaginatorProps,
} from "@/components/shared/cursor-paginator";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { EmptyState } from "@/components/ui/empty-state";
import { InternalLink } from "@/components/ui/internal-link";
import { ListSkeleton } from "@/components/ui/list-skeleton";
import { PageHeader } from "@/components/ui/page-header";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { SearchInput } from "@/components/ui/search-input";
import { useCursorPageNav } from "@/hooks/use-cursor-page-nav";
import {
  accessibleNamespacesOptions,
  useAccessibleNamespaces,
} from "@/lib/api/accessible-namespaces";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1NamespacesGetOptions } from "@/lib/api/query-options";
import { withRouteErrors } from "@/lib/route-errors";
import { cn } from "@/lib/utils";

const namespacesListSearchSchema = z.object({
  q: z.string().optional().catch(undefined),
  organization: z.string().optional().catch(undefined),
});

export const Route = createFileRoute("/_authenticated/namespaces/")({
  validateSearch: namespacesListSearchSchema,
  loader: ({ context }) =>
    withRouteErrors(() =>
      context.queryClient.fetchQuery(
        accessibleNamespacesOptions(context.queryClient)
      )
    ),
  component: NamespacesListPage,
});

function OrganizationFilter({
  value,
  organizations,
  onChange,
}: {
  value?: string;
  organizations: { id: string; name: string }[];
  onChange: (organizationId: string | undefined) => void;
}) {
  const [open, setOpen] = useState(false);
  const sortedOrganizations = useMemo(
    () => organizations.slice().sort((a, b) => a.name.localeCompare(b.name)),
    [organizations]
  );
  const selected = sortedOrganizations.find(
    (organization) => organization.id === value
  );
  const label = selected?.name ?? "All organizations";

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            aria-label="Filter by organization"
            className={cn(
              "border-border bg-card hover:bg-card dark:bg-input dark:hover:bg-input/80 h-9 w-56 justify-between rounded-md border font-normal shadow-none"
            )}
          />
        }
      >
        <span className="truncate">{label}</span>
        <ChevronsUpDownIcon className="text-muted-foreground size-4 shrink-0 opacity-50" />
      </PopoverTrigger>
      <PopoverContent className="w-56 p-0" align="start">
        <Command>
          <CommandInput placeholder="Search organizations..." />
          <CommandList>
            <CommandEmpty>No organization found.</CommandEmpty>
            <CommandGroup>
              <CommandItem
                value="All organizations"
                data-checked={!value || undefined}
                onSelect={() => {
                  onChange(undefined);
                  setOpen(false);
                }}
              >
                All organizations
              </CommandItem>
              {sortedOrganizations.map((organization) => (
                <CommandItem
                  key={organization.id}
                  value={organization.name}
                  data-checked={value === organization.id || undefined}
                  onSelect={() => {
                    onChange(organization.id);
                    setOpen(false);
                  }}
                >
                  {organization.name}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

function NamespacesListPage() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const { data: accessibleWorkspace } = useAccessibleNamespaces();
  const organizations = accessibleWorkspace?.organizations ?? [];
  const pageNav = useCursorPageNav({
    resetKey: `${search.q ?? ""}|${search.organization ?? ""}`,
  });
  const { data: namespacesPage, isLoading } = useQuery(
    v1NamespacesGetOptions({
      query: cursorPageQuery(pageNav.pageToken),
    })
  );
  const namespaces = useMemo(
    () =>
      (namespacesPage?.items ?? []).map((namespace) => ({
        ...namespace,
        organizationId: namespace.organization.id,
        organizationName: namespace.organization.name,
      })),
    [namespacesPage?.items]
  );

  const filteredNamespaces = useMemo(() => {
    const query = search.q?.trim().toLowerCase() ?? "";
    const organizationId = search.organization;

    return namespaces
      .filter((namespace) => {
        if (organizationId && namespace.organizationId !== organizationId) {
          return false;
        }
        if (!query) return true;
        return [
          namespace.name,
          namespace.description,
          namespace.organizationName,
        ]
          .filter(Boolean)
          .join(" ")
          .toLowerCase()
          .includes(query);
      })
      .sort((a, b) => a.name.localeCompare(b.name));
  }, [namespaces, search.organization, search.q]);

  const hasActiveFilters = Boolean(search.q?.trim() || search.organization);

  return (
    <ContentWidth width="overview" className="space-y-6">
      <PageHeader
        title="Namespaces"
        description="Namespaces you can open as operational context."
      />

      <div className="bg-background sticky top-0 z-10 flex flex-wrap items-center gap-2 py-3">
        <SearchInput
          value={search.q ?? ""}
          onChange={(value) =>
            void navigate({
              search: (previous) => ({
                ...previous,
                q: value || undefined,
              }),
              replace: true,
            })
          }
          placeholder="Search namespaces..."
          className="min-w-60"
        />
        <OrganizationFilter
          value={search.organization}
          organizations={organizations}
          onChange={(organizationId) =>
            void navigate({
              search: (previous) => ({
                ...previous,
                organization: organizationId,
              }),
              replace: true,
            })
          }
        />
      </div>

      {isLoading ? (
        <ListSkeleton />
      ) : filteredNamespaces.length > 0 ? (
        <>
          <AppList>
            {filteredNamespaces.map((namespace) => (
              <EntityLink
                key={namespace.id}
                type="namespace"
                href={`/namespaces/${namespace.id}`}
                title={namespace.name}
                subtitle={
                  <NamespaceEntitySubtitle
                    description={namespace.description}
                    organizationName={namespace.organizationName}
                    projectCount={namespace.project_count ?? 0}
                    documentCount={namespace.document_count ?? 0}
                  />
                }
              />
            ))}
          </AppList>
          <CursorPaginator {...cursorPaginatorProps(namespacesPage, pageNav)} />
        </>
      ) : (
        <EmptyState
          icon={<Layers3Icon />}
          title={hasActiveFilters ? "No matching namespaces" : "No namespaces"}
          description={
            hasActiveFilters
              ? "Try a different search or organization filter."
              : "Join or create a namespace from Settings to establish an operational context."
          }
          action={
            hasActiveFilters ? (
              <Button
                variant="outline"
                onClick={() =>
                  void navigate({
                    search: { q: undefined, organization: undefined },
                    replace: true,
                  })
                }
              >
                Clear filters
              </Button>
            ) : (
              <Button
                variant="outline"
                render={<InternalLink to="/settings/namespaces" />}
              >
                Manage namespaces
              </Button>
            )
          }
        />
      )}
    </ContentWidth>
  );
}
