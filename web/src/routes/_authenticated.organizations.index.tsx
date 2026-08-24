import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Building2Icon, PlusIcon } from "lucide-react";
import { useMemo } from "react";
import { z } from "zod";

import { ContentWidth } from "@/components/layout/content-width";
import {
  CursorPaginator,
  cursorPaginatorProps,
} from "@/components/shared/cursor-paginator";
import { AppList, EntityLink } from "@/components/shared/entity-link";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { InternalLink } from "@/components/ui/internal-link";
import { ListSkeleton } from "@/components/ui/list-skeleton";
import { PageHeader } from "@/components/ui/page-header";
import { SearchInput } from "@/components/ui/search-input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { StatusIndicator } from "@/components/ui/status-indicator";
import { useCursorPageNav } from "@/hooks/use-cursor-page-nav";
import {
  ResourceType,
  usePermissions,
  withResourceType,
} from "@/hooks/use-permissions";
import { cursorPageQuery } from "@/lib/api/cursor-pages";
import { v1OrganizationsGetOptions } from "@/lib/api/query-options";
import { zOrganizationStatus } from "@/lib/api/schemas";
import { Action, can } from "@/lib/auth/permissions";
import { loadOrganizations } from "@/lib/route-data";
import { withRouteErrors } from "@/lib/route-errors";
import { pluralize } from "@/lib/utils";

const organizationsListSearchSchema = z.object({
  q: z.string().optional().catch(undefined),
  status: z
    .union([z.literal("all"), zOrganizationStatus])
    .optional()
    .catch(undefined),
});

export const Route = createFileRoute("/_authenticated/organizations/")({
  validateSearch: organizationsListSearchSchema,
  loader: ({ context }) =>
    withRouteErrors(() => loadOrganizations(context.queryClient)),
  component: OrganizationsListPage,
});

function OrganizationsListPage() {
  const search = Route.useSearch();
  const navigate = Route.useNavigate();
  const statusFilter = search.status ?? zOrganizationStatus.enum.active;
  const pageNav = useCursorPageNav({
    resetKey: `${search.q ?? ""}|${statusFilter}`,
  });
  const { data: organizationsPage, isLoading } = useQuery(
    v1OrganizationsGetOptions({
      query: cursorPageQuery(pageNav.pageToken),
    })
  );
  const organizations = organizationsPage?.items;
  const { data: systemPermissions } = usePermissions(
    withResourceType(ResourceType.Installation)
  );
  const canCreate = can(systemPermissions, Action.OrganizationCreate);

  const filteredOrganizations = useMemo(() => {
    if (!organizations) return [];
    const query = search.q?.trim().toLowerCase() ?? "";

    return organizations
      .filter((organization) => {
        if (statusFilter !== "all" && organization.status !== statusFilter) {
          return false;
        }
        if (!query) return true;
        return [organization.name, organization.email]
          .filter(Boolean)
          .join(" ")
          .toLowerCase()
          .includes(query);
      })
      .sort((a, b) => {
        if (a.status !== b.status) {
          return a.status === zOrganizationStatus.enum.active ? -1 : 1;
        }
        return a.name.localeCompare(b.name);
      });
  }, [organizations, search.q, statusFilter]);

  const hasActiveFilters = Boolean(
    search.q?.trim() || statusFilter !== zOrganizationStatus.enum.active
  );

  return (
    <ContentWidth width="overview" className="space-y-6">
      <PageHeader
        title="Organizations"
        description="Organizations you can open as operational context."
        actions={
          canCreate ? (
            <Button
              size="sm"
              render={<InternalLink to="/settings/organizations/new" />}
            >
              <PlusIcon />
              Create organization
            </Button>
          ) : undefined
        }
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
          placeholder="Search organizations..."
          className="min-w-60"
        />
        <Select
          value={statusFilter}
          onValueChange={(value) =>
            void navigate({
              search: (previous) => ({
                ...previous,
                status:
                  !value || value === zOrganizationStatus.enum.active
                    ? undefined
                    : value,
              }),
              replace: true,
            })
          }
          items={{
            all: "All statuses",
            ...Object.fromEntries(
              zOrganizationStatus.options.map((status) => [
                status,
                status.charAt(0).toUpperCase() + status.slice(1),
              ])
            ),
          }}
        >
          <SelectTrigger className="w-44">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All statuses</SelectItem>
            {zOrganizationStatus.options.map((status) => (
              <SelectItem key={status} value={status}>
                {status.charAt(0).toUpperCase() + status.slice(1)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {isLoading ? (
        <ListSkeleton />
      ) : filteredOrganizations.length > 0 ? (
        <>
          <AppList>
            {filteredOrganizations.map((organization) => (
              <EntityLink
                key={organization.id}
                type="organization"
                href={`/organizations/${organization.id}`}
                title={organization.name}
                imageUrl={organization.logo}
                subtitle={
                  <span className="flex items-center gap-2">
                    <StatusIndicator status={organization.status} />
                    <span className="truncate">
                      {organization.email ||
                        `${organization.member_count ?? 0} ${pluralize(
                          organization.member_count ?? 0,
                          "member",
                          "members"
                        )}`}
                    </span>
                  </span>
                }
              />
            ))}
          </AppList>
          <CursorPaginator
            {...cursorPaginatorProps(organizationsPage, pageNav)}
          />
        </>
      ) : (
        <EmptyState
          icon={<Building2Icon />}
          title={
            hasActiveFilters ? "No matching organizations" : "No organizations"
          }
          description={
            hasActiveFilters
              ? "Try a different search or status filter."
              : "Join or create an organization from Settings to establish workspace context."
          }
          action={
            hasActiveFilters ? (
              <Button
                variant="outline"
                onClick={() =>
                  void navigate({
                    search: { q: undefined, status: undefined },
                    replace: true,
                  })
                }
              >
                Clear filters
              </Button>
            ) : canCreate ? (
              <Button
                variant="outline"
                render={<InternalLink to="/settings/organizations/new" />}
              >
                <PlusIcon />
                Create organization
              </Button>
            ) : (
              <Button
                variant="outline"
                render={<InternalLink to="/settings/organizations" />}
              >
                Manage organizations
              </Button>
            )
          }
        />
      )}
    </ContentWidth>
  );
}
